import app from "./app.json";
import type { ExpoConfig } from "expo/config";

const googleClient = process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID;
const clientPrefix = googleClient?.match(/^([^.]+)\.apps\.googleusercontent\.com$/)?.[1];
const googleScheme = clientPrefix ? `com.googleusercontent.apps.${clientPrefix}` : null;

export default {
  ...app.expo,
  scheme: googleScheme ? ["pausa", googleScheme] : "pausa",
} as ExpoConfig;
