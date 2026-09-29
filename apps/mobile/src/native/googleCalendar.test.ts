import { beforeEach, expect, test, vi } from "vitest";

const state = vi.hoisted(() => { process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID = "test.apps.googleusercontent.com"; return ({
  storage: new Map<string, string>(), secure: new Map<string, string>(), owner: "account-a",
  calls: [] as { path: string; method?: string; body?: string; idempotencyKey?: string }[],
  requests: [] as { url: string; init?: RequestInit }[], ids: ["personal", "work"] as string[],
  busy: [] as { start: string; end: string }[], freeBusyError: false, listError: false, listHang: false, switchOwnerDuringList: false, revokeError: false, revokeHang: false, failDelete: false, failSelectionSave: false, missingRefreshToken: false, switchOwnerOnExchange: false, switchOwnerDuringBusy: false, refreshBarrier: null as Promise<void> | null, refreshes: 0, uuids: 0,
}); });

vi.mock("@react-native-async-storage/async-storage", () => ({ default: {
  getItem: async (key: string) => state.storage.get(key) ?? null,
  setItem: async (key: string, value: string) => { if (state.failSelectionSave && key === "calendar.google.selection.v1") throw new Error("Storage unavailable"); state.storage.set(key, value); },
  removeItem: async (key: string) => { state.storage.delete(key); },
} }));
vi.mock("expo-secure-store", () => ({
  getItemAsync: async (key: string) => state.secure.get(key) ?? null,
  setItemAsync: async (key: string, value: string) => { state.secure.set(key, value); },
  deleteItemAsync: async (key: string) => { state.secure.delete(key); },
}));
vi.mock("expo-crypto", () => ({ randomUUID: () => `request-${++state.uuids}` }));
vi.mock("../api/auth", () => ({ getAccessToken: async () => `${state.owner}-token` }));
vi.mock("expo-auth-session", () => ({
  ResponseType: { Code: "code" },
  AuthRequest: class { codeVerifier = "verifier"; async promptAsync() { return { type: "success", params: { code: "code" } }; } },
  exchangeCodeAsync: async () => { if (state.switchOwnerOnExchange) state.owner = "account-b"; return { accessToken: "first-token", refreshToken: state.missingRefreshToken ? undefined : "refresh-token", issuedAt: Date.now() / 1000, expiresIn: 3600 }; },
  refreshAsync: async () => { state.refreshes++; await state.refreshBarrier; return { accessToken: "new-token", refreshToken: "new-refresh", issuedAt: Date.now() / 1000, expiresIn: 3600 }; },
}));
vi.mock("./device", () => ({ installationId: async () => "device-a", currentInstallationId: async () => "device-a" }));
vi.mock("../api/client", () => ({ api: async (path: string, init?: { method?: string; body?: string; idempotencyKey?: string; expectedToken?: string }) => {
  if (init?.expectedToken && init.expectedToken !== `${state.owner}-token`) throw new Error("Account changed during request");
  state.calls.push({ path, ...init });
  if (path === "/v1/bootstrap") return { learner: { userId: state.owner } };
  if (state.failDelete && path.startsWith("/v1/calendar/google/windows") && init?.method === "DELETE") { state.failDelete = false; throw new Error("Network unavailable"); }
  return {};
} }));

import { clearGoogleCalendarOnSignOut, connectGoogleCalendar, disconnectGoogleCalendar, googleCalendarStatus, reconcileGoogleCalendar, retryGoogleCalendarRevocations, selectGoogleCalendars } from "./googleCalendar";

beforeEach(() => {
  vi.useRealTimers(); state.storage.clear(); state.secure.clear(); state.owner = "account-a"; state.calls = []; state.requests = [];
  state.ids = ["personal", "work"]; state.busy = []; state.freeBusyError = false; state.listError = false; state.listHang = false; state.switchOwnerDuringList = false; state.revokeError = false; state.revokeHang = false; state.failDelete = false; state.failSelectionSave = false; state.missingRefreshToken = false; state.switchOwnerOnExchange = false; state.switchOwnerDuringBusy = false; state.refreshBarrier = null; state.refreshes = 0; state.uuids = 0;
  vi.stubGlobal("fetch", async (url: string, init?: RequestInit) => {
    state.requests.push({ url, init });
    if (url.includes("calendarList")) {
      if (state.listHang) return new Promise((_, reject) => init?.signal?.addEventListener("abort", () => reject(new Error("aborted"))));
      if (state.switchOwnerDuringList) state.owner = "account-b";
      return state.listError ? { ok: false, status: 403 } : { ok: true, status: 200, json: async () => ({ items: state.ids.map(id => ({ id, summary: id })) }) };
    }
    if (url.includes("freeBusy")) {
      const ids = (JSON.parse(String(init?.body)) as { items: { id: string }[] }).items.map(item => item.id);
      if (state.switchOwnerDuringBusy) state.owner = "account-b";
      return { ok: true, status: 200, json: async () => ({ calendars: Object.fromEntries(ids.map(id => [id, state.freeBusyError ? { errors: [{ reason: "notFound" }] } : { busy: state.busy }])) }) };
    }
    if (url.includes("revoke")) {
      if (state.revokeHang) return new Promise((_, reject) => init?.signal?.addEventListener("abort", () => reject(new Error("aborted"))));
      return state.revokeError ? { ok: false, status: 503 } : { ok: true, status: 200 };
    }
    throw new Error(`Unexpected Google request: ${url}`);
  });
});

