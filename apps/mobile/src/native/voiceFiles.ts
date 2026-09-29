import * as FileSystem from "expo-file-system/legacy";

export async function readAndDeleteRecording(uri: string) {
  try {
    return await FileSystem.readAsStringAsync(uri, { encoding: FileSystem.EncodingType.Base64 });
  } finally {
    await FileSystem.deleteAsync(uri, { idempotent: true });
  }
}
