import AsyncStorage from "@react-native-async-storage/async-storage";
import * as AuthSession from "expo-auth-session";
import * as Crypto from "expo-crypto";
import * as SecureStore from "expo-secure-store";
import { api } from "../api/client";
import { getAccessToken } from "../api/auth";
import { freeWindow, type BusyPeriod } from "./calendarAvailability";
import { currentInstallationId, installationId } from "./device";

const clientId = process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID;
const clientPrefix = clientId?.match(/^([^.]+)\.apps\.googleusercontent\.com$/)?.[1];
const tokenKey = "calendar.google.token.v1";
const revocationKey = "calendar.google.revoke.v1";
const selectionKey = "calendar.google.selection.v1";
const discovery = {
  authorizationEndpoint: "https://accounts.google.com/o/oauth2/v2/auth",
  tokenEndpoint: "https://oauth2.googleapis.com/token",
  revocationEndpoint: "https://oauth2.googleapis.com/revoke",
};
const calendarAPI = "https://www.googleapis.com/calendar/v3";
type Connection = { ownerId: string; accessToken: string; refreshToken: string; expiresAt: number };
type Selection = { ownerId: string; ids: string[]; pending?: boolean; disconnecting?: boolean; requestKey?: string };
type CalendarChoice = { id: string; title: string };
export type GoogleCalendarStatus = { available: boolean; connected: boolean; calendars: CalendarChoice[]; selectedIds: string[]; syncPending: boolean; loadError?: string };
let pending: Promise<void> = Promise.resolve();
let revocationWork: Promise<void> = Promise.resolve();
function serialized<T>(work: () => Promise<T>): Promise<T> {
  const result = pending.then(work);
  pending = result.then(() => {}, () => {});
  return result;
}

function serializeRevocations<T>(work: () => Promise<T>): Promise<T> {
  const result = revocationWork.then(work);
  revocationWork = result.then(() => {}, () => {});
  return result;
}

async function queuedRevocations(): Promise<string[]> {
  const raw = await SecureStore.getItemAsync(revocationKey);
  if (!raw) return [];
  const tokens: unknown = JSON.parse(raw);
  if (!Array.isArray(tokens) || !tokens.every(token => typeof token === "string" && token.length > 0)) throw new Error("Google Calendar revocation state is invalid.");
  return tokens;
}

function rememberRevocation(token: string) { return serializeRevocations(async () => {
  const tokens = await queuedRevocations();
  if (!tokens.includes(token)) await SecureStore.setItemAsync(revocationKey, JSON.stringify([...tokens, token]));
}); }

function retryRevocations() { return serializeRevocations(async () => {
  const tokens = await queuedRevocations();
  for (const token of tokens) {
    await revokeGoogleToken(token);
    const remaining = (await queuedRevocations()).filter(item => item !== token);
    if (remaining.length) await SecureStore.setItemAsync(revocationKey, JSON.stringify(remaining));
    else await SecureStore.deleteItemAsync(revocationKey);
  }
}); }

function forgetRevocation(token: string) { return serializeRevocations(async () => {
  const remaining = (await queuedRevocations()).filter(item => item !== token);
  if (remaining.length) await SecureStore.setItemAsync(revocationKey, JSON.stringify(remaining));
  else await SecureStore.deleteItemAsync(revocationKey);
}); }

export function retryGoogleCalendarRevocations() { return serialized(retryRevocations); }

async function ownerId() {
  return (await api<{ learner: { userId: string } }>("/v1/bootstrap")).learner.userId;
}

async function connection(owner: string): Promise<Connection | null> {
  const raw = await SecureStore.getItemAsync(tokenKey);
  if (!raw) return null;
  let saved: Connection | null = null;
  try {
    const value = JSON.parse(raw) as Connection;
    if (typeof value.ownerId === "string" && typeof value.accessToken === "string" && typeof value.refreshToken === "string" && typeof value.expiresAt === "number") saved = value;
  } catch { /* Invalid local state is not a connection. */ }
  if (saved?.ownerId !== owner) {
    if (saved?.refreshToken) await rememberRevocation(saved.refreshToken);
    await SecureStore.deleteItemAsync(tokenKey);
    await AsyncStorage.removeItem(selectionKey);
    return null;
  }
  return saved;
}

async function selection(owner: string): Promise<Selection> {
  const raw = await AsyncStorage.getItem(selectionKey);
  if (raw) {
    try {
      const saved = JSON.parse(raw) as Selection;
      if (saved.ownerId === owner && Array.isArray(saved.ids) && saved.ids.every(id => typeof id === "string")) return saved;
    } catch { /* Invalid local state cannot grant access. */ }
    await AsyncStorage.removeItem(selectionKey);
  }
  return { ownerId: owner, ids: [] };
}

