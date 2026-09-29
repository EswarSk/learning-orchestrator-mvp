import * as Notifications from "expo-notifications";
import Constants from "expo-constants";
import * as Crypto from "expo-crypto";
import { api } from "../api/client";
import { installationId } from "./device";

Notifications.setNotificationHandler({ handleNotification: async () => ({
  shouldShowBanner: true, shouldShowList: true, shouldPlaySound: false, shouldSetBadge: false,
}) });

export async function requestPushToken() {
  const permission = await Notifications.requestPermissionsAsync();
  if (!permission.granted) throw new Error("Allow notifications in iOS Settings to receive learning invitations.");
  const projectId = Constants.easConfig?.projectId ?? Constants.expoConfig?.extra?.eas?.projectId ?? process.env.EXPO_PUBLIC_EAS_PROJECT_ID;
  if (!projectId) throw new Error("Push notifications are not configured for this build.");
  return (await Notifications.getExpoPushTokenAsync({ projectId })).data;
}

export async function notificationStatus() {
  const id = await installationId();
  const [server, permission] = await Promise.all([
    api<{ available: boolean; enabled: boolean }>(`/v1/notifications?installationId=${encodeURIComponent(id)}`),
    Notifications.getPermissionsAsync(),
  ]);
  return { ...server, systemAllowed: permission.granted };
}

export async function enableNotifications() {
  const token = await requestPushToken();
  const id = await installationId();
  await api(`/v1/devices/${encodeURIComponent(id)}`, { method: "PUT", body: JSON.stringify({ pushToken: token, notificationPermission: "granted" }) });
  await api("/v1/consents", { method: "POST", idempotencyKey: Crypto.randomUUID(), body: JSON.stringify({ purpose: "notifications", policyVersion: "1", granted: true }) });
}

export async function disableNotifications() {
  await api("/v1/consents", { method: "POST", idempotencyKey: Crypto.randomUUID(), body: JSON.stringify({ purpose: "notifications", policyVersion: "1", granted: false }) });
  const id = await installationId();
  await api(`/v1/devices/${encodeURIComponent(id)}`, { method: "PUT", body: JSON.stringify({ pushToken: null, notificationPermission: "denied" }) });
}