test("selected Google calendars send only derived availability and stay separate from iPhone", async () => {
  await connectGoogleCalendar();
  await expect(selectGoogleCalendars(["other-account"])).rejects.toThrow("connected Google account");
  state.busy = [{ start: new Date(Date.now() + 40 * 60_000).toISOString(), end: new Date(Date.now() + 60 * 60_000).toISOString() }];
  await selectGoogleCalendars(["personal"]);
  const snapshot = state.calls.find(call => call.path === "/v1/calendar/google/windows" && call.method === "POST");
  expect(snapshot).toBeTruthy();
  expect(JSON.parse(snapshot!.body!).windows).toHaveLength(1);
  expect(snapshot!.body).not.toContain("personal");
  expect(snapshot!.body).not.toContain("busy");
  expect(state.calls.some(call => call.path.includes("/calendar/ios/"))).toBe(false);
  expect((await googleCalendarStatus()).selectedIds).toEqual(["personal"]);
});

test("a failed Google calendar scan withdraws availability and retries on foreground", async () => {
  await connectGoogleCalendar();
  await selectGoogleCalendars(["personal"]);
  state.freeBusyError = true;
  await expect(reconcileGoogleCalendar()).rejects.toThrow("every selected calendar");
  const snapshots = state.calls.filter(call => call.path === "/v1/calendar/google/windows" && call.method === "POST");
  expect(JSON.parse(snapshots.at(-1)!.body!).windows).toEqual([]);
  state.freeBusyError = false;
  await reconcileGoogleCalendar();
  expect(JSON.parse(state.calls.filter(call => call.path === "/v1/calendar/google/windows" && call.method === "POST").at(-1)!.body!).windows).toHaveLength(1);
});

test("a lost Google calendar-list grant withdraws the prior window", async () => {
  await connectGoogleCalendar();
  await selectGoogleCalendars(["personal"]);
  state.listError = true;
  expect(await googleCalendarStatus()).toMatchObject({ available: true, connected: true, selectedIds: ["personal"], calendars: [], loadError: "Couldn’t read Google calendars. Retry or disconnect." });
  const withdrawnOnStatus = state.calls.filter(call => call.path === "/v1/calendar/google/windows" && call.method === "POST");
  expect(JSON.parse(withdrawnOnStatus.at(-1)!.body!).windows).toEqual([]);
  await expect(reconcileGoogleCalendar()).rejects.toThrow("could not be read");
  const snapshots = state.calls.filter(call => call.path === "/v1/calendar/google/windows" && call.method === "POST");
  expect(JSON.parse(snapshots.at(-1)!.body!).windows).toEqual([]);
});

test("an unresponsive Google read times out and withdraws old availability", async () => {
  await connectGoogleCalendar();
  await selectGoogleCalendars(["personal"]);
  state.listHang = true;
  vi.useFakeTimers();
  const status = googleCalendarStatus();
  await vi.advanceTimersByTimeAsync(7_000);
  expect(await status).toMatchObject({ connected: true, loadError: "Couldn’t read Google calendars. Retry or disconnect." });
  const snapshots = state.calls.filter(call => call.path === "/v1/calendar/google/windows" && call.method === "POST");
  expect(JSON.parse(snapshots.at(-1)!.body!).windows).toEqual([]);
});

test("disconnect retries server cancellation before Google token revocation", async () => {
  await connectGoogleCalendar();
  await selectGoogleCalendars(["personal"]);
  state.failDelete = true;
  await expect(disconnectGoogleCalendar()).rejects.toThrow("Network unavailable");
  expect((await googleCalendarStatus()).syncPending).toBe(true);
  expect(state.secure.size).toBe(1);
  await reconcileGoogleCalendar();
  expect((await googleCalendarStatus()).connected).toBe(false);
  expect(state.secure.size).toBe(0);
  expect(state.requests.some(request => request.url.includes("revoke"))).toBe(true);
  await connectGoogleCalendar();
  await selectGoogleCalendars(["personal"]);
  expect((await googleCalendarStatus()).selectedIds).toEqual(["personal"]);
});

