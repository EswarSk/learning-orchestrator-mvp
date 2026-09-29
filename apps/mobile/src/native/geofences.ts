import * as Crypto from "expo-crypto";
import * as Location from "expo-location";
import * as TaskManager from "expo-task-manager";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { ApiError, api } from "../api/client";
import { currentInstallationId, installationId } from "./device";

export const GEOFENCE_TASK = "pausa-geofence-v1";
const regionKey = "location.region.v1";
const ownerKey = "location.owner.v1";
const disconnectPrefix = "location.disconnect.";
const eventPrefix = "location.event.";
type Region = { id: string; version: number; trackId: string; latitude: number; longitude: number; radiusMeters: number };
type QueuedEvent = { schemaVersion: 1; installationId: string; clientEventId: string; regionId: string; regionVersion: number; transition: "enter" | "exit"; observedAt: string };
let queueWork: Promise<void> = Promise.resolve();

function serialized<T>(work: () => Promise<T>): Promise<T> {
  const result = queueWork.then(work);
  queueWork = result.then(() => {}, () => {});
  return result;
}

async function flush() {
  const pending = (await AsyncStorage.getAllKeys()).filter(key => key.startsWith(disconnectPrefix)).sort();
  for (const key of pending) {
    try {
      await flushDisconnect(decodeURIComponent(key.slice(disconnectPrefix.length)));
    } catch (error) {
      if (error instanceof ApiError && error.status === 409) continue;
      return;
    }
  }
  const active = await currentInstallationId();
  if (!active) return;
  const keys = (await AsyncStorage.getAllKeys()).filter(key => key.startsWith(eventPrefix)).sort();
  for (const key of keys) {
    const raw = await AsyncStorage.getItem(key);
    if (!raw) continue;
    const event = JSON.parse(raw) as QueuedEvent;
    if (event.installationId !== active || Date.now() - Date.parse(event.observedAt) > 29 * 60_000) {
      await AsyncStorage.removeItem(key);
      continue;
    }
    try {
      await api("/v1/signals/location", { method: "POST", body: JSON.stringify(event) });
      await AsyncStorage.removeItem(key);
    } catch (error) {
      if (error instanceof ApiError && [403, 409, 422].includes(error.status)) { await AsyncStorage.removeItem(key); continue; }
      return;
    }
  }
}

async function flushDisconnect(ownerId: string) {
  if (!ownerId) throw new Error("Could not identify the account for this place.");
  const key = `${disconnectPrefix}${encodeURIComponent(ownerId)}`;
  if (!(await AsyncStorage.getItem(key))) return;
  await api(`/v1/location/regions?ownerId=${encodeURIComponent(ownerId)}`, { method: "DELETE" });
  await AsyncStorage.removeItem(key);
}

export function flushQueuedEvents() { return serialized(flush); }

TaskManager.defineTask(GEOFENCE_TASK, async ({ data, error }) => {
  if (error || !data) return;
  await serialized(async () => {
    const { eventType, region } = data as { eventType: Location.GeofencingEventType; region: Location.LocationRegion };
    const raw = await AsyncStorage.getItem(regionKey);
    const active = await currentInstallationId();
    if (!raw || !active || (eventType !== Location.GeofencingEventType.Enter && eventType !== Location.GeofencingEventType.Exit)) return;
    const saved = JSON.parse(raw) as Region;
    if (saved.id !== region.identifier) return;
    const event: QueuedEvent = {
      schemaVersion: 1, installationId: active, clientEventId: Crypto.randomUUID(),
      regionId: saved.id, regionVersion: saved.version,
      transition: eventType === Location.GeofencingEventType.Enter ? "enter" : "exit",
      observedAt: new Date().toISOString(),
    };
    await AsyncStorage.setItem(`${eventPrefix}${event.observedAt}.${event.clientEventId}`, JSON.stringify(event));
    await flush();
  });
});

