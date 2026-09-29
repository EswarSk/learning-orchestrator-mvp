import AsyncStorage from "@react-native-async-storage/async-storage";
import * as Crypto from "expo-crypto";
import * as Calendar from "expo-calendar";
import { api } from "../api/client";
import { getAccessToken } from "../api/auth";
import { installationId } from "./device";
import { freeWindow } from "./calendarAvailability";

const selectionKey = "calendar.ios.selection.v1";
let pending: Promise<void> = Promise.resolve();
function serialized<T>(work: () => Promise<T>): Promise<T> {
  const result = pending.then(work);
  pending = result.then(() => {}, () => {});
  return result;
}
type Selection = { ownerId: string; ids: string[]; pending?: boolean; requestKey?: string };
export type IOSCalendarStatus = {
  ownerId?: string;
  permissionGranted: boolean;
  calendars: { id: string; title: string; account: string }[];
  selectedIds: string[];
  syncPending: boolean;
};

async function ownerId() {
  const account = await api<{ learner: { userId: string } }>("/v1/bootstrap");
  return account.learner.userId;
}

export async function iosCalendarStatus(): Promise<IOSCalendarStatus> {
  const permission = await Calendar.getCalendarPermissions(false);
  if (!permission.granted) {
    return { permissionGranted: false, calendars: [], selectedIds: [], syncPending: false };
  }
  const [owner, calendars, raw] = await Promise.all([
    ownerId(), Calendar.getCalendars(Calendar.EntityTypes.EVENT), AsyncStorage.getItem(selectionKey),
  ]);
  let saved: Selection | null = null;
  try {
    const parsed = raw ? JSON.parse(raw) as Selection : null;
    if (parsed && typeof parsed.ownerId === "string" && Array.isArray(parsed.ids) && parsed.ids.every(id => typeof id === "string")) saved = parsed;
  } catch { /* An interrupted or old local write is not a calendar grant. */ }
  const available = new Set(calendars.map(calendar => calendar.id));
  const selectedIds = saved?.ownerId === owner ? saved.ids.filter(id => available.has(id)) : [];
  if (raw && saved?.ownerId !== owner) await AsyncStorage.removeItem(selectionKey);
  return {
    ownerId: owner,
    permissionGranted: true,
    calendars: calendars.map(calendar => ({ id: calendar.id, title: calendar.title, account: calendar.source?.name || "On this iPhone" })),
    selectedIds,
    syncPending: saved?.ownerId === owner && saved.pending === true,
  };
}

export async function requestIOSCalendarAccess() {
  const permission = await Calendar.requestCalendarPermissions(false);
  if (!permission.granted) throw new Error("Allow full Calendar access in iOS Settings to choose calendars. Orbit will not edit events.");
  return iosCalendarStatus();
}

export function selectIOSCalendars(ids: string[]) { return serialized(async () => {
  const expectedToken = await getAccessToken();
  const status = await iosCalendarStatus();
  if (!status.permissionGranted) throw new Error("Calendar access is off in iOS Settings.");
  const available = new Set(status.calendars.map(calendar => calendar.id));
  if (new Set(ids).size !== ids.length || ids.some(id => !available.has(id))) throw new Error("Choose calendars from this iPhone.");
  if (await getAccessToken() !== expectedToken) throw new Error("Account changed during Calendar selection.");
  await AsyncStorage.setItem(selectionKey, JSON.stringify({ ownerId: status.ownerId, ids, pending: true, requestKey: Crypto.randomUUID() }));
  if (ids.length) await syncSelection(ids, status.ownerId!, expectedToken);
  else await clearIOSCalendarsNow(expectedToken);
  return { ...status, selectedIds: ids, syncPending: false };
}); }

async function syncSelection(ids: string[], ownerId: string, expectedToken: string) {
  const id = await installationId();
  const raw = await AsyncStorage.getItem(selectionKey);
  const selection = raw ? JSON.parse(raw) as Selection : null;
  if (selection?.pending) {
    if (!selection.requestKey) throw new Error("Calendar choice is incomplete. Select calendars again.");
    await api(`/v1/devices/${encodeURIComponent(id)}`, { method: "PUT", expectedToken, body: JSON.stringify({ calendarPermission: "enabled" }) });
    await api("/v1/consents", { method: "POST", expectedToken, idempotencyKey: selection.requestKey, body: JSON.stringify({ purpose: "ios_calendar_context", policyVersion: "1", granted: true }) });
  }
  await publishAvailability(ids, id, expectedToken);
  await AsyncStorage.setItem(selectionKey, JSON.stringify({ ownerId, ids }));
}

async function clearIOSCalendarsNow(expectedToken: string) {
  const raw = await AsyncStorage.getItem(selectionKey);
  if (raw) {
    let saved: Selection | null = null;
    try { saved = JSON.parse(raw) as Selection; } catch { /* Invalid local state cannot grant access. */ }
    if (saved?.ownerId !== await ownerId()) { await AsyncStorage.removeItem(selectionKey); return; }
    const id = await installationId();
    await api(`/v1/calendar/ios/windows?installationId=${encodeURIComponent(id)}`, { method: "DELETE", expectedToken });
  }
  await AsyncStorage.removeItem(selectionKey);
}
export function reconcileIOSCalendarPermission() { return serialized(async () => {
  const expectedToken = await getAccessToken();
  if (!(await Calendar.getCalendarPermissions(false)).granted) { await clearIOSCalendarsNow(expectedToken); return; }
  const status = await iosCalendarStatus();
  if (status.selectedIds.length) {
    await syncSelection(status.selectedIds, status.ownerId!, expectedToken);
  } else await clearIOSCalendarsNow(expectedToken);
}); }

export function availabilityWindows(events: Pick<Calendar.ExpoCalendarEvent, "startDate" | "endDate" | "status" | "availability">[], now: Date) {
  return freeWindow(events.filter(event => event.status !== Calendar.EventStatus.CANCELED && event.availability !== Calendar.Availability.FREE)
    .map(event => ({ start: event.startDate, end: event.endDate })), now);
}

async function publishAvailability(ids: string[], id: string, expectedToken: string) {
  const send = (windows: ReturnType<typeof availabilityWindows>) =>
    api("/v1/calendar/ios/windows", { method: "POST", expectedToken, body: JSON.stringify({ installationId: id, windows }) });
  let windows: ReturnType<typeof availabilityWindows>;
  try {
    const scanStarted = new Date();
    const events = await Calendar.listEvents(ids, scanStarted, new Date(scanStarted.getTime() + 95 * 60_000));
    const now = new Date();
    if (now.getTime() - scanStarted.getTime() > 3 * 60_000) throw new Error("Calendar scan took too long. Reopen the app to retry.");
    windows = availabilityWindows(events, now);
  } catch (error) {
    await send([]).catch(() => {});
    throw error;
  }
  await send(windows);
}
