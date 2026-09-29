import "../src/native/geofences";
import { useEffect } from "react";
import { Stack } from "expo-router";
import { focusManager, QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AppState, LogBox } from "react-native";
import { PaperProvider } from "react-native-paper";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { AuthProvider } from "../src/api/auth";
import { removeTemporaryExport } from "../src/native/dataExport";
import { theme } from "../src/ui/theme";

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: 1, staleTime: 30_000 } } });
if (__DEV__) LogBox.ignoreLogs(["Sending `onAnimatedValueUpdate` with no listeners registered."]);

export default function RootLayout() {
  useEffect(() => { void removeTemporaryExport().catch(error => console.warn("Could not clear temporary data export", error)); }, []);
  useEffect(() => {
    const subscription = AppState.addEventListener("change", state => focusManager.setFocused(state === "active"));
    return () => subscription.remove();
  }, []);
  return <SafeAreaProvider><PaperProvider theme={theme}><QueryClientProvider client={queryClient}><AuthProvider>
    <Stack screenOptions={{ headerShown: false }}><Stack.Screen name="index"/><Stack.Screen name="welcome"/><Stack.Screen name="sign-in"/><Stack.Screen name="(tabs)"/><Stack.Screen name="course"/><Stack.Screen name="progress"/><Stack.Screen name="new-track"/><Stack.Screen name="context"/><Stack.Screen name="learning-profile"/><Stack.Screen name="practice" options={{ presentation: "fullScreenModal" }}/><Stack.Screen name="voice-preview" options={{ presentation: "fullScreenModal" }}/></Stack>
  </AuthProvider></QueryClientProvider></PaperProvider></SafeAreaProvider>;
}