export async function locationStatus() {
  const [permission, raw] = await Promise.all([Location.getBackgroundPermissionsAsync(), AsyncStorage.getItem(regionKey)]);
  return Boolean(permission.granted && raw && (JSON.parse(raw) as Region).trackId && await TaskManager.isTaskRegisteredAsync(GEOFENCE_TASK));
}

export async function reconcileLocationPermission() {
  return serialized(async () => {
    if (!(await AsyncStorage.getItem(regionKey))) return;
    if (!(await Location.getBackgroundPermissionsAsync()).granted) await disconnectLocationNow();
  });
}

export async function selectedLocationTrack() {
  const raw = await AsyncStorage.getItem(regionKey);
  return raw ? (JSON.parse(raw) as Region).trackId : null;
}

export async function enableCurrentPlace(trackId: string) {
  if (!trackId) throw new Error("Choose the learning path for this place.");
  const foreground = await Location.requestForegroundPermissionsAsync();
  if (!foreground.granted) throw new Error("Allow location while using the app to choose this place.");
  const background = await Location.requestBackgroundPermissionsAsync();
  if (!background.granted) throw new Error("Allow Always location access in iOS Settings to recognize this place later.");
  const position = await Location.getCurrentPositionAsync({ accuracy: Location.Accuracy.High });
  if (position.coords.accuracy == null || position.coords.accuracy > 100) throw new Error("Couldn’t locate this place accurately. Try again outdoors or near a window.");
  let enrolled = false;
  try {
    return await serialized(async () => {
      const account = await api<{ learner: { userId: string } }>("/v1/bootstrap");
      const ownerId = account.learner.userId;
      await flushDisconnect(ownerId);
      await AsyncStorage.setItem(ownerKey, ownerId);
      enrolled = true;
      const id = await installationId();
      await api(`/v1/devices/${encodeURIComponent(id)}`, { method: "PUT", body: JSON.stringify({ locationPermission: "enabled" }) });
      await api("/v1/consents", { method: "POST", idempotencyKey: Crypto.randomUUID(), body: JSON.stringify({ purpose: "location_context", policyVersion: "1", granted: true }) });
      const region = await api<Region>("/v1/location/regions", { method: "POST", body: JSON.stringify({ installationId: id, trackId, latitude: position.coords.latitude, longitude: position.coords.longitude }) });
      await Location.startGeofencingAsync(GEOFENCE_TASK, [{ identifier: region.id, latitude: region.latitude, longitude: region.longitude, radius: region.radiusMeters, notifyOnEnter: true, notifyOnExit: true }]);
      await AsyncStorage.setItem(regionKey, JSON.stringify(region));
      return region;
    });
  } catch (error) {
    if (enrolled) await disconnectLocation().catch(() => {});
    throw error;
  }
}

async function clearRegionsNow() {
  await AsyncStorage.multiRemove([regionKey, ownerKey]);
  if (await TaskManager.isTaskRegisteredAsync(GEOFENCE_TASK)) await Location.stopGeofencingAsync(GEOFENCE_TASK);
  const keys = (await AsyncStorage.getAllKeys()).filter(key => key.startsWith(eventPrefix));
  if (keys.length) await AsyncStorage.multiRemove(keys);
}

export function clearRegions() { return serialized(clearRegionsNow); }

async function disconnectLocationNow() {
  let ownerId: string | null = null;
  try {
    ownerId = await AsyncStorage.getItem(ownerKey) ?? (await api<{ learner: { userId: string } }>("/v1/bootstrap")).learner.userId;
    if (!ownerId) throw new Error("Could not identify the account for this place.");
    await AsyncStorage.setItem(`${disconnectPrefix}${encodeURIComponent(ownerId)}`, "1");
  } finally { await clearRegionsNow(); }
  await flushDisconnect(ownerId);
}

export function disconnectLocation() { return serialized(disconnectLocationNow); }