async function accessToken(owner: string, force = false) {
  const saved = await connection(owner);
  if (!saved) throw new Error("Connect Google Calendar again.");
  if (!force && saved.expiresAt > Date.now() / 1000 + 60) return saved.accessToken;
  if (!clientId) throw new Error("Google Calendar is not configured.");
  const token = await AuthSession.refreshAsync({ clientId, refreshToken: saved.refreshToken }, discovery);
  if (!token.accessToken || !token.expiresIn) throw new Error("Google Calendar access expired. Connect it again.");
  const next = { ...saved, accessToken: token.accessToken, refreshToken: token.refreshToken || saved.refreshToken, expiresAt: token.issuedAt + token.expiresIn };
  await SecureStore.setItemAsync(tokenKey, JSON.stringify(next));
  return next.accessToken;
}

async function googleJSON<T>(owner: string, path: string, init?: RequestInit): Promise<T> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 7_000);
  try {
    const send = async (token: string) => fetch(`${calendarAPI}${path}`, { ...init, headers: { authorization: `Bearer ${token}`, ...(init?.body ? { "content-type": "application/json" } : {}) }, signal: controller.signal });
    let response = await send(await accessToken(owner));
    if (response.status === 401) response = await send(await accessToken(owner, true));
    if (!response.ok) throw new Error(`Google Calendar could not be read (${response.status}).`);
    return await response.json() as T;
  } finally { clearTimeout(timeout); }
}

async function listCalendars(owner: string): Promise<CalendarChoice[]> {
  const calendars: CalendarChoice[] = [];
  const seen = new Set<string>();
  let page = "";
  do {
    const query = new URLSearchParams({ maxResults: "250", fields: "items(id,summary),nextPageToken" });
    if (page) query.set("pageToken", page);
    const result = await googleJSON<{ items?: { id?: unknown; summary?: unknown }[]; nextPageToken?: unknown }>(owner, `/users/me/calendarList?${query}`);
    if (result.items !== undefined && !Array.isArray(result.items)) throw new Error("Google returned an invalid calendar list.");
    for (const item of result.items || []) {
      if (typeof item.id !== "string" || !item.id || typeof item.summary !== "string") throw new Error("Google returned an invalid calendar.");
      if (!seen.has(item.id)) calendars.push({ id: item.id, title: item.summary });
      seen.add(item.id);
    }
    if (result.nextPageToken !== undefined && typeof result.nextPageToken !== "string") throw new Error("Google returned an invalid calendar page.");
    if (result.nextPageToken === page && page) throw new Error("Google returned a repeated calendar page.");
    page = result.nextPageToken || "";
  } while (page);
  return calendars;
}

async function busyPeriods(owner: string, ids: string[], started: Date): Promise<BusyPeriod[]> {
  const busy: BusyPeriod[] = [];
  for (let offset = 0; offset < ids.length; offset += 50) {
    const batch = ids.slice(offset, offset + 50);
    const result = await googleJSON<{ calendars?: Record<string, { busy?: BusyPeriod[]; errors?: unknown[] }> }>(owner, "/freeBusy", {
      method: "POST",
      body: JSON.stringify({ timeMin: started.toISOString(), timeMax: new Date(started.getTime() + 95 * 60_000).toISOString(), items: batch.map(id => ({ id })) }),
    });
    for (const id of batch) {
      const entry = result.calendars?.[id];
      if (!entry || !Array.isArray(entry.busy) || (entry.errors !== undefined && (!Array.isArray(entry.errors) || entry.errors.length > 0))) throw new Error("Google could not confirm availability for every selected calendar.");
      for (const period of entry.busy) {
        if (typeof period.start !== "string" || typeof period.end !== "string" || !Number.isFinite(Date.parse(period.start)) || !Number.isFinite(Date.parse(period.end)) || Date.parse(period.end) <= Date.parse(period.start)) {
          throw new Error("Google returned invalid availability.");
        }
        busy.push(period);
      }
    }
  }
  return busy;
}

async function sendWindows(installation: string, windows: ReturnType<typeof freeWindow>, expectedToken: string) {
  await api("/v1/calendar/google/windows", { method: "POST", expectedToken, body: JSON.stringify({ installationId: installation, windows }) });
}

async function withdrawAvailability(expectedToken: string) {
  const installation = await currentInstallationId();
  if (installation) await sendWindows(installation, [], expectedToken).catch(() => {});
}

