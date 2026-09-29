import { beforeEach, expect, test, vi } from "vitest";

const state = vi.hoisted(() => ({
  storage: new Map<string, string>(), owner: "account-a", granted: true, requestedWriteOnly: null as boolean | null,
  calls: [] as { path: string; body?: string; method?: string; idempotencyKey?: string }[], events: [] as { startDate: Date; endDate: Date; status: string; availability: string; title: string }[],
  calendarIds: ["personal", "work"] as string[],
  scannedIds: [] as string[][], scanDelayMs: 0, scanError: false, switchOwnerDuringScan: false, failNext: null as "POST" | "DELETE" | null, uuidCounter: 0,
}));
vi.mock("expo-crypto", () => ({ randomUUID: () => `request-${++state.uuidCounter}` }));
vi.mock("./device", () => ({ installationId: async () => "device-a" }));
vi.mock("../api/auth", () => ({ getAccessToken: async () => `${state.owner}-token` }));
vi.mock("@react-native-async-storage/async-storage", () => ({ default: {
  getItem: async (key: string) => state.storage.get(key) ?? null,
  setItem: async (key: string, value: string) => { state.storage.set(key, value); },
  removeItem: async (key: string) => { state.storage.delete(key); },
} }));
vi.mock("expo-calendar", () => ({
  EntityTypes: { EVENT: "event" },
  EventStatus: { CANCELED: "canceled", CONFIRMED: "confirmed" }, Availability: { FREE: "free", BUSY: "busy" },
  getCalendarPermissions: async () => ({ granted: state.granted }),
  requestCalendarPermissions: async (writeOnly: boolean) => { state.requestedWriteOnly = writeOnly; return { granted: state.granted }; },
  listEvents: async (ids: string[]) => { state.scannedIds.push(ids); if (state.scanError) throw new Error("Calendar unavailable"); if (state.switchOwnerDuringScan) state.owner = "account-b"; if (state.scanDelayMs) vi.setSystemTime(Date.now() + state.scanDelayMs); return state.events; },
  getCalendars: async () => [
    { id: "personal", title: "Personal", source: { name: "iCloud" } },
    { id: "work", title: "Work", source: { name: "Google" } },
  ].filter(calendar => state.calendarIds.includes(calendar.id)),
}));
vi.mock("../api/client", () => ({ api: async (path: string, init?: { body?: string; method?: string; idempotencyKey?: string; expectedToken?: string }) => {
  if (init?.expectedToken && init.expectedToken !== `${state.owner}-token`) throw new Error("Account changed during request");
  state.calls.push({ path, ...init });
  if (path.startsWith("/v1/calendar/ios/windows") && init?.method === state.failNext) { state.failNext = null; throw new Error("Network unavailable"); }
  return path === "/v1/bootstrap" ? { learner: { userId: state.owner } } : {};
} }));

import * as Calendar from "expo-calendar";
import { availabilityWindows, iosCalendarStatus, reconcileIOSCalendarPermission, requestIOSCalendarAccess, selectIOSCalendars } from "./iosCalendar";

beforeEach(() => { vi.useRealTimers(); state.storage.clear(); state.calls = []; state.events = []; state.calendarIds = ["personal", "work"]; state.scannedIds = []; state.scanDelayMs = 0; state.scanError = false; state.switchOwnerDuringScan = false; state.failNext = null; state.uuidCounter = 0; state.owner = "account-a"; state.granted = true; state.requestedWriteOnly = null; });

test("calendar choices remain account-bound and stop when iOS permission is revoked", async () => {
  expect((await requestIOSCalendarAccess()).calendars).toHaveLength(2);
  expect(state.requestedWriteOnly).toBe(false);
  await expect(selectIOSCalendars(["unknown"])).rejects.toThrow("Choose calendars");
  await selectIOSCalendars(["personal", "work"]);
  expect((await iosCalendarStatus()).selectedIds).toEqual(["personal", "work"]);
  state.owner = "account-b";
  expect((await iosCalendarStatus()).selectedIds).toEqual([]);
  expect(state.storage.size).toBe(0);
  state.owner = "account-a";
  await selectIOSCalendars(["personal"]);
  state.granted = false;
  await reconcileIOSCalendarPermission();
  expect((await iosCalendarStatus()).selectedIds).toEqual([]);
  expect(state.storage.size).toBe(0);
  expect(state.calls.some(call => call.method === "DELETE" && call.path.startsWith("/v1/calendar/ios/windows"))).toBe(true);
});

test("calendar sync sends only free windows and reconciles after editing selection", async () => {
  const now = new Date();
  state.events = [
    { startDate: new Date(now.getTime() + 45 * 60_000), endDate: new Date(now.getTime() + 75 * 60_000), status: "confirmed", availability: "busy", title: "Private meeting" },
    { startDate: new Date(now.getTime() + 95 * 60_000), endDate: new Date(now.getTime() + 125 * 60_000), status: "confirmed", availability: "busy", title: "Private appointment" },
  ];
  await selectIOSCalendars(["personal"]);
  const sent = state.calls.find(call => call.path === "/v1/calendar/ios/windows" && call.method === "POST");
  expect(sent).toBeTruthy();
  expect(sent?.body).not.toContain("Private meeting");
  expect(JSON.parse(sent!.body!).windows).toHaveLength(1);
  await selectIOSCalendars([]);
  expect(state.calls.some(call => call.method === "DELETE" && call.path.startsWith("/v1/calendar/ios/windows"))).toBe(true);
  expect((await iosCalendarStatus()).selectedIds).toEqual([]);
  await selectIOSCalendars(["personal"]);
  expect((await iosCalendarStatus()).selectedIds).toEqual(["personal"]);
  expect(state.calls.filter(call => call.path === "/v1/calendar/ios/windows" && call.method === "POST")).toHaveLength(2);
});

