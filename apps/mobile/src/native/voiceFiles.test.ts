import { expect, test, vi } from "vitest";

const state = vi.hoisted(() => ({ failRead: false, deleted: [] as string[] }));
vi.mock("expo-file-system/legacy", () => ({
  EncodingType: { Base64: "base64" },
  readAsStringAsync: async () => {
    if (state.failRead) throw new Error("read failed");
    return "audio";
  },
  deleteAsync: async (uri: string) => { state.deleted.push(uri); },
}));

import { readAndDeleteRecording } from "./voiceFiles";

test("temporary recording is deleted after success and read failure", async () => {
  expect(await readAndDeleteRecording("recording-1.m4a")).toBe("audio");
  state.failRead = true;
  await expect(readAndDeleteRecording("recording-2.m4a")).rejects.toThrow("read failed");
  expect(state.deleted).toEqual(["recording-1.m4a", "recording-2.m4a"]);
});
