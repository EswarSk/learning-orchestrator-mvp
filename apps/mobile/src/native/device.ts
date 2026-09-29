import * as Crypto from "expo-crypto";
import * as SecureStore from "expo-secure-store";
import { api } from "../api/client";

const installationKey = "location.installation.v1";

export async function installationId() {
  let id = await SecureStore.getItemAsync(installationKey);
  if (!id) {
    id = Crypto.randomUUID();
    await SecureStore.setItemAsync(installationKey, id);
  }
  return id;
}

export async function currentInstallationId() { return SecureStore.getItemAsync(installationKey); }

export async function removeDevice() {
  const id = await currentInstallationId();
  if (id) await api(`/v1/devices/${encodeURIComponent(id)}`, { method: "DELETE" });
  await SecureStore.deleteItemAsync(installationKey);
}