test("overlapping, canceled, and free events do not create false gaps", () => {
  const now = new Date("2026-09-26T12:00:00Z");
  const event = (start: number, end: number, status = Calendar.EventStatus.CONFIRMED, availability = Calendar.Availability.BUSY) => ({
    startDate: new Date(now.getTime() + start * 60_000), endDate: new Date(now.getTime() + end * 60_000), status, availability,
  });
  expect(availabilityWindows([event(10, 40), event(20, 60), event(80, 100)], now)).toEqual([]);
  expect(availabilityWindows([event(0, 90, Calendar.EventStatus.CANCELED), event(0, 90, Calendar.EventStatus.CONFIRMED, Calendar.Availability.FREE)], now)).toEqual([
    { start: event(0, 2).endDate.toISOString(), end: event(0, 92).endDate.toISOString() },
  ]);
});

test("opening during a free gap creates a fresh near-term window", () => {
  const now = new Date("2026-09-26T12:00:00Z");
  const minutes = (value: number) => new Date(now.getTime() + value * 60_000);
  expect(availabilityWindows([], now)[0]).toEqual({ start: minutes(2).toISOString(), end: minutes(92).toISOString() });
  const events = [
    { startDate: minutes(-30), endDate: minutes(-10), status: Calendar.EventStatus.CONFIRMED, availability: Calendar.Availability.BUSY },
    { startDate: minutes(35), endDate: minutes(65), status: Calendar.EventStatus.CONFIRMED, availability: Calendar.Availability.BUSY },
  ];
  expect(availabilityWindows(events, now)[0]).toEqual({ start: minutes(2).toISOString(), end: minutes(35).toISOString() });
  expect(availabilityWindows(events, now)).toHaveLength(1);
});

test("a slow calendar scan uses a fresh window and retries if the scan is stale", async () => {
  vi.useFakeTimers();
  vi.setSystemTime(new Date("2026-09-26T12:00:00Z"));
  state.scanDelayMs = 60_000;
  await selectIOSCalendars(["personal"]);
  const first = state.calls.find(call => call.path === "/v1/calendar/ios/windows" && call.method === "POST");
  expect(JSON.parse(first!.body!).windows[0].start).toBe("2026-09-26T12:03:00.000Z");

  state.scanDelayMs = 4 * 60_000;
  await expect(selectIOSCalendars(["personal"])).rejects.toThrow("Calendar scan took too long");
  expect((await iosCalendarStatus()).syncPending).toBe(true);
  state.scanDelayMs = 0;
  await reconcileIOSCalendarPermission();
  expect((await iosCalendarStatus()).syncPending).toBe(false);
});

test("a failed device scan withdraws the previous availability instead of offering a false gap", async () => {
  await selectIOSCalendars(["personal"]);
  state.scanError = true;
  await expect(reconcileIOSCalendarPermission()).rejects.toThrow("Calendar unavailable");
  const snapshots = state.calls.filter(call => call.path === "/v1/calendar/ios/windows" && call.method === "POST");
  expect(JSON.parse(snapshots.at(-1)!.body!).windows).toEqual([]);
  expect((await iosCalendarStatus()).selectedIds).toEqual(["personal"]);
});

test("a calendar removed in iOS disconnects its last availability", async () => {
  await selectIOSCalendars(["personal"]);
  state.calendarIds = ["work"];
  await reconcileIOSCalendarPermission();
  expect(state.storage.size).toBe(0);
  expect(state.calls.some(call => call.method === "DELETE" && call.path.startsWith("/v1/calendar/ios/windows"))).toBe(true);
});

test("failed edits and disconnects retry the new choice, not the old calendars", async () => {
  await selectIOSCalendars(["personal", "work"]);
  state.failNext = "POST";
  await expect(selectIOSCalendars(["personal"])).rejects.toThrow("Network unavailable");
  expect((await iosCalendarStatus()).selectedIds).toEqual(["personal"]);
  expect((await iosCalendarStatus()).syncPending).toBe(true);
  await reconcileIOSCalendarPermission();
  expect(state.scannedIds.at(-1)).toEqual(["personal"]);
  expect((await iosCalendarStatus()).syncPending).toBe(false);
  expect(state.calls.filter(call => call.path === "/v1/consents").map(call => call.idempotencyKey)).toEqual(["request-1", "request-2", "request-2"]);
  state.failNext = "DELETE";
  await expect(selectIOSCalendars([])).rejects.toThrow("Network unavailable");
  expect((await iosCalendarStatus()).selectedIds).toEqual([]);
  expect((await iosCalendarStatus()).syncPending).toBe(true);
  expect(state.storage.size).toBe(1);
  await reconcileIOSCalendarPermission();
  expect(state.storage.size).toBe(0);
});

test("an iPhone scan cannot upload the previous account's availability after a switch", async () => {
  state.switchOwnerDuringScan = true;
  await expect(selectIOSCalendars(["personal"])).rejects.toThrow("Account changed");
  expect(state.calls.some(call => call.path === "/v1/calendar/ios/windows")).toBe(false);
  expect((await iosCalendarStatus()).selectedIds).toEqual([]);
});
