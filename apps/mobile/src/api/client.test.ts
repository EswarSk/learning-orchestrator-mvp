import { expect, test, vi } from "vitest";

const state = vi.hoisted(() => { process.env.EXPO_PUBLIC_API_URL = "https://example.test"; return { token: "account-a-token" }; });
vi.mock("./auth", () => ({ getAccessToken: async () => state.token }));
vi.mock("@react-native-async-storage/async-storage", () => ({ default: {} }));

import { api } from "./client";

test("account-bound requests cannot be sent under a different session", async () => {
  const send = vi.fn(async () => ({ ok: true, json: async () => ({ accepted: true }) }));
  vi.stubGlobal("fetch", send);
  state.token = "account-b-token";
  await expect(api("/calendar", { method: "POST", expectedToken: "account-a-token", body: "{}" })).rejects.toThrow("Account changed");
  expect(send).not.toHaveBeenCalled();

  state.token = "account-a-token";
  await api("/calendar", { method: "POST", expectedToken: "account-a-token", body: "{}" });
  expect(send).toHaveBeenCalledOnce();
  expect(((send.mock.calls[0] as unknown as [string, RequestInit])[1].headers as Headers).get("authorization")).toBe("Bearer account-a-token");
});

test("a response from the previous account is never returned after a switch", async () => {
  state.token = "account-a-token";
  let finish!: (value: { ok: boolean; json: () => Promise<object> }) => void;
  vi.stubGlobal("fetch", () => new Promise(resolve => { finish = resolve; }));
  const request = api("/v1/bootstrap");
  await vi.waitFor(() => expect(finish).toBeTypeOf("function"));
  state.token = "account-b-token";
  finish({ ok: true, json: async () => ({ learner: { userId: "account-a" } }) });
  await expect(request).rejects.toThrow("Account changed during request");
});
