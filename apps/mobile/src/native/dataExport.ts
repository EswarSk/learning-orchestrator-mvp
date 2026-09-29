import { Share } from "react-native";
import * as FileSystem from "expo-file-system/legacy";

const exportName = "orbit-data-export.json";

export async function removeTemporaryExport() {
  if (FileSystem.cacheDirectory) await FileSystem.deleteAsync(`${FileSystem.cacheDirectory}${exportName}`, { idempotent: true });
}

export async function shareDataExport(data: object) {
  if (!FileSystem.cacheDirectory) throw new Error("File storage is unavailable");
  const uri = `${FileSystem.cacheDirectory}${exportName}`;
  await removeTemporaryExport();
  try {
    await FileSystem.writeAsStringAsync(uri, JSON.stringify(data, null, 2));
    await Share.share({ url: uri, title: "Orbit data export" });
  } finally {
    await removeTemporaryExport();
  }
}