async function publishAvailability(owner: string, ids: string[], installation: string, expectedToken: string) {
  let windows: ReturnType<typeof freeWindow>;
  try {
    const started = new Date();
    const busy = await busyPeriods(owner, ids, started);
    const now = new Date();
    if (now.getTime() - started.getTime() > 3 * 60_000) throw new Error("Google Calendar scan took too long. Reopen the app to retry.");
    windows = freeWindow(busy, now);
  } catch (error) {
    await sendWindows(installation, [], expectedToken).catch(() => {});
    throw error;
  }
  await sendWindows(installation, windows, expectedToken);
}

async function readGoogleCalendarStatus(): Promise<GoogleCalendarStatus> {
  if (!clientPrefix) return { available: false, connected: false, calendars: [], selectedIds: [], syncPending: false };
  const expectedToken = await getAccessToken();
  const owner = await ownerId();
  const saved = await connection(owner);
  if (!saved) return { available: true, connected: false, calendars: [], selectedIds: [], syncPending: false };
  const choice = await selection(owner);
  if (choice.disconnecting) return { available: true, connected: true, calendars: [], selectedIds: [], syncPending: true };
  let calendars: CalendarChoice[];
  try { calendars = await listCalendars(owner); }
  catch {
    if (await getAccessToken() !== expectedToken) throw new Error("Account changed during Calendar read.");
    if (choice.ids.length) await withdrawAvailability(expectedToken);
    return { available: true, connected: true, calendars: [], selectedIds: choice.ids, syncPending: Boolean(choice.pending), loadError: "Couldn’t read Google calendars. Retry or disconnect." };
  }
  if (await getAccessToken() !== expectedToken) throw new Error("Account changed during Calendar read.");
  const available = new Set(calendars.map(calendar => calendar.id));
  return { available: true, connected: true, calendars, selectedIds: choice.ids.filter(id => available.has(id)), syncPending: Boolean(choice.pending) };
}

export function googleCalendarStatus() { return serialized(readGoogleCalendarStatus); }

export function connectGoogleCalendar() { return serialized(async () => {
  if (!clientId || !clientPrefix) throw new Error("Google Calendar is not configured for this iPhone build.");
  await retryRevocations();
  const owner = await ownerId();
  if (await connection(owner)) return readGoogleCalendarStatus();
  const redirectUri = `com.googleusercontent.apps.${clientPrefix}:/oauthredirect`;
  const request = new AuthSession.AuthRequest({ clientId, redirectUri, responseType: AuthSession.ResponseType.Code, usePKCE: true,
    scopes: ["https://www.googleapis.com/auth/calendar.calendarlist.readonly", "https://www.googleapis.com/auth/calendar.freebusy"],
    extraParams: { access_type: "offline", prompt: "consent" },
  });
  const result = await request.promptAsync(discovery);
  if (result.type !== "success" || !result.params.code || !request.codeVerifier) throw new Error("Google Calendar connection was canceled.");
  const token = await AuthSession.exchangeCodeAsync({ clientId, code: result.params.code, redirectUri, extraParams: { code_verifier: request.codeVerifier } }, discovery);
  if (!token.accessToken || !token.expiresIn) throw new Error("Google did not grant renewable Calendar access. Try connecting again.");
  if (!token.refreshToken) {
    await rememberRevocation(token.accessToken);
    await retryRevocations().catch(() => {});
    throw new Error("Google did not grant renewable Calendar access. Try connecting again.");
  }
  await rememberRevocation(token.refreshToken);
  if (await ownerId() !== owner) {
    await retryRevocations().catch(() => {});
    throw new Error("Account changed during Google connection.");
  }
  try {
    await SecureStore.setItemAsync(tokenKey, JSON.stringify({ ownerId: owner, accessToken: token.accessToken, refreshToken: token.refreshToken, expiresAt: token.issuedAt + token.expiresIn }));
    await AsyncStorage.setItem(selectionKey, JSON.stringify({ ownerId: owner, ids: [] }));
    await forgetRevocation(token.refreshToken);
  } catch (error) {
    await SecureStore.deleteItemAsync(tokenKey).catch(() => {});
    await AsyncStorage.removeItem(selectionKey).catch(() => {});
    await retryRevocations().catch(() => {});
    throw error;
  }
  return readGoogleCalendarStatus();
}); }

async function sync(owner: string, choice: Selection, expectedToken: string) {
  const installation = await installationId();
  if (choice.pending) {
    if (!choice.requestKey) throw new Error("Google Calendar choice is incomplete. Choose calendars again.");
    await api(`/v1/devices/${encodeURIComponent(installation)}`, { method: "PUT", expectedToken, body: JSON.stringify({ googleCalendarPermission: "enabled" }) });
    await api("/v1/consents", { method: "POST", expectedToken, idempotencyKey: choice.requestKey, body: JSON.stringify({ purpose: "google_calendar_context", policyVersion: "1", granted: true }) });
  }
  await publishAvailability(owner, choice.ids, installation, expectedToken);
  await AsyncStorage.setItem(selectionKey, JSON.stringify({ ownerId: owner, ids: choice.ids }));
}

