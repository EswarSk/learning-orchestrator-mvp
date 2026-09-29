import { beforeEach, expect, test, vi } from "vitest";

const state = vi.hoisted(() => ({
  storage: new Map<string, string>(),
  owner: "account-a",
  backgroundGranted: true,
  failDelete: true,
  deletes: [] as string[],
  requests: [] as string[],
  stopped: 0,
}));

vi.mock("@react-native-async-storage/async-storage", () => ({ default: {
  getItem: async (key: string) => state.storage.get(key) ?? null,
  setItem: async (key: string, value: string) => { state.storage.set(key, value); },
  removeItem: async (key: string) => { state.storage.delete(key); },
  multiRemove: async (keys: string[]) => { keys.forEach(key => state.storage.delete(key)); },
  getAllKeys: async () => [...state.storage.keys()],
} }));
vi.mock("expo-task-manager", () => ({ defineTask: () => {}, isTaskRegisteredAsync: async () => true }));
vi.mock("expo-location", () => ({
  Accuracy: { High: 1 },
  getBackgroundPermissionsAsync: async () => ({ granted: state.backgroundGranted }),
  requestForegroundPermissionsAsync: async () => ({ granted: true }),
  requestBackgroundPermissionsAsync: async () => ({ granted: true }),
  getCurrentPositionAsync: async () => ({ coords: { accuracy: 5, latitude: 1, longitude: 2 } }),
  startGeofencingAsync: async () => {},
  stopGeofencingAsync: async () => { state.stopped++; },
}));
vi.mock("expo-crypto", () => ({ randomUUID: () => "event" }));
vi.mock("./device", () => ({ currentInstallationId: async () => null, installationId: async () => "device" }));
vi.mock("../api/client", () => {
  class ApiError extends Error { constructor(message: string, readonly status: number) { super(message); } }
  return { ApiError, api: async (path: string) => {
    state.requests.push(path);
    if (path === "/v1/bootstrap") return { learner: { userId: state.owner } };
    if (path.startsWith("/v1/location/regions?")) {
      state.deletes.push(path);
      if (new URL(`http://test${path}`).searchParams.get("ownerId") !== state.owner) throw new ApiError("Account changed", 409);
      if (state.failDelete) throw new Error("offline");
      return { removed: true };
    }
    if (path === "/v1/location/regions") return { id: "place", version: 1, trackId: "track", latitude: 1, longitude: 2, radiusMeters: 150 };
    if (path.startsWith("/v1/devices/") || path === "/v1/consents") return {};
    throw new Error(`Unexpected request: ${path}`);
  } };
});

import { disconnectLocation, enableCurrentPlace, flushQueuedEvents, reconcileLocationPermission } from "./geofences";

beforeEach(() => {
  state.storage.clear();
  state.owner = "account-a";
  state.backgroundGranted = true;
  state.failDelete = true;
  state.deletes = [];
  state.requests = [];
  state.stopped = 0;
});

test("revoking iOS background permission queues account revocation on foreground", async () => {
  state.storage.set("location.owner.v1", "account-a");
  state.storage.set("location.region.v1", "saved-place");
  state.backgroundGranted = false;
  await expect(reconcileLocationPermission()).rejects.toThrow("offline");
  expect(state.stopped).toBe(1);
  expect(state.storage.has("location.region.v1")).toBe(false);
  expect(state.storage.has("location.disconnect.account-a")).toBe(true);
});

test("pending removal must succeed before saving another place", async () => {
  state.storage.set("location.disconnect.account-a", "1");
  await expect(enableCurrentPlace("track")).rejects.toThrow("offline");
  expect(state.requests).toEqual(["/v1/bootstrap", "/v1/location/regions?ownerId=account-a"]);
  expect(state.storage.has("location.disconnect.account-a")).toBe(true);

  state.failDelete = false;
  state.requests = [];
  await enableCurrentPlace("track");
  expect(state.requests).toEqual([
    "/v1/bootstrap", "/v1/location/regions?ownerId=account-a",
    "/v1/devices/device", "/v1/consents", "/v1/location/regions",
  ]);
  expect(state.storage.has("location.disconnect.account-a")).toBe(false);
  expect(state.storage.has("location.region.v1")).toBe(true);
});

test("offline disconnect stops monitoring and retries only for its original account", async () => {
  state.storage.set("location.owner.v1", "account-a");
  state.storage.set("location.region.v1", "saved-place");
  await expect(disconnectLocation()).rejects.toThrow("offline");
  expect(state.stopped).toBe(1);
  expect(state.storage.has("location.region.v1")).toBe(false);
  expect(state.storage.has("location.disconnect.account-a")).toBe(true);

  state.owner = "account-b";
  state.failDelete = false;
  state.storage.set("location.disconnect.account-b", "1");
  state.requests = [];
  await flushQueuedEvents();
  expect(state.deletes).toEqual([
    "/v1/location/regions?ownerId=account-a",
    "/v1/location/regions?ownerId=account-a",
    "/v1/location/regions?ownerId=account-b",
  ]);
  expect(state.requests).toEqual(["/v1/location/regions?ownerId=account-a", "/v1/location/regions?ownerId=account-b"]);
  expect(state.storage.has("location.disconnect.account-a")).toBe(true);
  expect(state.storage.has("location.disconnect.account-b")).toBe(false);

  state.owner = "account-a";
  state.requests = [];
  await flushQueuedEvents();
  expect(state.deletes).toHaveLength(4);
  expect(state.requests).toEqual(["/v1/location/regions?ownerId=account-a"]);
  expect(state.storage.has("location.disconnect.account-a")).toBe(false);
});
