import { Redirect, Tabs, router } from "expo-router";
import { useEffect } from "react";
import * as Notifications from "expo-notifications";
import Ionicons from "@expo/vector-icons/Ionicons";
import { StyleSheet, View } from "react-native";
import { colors } from "../../src/ui/theme";
import { useAuth } from "../../src/api/auth";

export default function TabLayout() {
  const { session, loading } = useAuth();
  const notification = Notifications.useLastNotificationResponse();
  useEffect(() => {
    if (session && notification?.notification.request.content.data?.offerId) {
      router.push("/(tabs)/opportunities");
      void Notifications.clearLastNotificationResponseAsync();
    }
  }, [session, notification]);
  if (loading) return null;
  if (!session) return <Redirect href="/sign-in"/>;
  const icon = (name: keyof typeof Ionicons.glyphMap, active: keyof typeof Ionicons.glyphMap) => ({ focused }: { focused: boolean }) => <View style={[styles.icon, focused && styles.selected]}><Ionicons name={focused ? active : name} size={21} color={focused ? colors.background : colors.muted}/></View>;
  return <Tabs screenOptions={{ headerShown: false, tabBarActiveTintColor: colors.ink, tabBarInactiveTintColor: colors.muted, tabBarStyle: styles.bar, tabBarLabelStyle: styles.label, tabBarItemStyle: styles.item, tabBarLabelPosition: "below-icon" }}>
    <Tabs.Screen name="index" options={{ title: "Today", tabBarIcon: icon("home-outline", "home") }}/>
    <Tabs.Screen name="learn" options={{ title: "Learn", tabBarIcon: icon("albums-outline", "albums") }}/>
    <Tabs.Screen name="opportunities" options={{ title: "Moments", tabBarIcon: icon("sparkles-outline", "sparkles") }}/>
    <Tabs.Screen name="settings" options={{ title: "You", tabBarIcon: icon("person-outline", "person") }}/>
  </Tabs>;
}

const styles = StyleSheet.create({
  bar: { height: 85, paddingTop: 7, paddingBottom: 12, borderTopWidth: 1, borderTopColor: colors.line, backgroundColor: colors.background },
  item: { paddingVertical: 2 }, label: { fontSize: 10, fontWeight: "600", letterSpacing: .2, marginTop: 2 }, icon: { width: 43, height: 33, borderRadius: 12, alignItems: "center", justifyContent: "center" }, selected: { backgroundColor: colors.ink },
});
