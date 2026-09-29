import AsyncStorage from "@react-native-async-storage/async-storage";
import { getAccessToken } from "./auth";

const baseUrl = process.env.EXPO_PUBLIC_API_URL;
export class ApiError extends Error { constructor(message: string, readonly status: number) { super(message); } }

export async function api<T>(path: string, init?: RequestInit & { idempotencyKey?: string; expectedToken?: string }): Promise<T> {
  if (!baseUrl) throw new Error("App services are not configured");
  const { expectedToken, idempotencyKey, ...request } = init ?? {};
  const token = await getAccessToken();
  if (expectedToken && token !== expectedToken) throw new Error("Account changed during request");
  const headers = new Headers(request.headers);
  headers.set("authorization", `Bearer ${token}`);
  headers.set("content-type", "application/json");
  if (idempotencyKey) headers.set("idempotency-key", idempotencyKey);
  const response = await fetch(`${baseUrl}${path}`, { ...request, headers });
  const body = await response.json();
  if (await getAccessToken() !== token) throw new Error("Account changed during request");
  if (!response.ok) {
    throw new ApiError(body?.error?.message ?? `Request failed (${response.status})`, response.status);
  }
  return body as T;
}

export async function cached<T>(key: string, load: () => Promise<T>): Promise<T> {
  const token = await getAccessToken();
  try {
    const value = await load();
    if (await getAccessToken() !== token) throw new Error("Account changed");
    await AsyncStorage.setItem(`cache:${key}`, JSON.stringify(value));
    return value;
  } catch (error) {
    if (error instanceof ApiError && error.status < 500) throw error;
    if (await getAccessToken() !== token) throw error;
    const saved = await AsyncStorage.getItem(`cache:${key}`);
    if (saved) return JSON.parse(saved) as T;
    throw error;
  }
}