test("a failed Google revocation allows disconnect but blocks reconnect until retried", async () => {
  await connectGoogleCalendar();
  await selectGoogleCalendars(["personal"]);
  state.revokeError = true;
	await disconnectGoogleCalendar();
	expect((await googleCalendarStatus()).connected).toBe(false);
  expect(state.secure.size).toBe(1);
  expect(state.secure.has("calendar.google.token.v1")).toBe(false);
  expect(state.secure.has("calendar.google.revoke.v1")).toBe(true);
  expect(state.storage.size).toBe(0);
  await expect(connectGoogleCalendar()).rejects.toThrow("Google access could not be revoked");
  expect(state.secure.has("calendar.google.token.v1")).toBe(false);
  expect(state.secure.has("calendar.google.revoke.v1")).toBe(true);
  state.revokeError = false;
  await retryGoogleCalendarRevocations();
  await connectGoogleCalendar();
  await selectGoogleCalendars(["personal"]);
  expect((await googleCalendarStatus()).selectedIds).toEqual(["personal"]);
  expect(state.secure.has("calendar.google.revoke.v1")).toBe(false);
});

test("an account change during Google authorization retains the unused grant for revocation", async () => {
  state.switchOwnerOnExchange = true;
  state.revokeError = true;
  await expect(connectGoogleCalendar()).rejects.toThrow("Account changed during Google connection");
  expect(state.secure.has("calendar.google.token.v1")).toBe(false);
  expect(state.secure.get("calendar.google.revoke.v1")).toContain("refresh-token");
  state.revokeError = false;
  await retryGoogleCalendarRevocations();
  expect(state.secure.size).toBe(0);
});

test("failed local save revokes the new Google grant instead of keeping a partial connection", async () => {
  state.failSelectionSave = true;
  await expect(connectGoogleCalendar()).rejects.toThrow("Storage unavailable");
  expect(state.secure.size).toBe(0);
  expect(state.storage.size).toBe(0);
  expect(state.requests.some(request => request.url.includes("revoke"))).toBe(true);
});

test("an unusable Google grant is revoked when no refresh token is returned", async () => {
  state.missingRefreshToken = true;
  await expect(connectGoogleCalendar()).rejects.toThrow("renewable Calendar access");
  expect(state.secure.size).toBe(0);
  expect(state.requests.find(request => request.url.includes("revoke"))?.init?.body).toContain("first-token");
});

test("sign-out stops waiting for an unresponsive Google revocation", async () => {
  await connectGoogleCalendar();
  state.revokeHang = true;
  vi.useFakeTimers();
  const signOut = clearGoogleCalendarOnSignOut();
  await vi.advanceTimersByTimeAsync(7_000);
  await signOut;
  expect(state.secure.has("calendar.google.token.v1")).toBe(false);
  expect(state.secure.has("calendar.google.revoke.v1")).toBe(true);
});

test("sign-out waits for a status refresh before clearing Google credentials", async () => {
  await connectGoogleCalendar();
  const key = "calendar.google.token.v1";
  state.secure.set(key, JSON.stringify({ ...JSON.parse(state.secure.get(key)!), expiresAt: 0 }));
  let release!: () => void;
  state.refreshBarrier = new Promise(resolve => { release = resolve; });
  const status = googleCalendarStatus();
  await vi.waitFor(() => expect(state.refreshes).toBe(1));
  const signOut = clearGoogleCalendarOnSignOut();
  await new Promise(resolve => setTimeout(resolve, 0));
  expect(state.secure.has(key)).toBe(true);
  release();
  await Promise.all([status, signOut]);
  expect(state.secure.has(key)).toBe(false);
  expect(state.secure.has("calendar.google.revoke.v1")).toBe(false);
});

test("Google tokens refresh and are erased on account change or sign-out", async () => {
  vi.useFakeTimers(); vi.setSystemTime(new Date("2026-09-27T12:00:00Z"));
  await connectGoogleCalendar();
  vi.setSystemTime(new Date("2026-09-27T13:10:00Z"));
  await selectGoogleCalendars(["personal"]);
  expect(state.refreshes).toBe(1);
  expect(state.requests.find(request => request.url.includes("freeBusy"))?.init?.headers).toMatchObject({ authorization: "Bearer new-token" });
  state.owner = "account-b";
  expect((await googleCalendarStatus()).connected).toBe(false);
  expect(state.secure.has("calendar.google.token.v1")).toBe(false);
  expect(state.secure.has("calendar.google.revoke.v1")).toBe(true);
  await retryGoogleCalendarRevocations();
  expect(state.secure.size).toBe(0);
  state.owner = "account-a";
  await connectGoogleCalendar();
  await clearGoogleCalendarOnSignOut();
  expect(state.secure.size).toBe(0);
  expect(state.storage.size).toBe(0);
});

test("a Google scan cannot upload the previous account's availability after a switch", async () => {
  await connectGoogleCalendar();
  state.switchOwnerDuringBusy = true;
  await expect(selectGoogleCalendars(["personal"])).rejects.toThrow("Account changed");
  expect(state.calls.some(call => call.path === "/v1/calendar/google/windows")).toBe(false);
  expect((await googleCalendarStatus()).selectedIds).toEqual([]);
});

test("a Google calendar list is not returned after an account switch", async () => {
  await connectGoogleCalendar();
  state.switchOwnerDuringList = true;
  await expect(googleCalendarStatus()).rejects.toThrow("Account changed");
});