async function clearAvailability(expectedToken: string) {
  const id = await currentInstallationId();
  if (id) await api(`/v1/calendar/google/windows?installationId=${encodeURIComponent(id)}`, { method: "DELETE", expectedToken });
}

async function revokeGoogleToken(token: string) {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 7_000);
  try {
    const response = await fetch(discovery.revocationEndpoint, { method: "POST", headers: { "content-type": "application/x-www-form-urlencoded" }, body: new URLSearchParams({ token }).toString(), signal: controller.signal });
    if (!response.ok && response.status !== 400) throw new Error("Google access could not be revoked. Orbit will retry when you reopen the app.");
  } finally { clearTimeout(timeout); }
}

export function selectGoogleCalendars(ids: string[]) { return serialized(async () => {
  const expectedToken = await getAccessToken();
  const status = await readGoogleCalendarStatus();
  if (!status.connected) throw new Error("Connect Google Calendar first.");
  if ((await selection(await ownerId())).disconnecting) throw new Error("Google Calendar is disconnecting. Reopen the app to retry.");
  const allowed = new Set(status.calendars.map(calendar => calendar.id));
  if (new Set(ids).size !== ids.length || ids.some(id => !allowed.has(id))) throw new Error("Choose calendars from the connected Google account.");
  const owner = await ownerId();
  if (await getAccessToken() !== expectedToken) throw new Error("Account changed during Calendar selection.");
  const choice: Selection = { ownerId: owner, ids, pending: true, requestKey: Crypto.randomUUID() };
  await AsyncStorage.setItem(selectionKey, JSON.stringify(choice));
  if (ids.length) await sync(owner, choice, expectedToken);
  else { await clearAvailability(expectedToken); await AsyncStorage.setItem(selectionKey, JSON.stringify({ ownerId: owner, ids: [] })); }
  return readGoogleCalendarStatus();
}); }

async function disconnectNow(owner: string, expectedToken: string) {
  const saved = await connection(owner);
  if (!saved) { await AsyncStorage.removeItem(selectionKey); return; }
  await AsyncStorage.setItem(selectionKey, JSON.stringify({ ownerId: owner, ids: [], disconnecting: true }));
  await clearAvailability(expectedToken);
  await rememberRevocation(saved.refreshToken);
  await SecureStore.deleteItemAsync(tokenKey);
  await AsyncStorage.removeItem(selectionKey);
  await retryRevocations().catch(() => { /* Keep the token queued for the next foreground attempt. */ });
}

export function disconnectGoogleCalendar() { return serialized(async () => {
  if (!clientPrefix) return;
  const expectedToken = await getAccessToken();
  await disconnectNow(await ownerId(), expectedToken);
}); }

export function clearGoogleCalendarOnSignOut() { return serialized(async () => {
  const raw = await SecureStore.getItemAsync(tokenKey);
  if (raw) {
    try {
      const saved = JSON.parse(raw) as Connection;
      if (typeof saved.refreshToken === "string" && saved.refreshToken) await rememberRevocation(saved.refreshToken);
    } catch (error) {
      if (!(error instanceof SyntaxError)) throw error;
    }
  }
  await SecureStore.deleteItemAsync(tokenKey);
  await AsyncStorage.removeItem(selectionKey);
  await retryRevocations().catch(() => { /* Keep the revocation token for the next foreground attempt. */ });
}); }

export function reconcileGoogleCalendar() { return serialized(async () => {
  if (!clientPrefix) return;
  const expectedToken = await getAccessToken();
  const owner = await ownerId();
  const saved = await connection(owner);
  if (!saved) return;
  const choice = await selection(owner);
  if (choice.disconnecting) { await disconnectNow(owner, expectedToken); return; }
  let calendars: CalendarChoice[];
  try { calendars = await listCalendars(owner); }
  catch (error) {
    await withdrawAvailability(expectedToken);
    throw error;
  }
  const available = new Set(calendars.map(calendar => calendar.id));
  const ids = choice.ids.filter(id => available.has(id));
  if (!ids.length) {
    await clearAvailability(expectedToken);
    await AsyncStorage.setItem(selectionKey, JSON.stringify({ ownerId: owner, ids: [] }));
  } else await sync(owner, { ...choice, ids }, expectedToken);
}); }
