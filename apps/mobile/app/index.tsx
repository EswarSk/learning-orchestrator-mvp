import { Redirect } from "expo-router";
import { ActivityIndicator } from "react-native-paper";
import { View } from "react-native";
import { useQuery } from "@tanstack/react-query";
import { useAuth } from "../src/api/auth";
import { api } from "../src/api/client";

export default function Index() {
  const { session, loading } = useAuth();
  const tracks = useQuery({ queryKey: ["tracks"], enabled: Boolean(session), queryFn: () => api<{ items: { id: string }[] }>("/v1/tracks") });
  if (loading || (session && tracks.isLoading)) return <View style={{ flex: 1, justifyContent: "center" }}><ActivityIndicator accessibilityLabel="Loading account"/></View>;
  if (!session) return <Redirect href="/welcome"/>;
  if (tracks.data?.items.length === 0) return <Redirect href={{ pathname: "/new-track", params: { onboarding: "1" } }}/>;
  return <Redirect href="/(tabs)"/>;
}
