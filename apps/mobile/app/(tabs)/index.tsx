import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { router } from "expo-router";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { api, cached } from "../../src/api/client";
import { Brand, Display, Kicker } from "../../src/ui/Orbit";
import { QueryState, Screen } from "../../src/ui/Screen";
import { colors, serif } from "../../src/ui/theme";
import { reviewStatusLabel } from "../../src/ui/reviewStatus";

type Track = { id: string; subjectId: string; subjectTitle?: string; reviewStatus?: string; goal: string; status: string };
type Subject = { id: string; title: string; category: string };
type Bootstrap = { nextNode: { id: string; title: string; objective: string } | null; nextReason: "new_activity" | "review_due" | "" };
type Offer = { id: string; contextLabel: string; status: string };

export default function Today() {
  const [selectedId, setSelectedId] = useState<string | null>(null);
  useEffect(() => { void AsyncStorage.getItem("selectedTrackId").then(setSelectedId); }, []);
  const tracks = useQuery({ queryKey: ["tracks"], queryFn: () => api<{ items: Track[] }>("/v1/tracks") });
  const subjects = useQuery({ queryKey: ["subjects"], queryFn: () => api<{ items: Subject[] }>("/v1/subjects") });
  const active = tracks.data?.items.find(track => track.id === selectedId) ?? tracks.data?.items[0];
  const bootstrap = useQuery({ queryKey: ["home", active?.id], enabled: Boolean(active), queryFn: () => cached(`bootstrap:${active!.id}`, () => api<Bootstrap>(`/v1/bootstrap?trackId=${encodeURIComponent(active!.id)}`)) });
  const offers = useQuery({ queryKey: ["opportunities"], queryFn: () => api<{ items: Offer[] }>("/v1/opportunities"), refetchInterval: 60_000 });
  const names = new Map(subjects.data?.items.map(subject => [subject.id, subject.title]));
  const offer = offers.data?.items.find(item => item.status === "ready");
  const next = bootstrap.data?.nextNode;
  const date = new Date().toLocaleDateString(undefined, { weekday: "long", day: "numeric", month: "short" }).toUpperCase();
  const openCourse = (id: string) => { void AsyncStorage.setItem("selectedTrackId", id); router.push({ pathname: "/course", params: { trackId: id } }); };

  return <QueryState loading={tracks.isLoading || subjects.isLoading} error={tracks.error ?? subjects.error} retry={() => { void tracks.refetch(); void subjects.refetch(); }}><Screen>
    <Brand/>
    <View style={styles.intro}><Kicker>{date}</Kicker><Display size={52}>Today.</Display></View>
    <Pressable accessibilityRole="button" onPress={() => offer ? router.push("/(tabs)/opportunities") : active ? openCourse(active.id) : router.push("/new-track")} style={styles.opening}>
      <View style={styles.openingHead}><Text style={styles.reverseLabel}>{offer ? "MOMENT AVAILABLE" : bootstrap.data?.nextReason === "review_due" ? "REVIEW DUE" : "UP NEXT"}</Text><Text style={styles.reverseLabel}>{active ? (active.subjectTitle ?? names.get(active.subjectId) ?? "YOUR PATH").toUpperCase() : "START HERE"}</Text></View>
      <View><Text style={styles.openingTitle}>{offer?.contextLabel ?? next?.title ?? (active ? "Choose your next step" : "Start your first path")}</Text>{active && !offer && reviewStatusLabel(active.reviewStatus) && <Text style={styles.reverseLabel}>{reviewStatusLabel(active.reviewStatus)?.toUpperCase()}</Text>}</View>
      <View style={styles.openingAction}><Text style={styles.openingActionText}>{offer ? "See moment" : active ? "Continue learning" : "Add a learning"}</Text><Text style={styles.openingActionText}>→</Text></View>
    </Pressable>
    <Pressable accessibilityRole="button" onPress={() => router.push("/(tabs)/learn")} style={styles.allPaths}><View><Kicker>YOUR LEARNINGS</Kicker><Text style={styles.pathsTitle}>{tracks.data?.items.length ?? 0} {(tracks.data?.items.length ?? 0) === 1 ? "path" : "paths"}</Text></View><Text style={styles.pathsArrow}>→</Text></Pressable>
  </Screen></QueryState>;
}

const styles = StyleSheet.create({
  intro: { gap: 12, marginTop: 20 },
  opening: { backgroundColor: colors.ink, marginTop: 5, paddingTop: 19, paddingHorizontal: 19, minHeight: 238, justifyContent: "space-between" },
  openingHead: { flexDirection: "row", justifyContent: "space-between", gap: 10 }, reverseLabel: { color: colors.background, opacity: .7, fontSize: 10, letterSpacing: 1.15 },
  openingTitle: { fontFamily: serif, fontSize: 35, lineHeight: 38, color: colors.background, marginVertical: 25, maxWidth: 280 },
  openingAction: { minHeight: 50, marginHorizontal: -19, paddingHorizontal: 19, borderTopWidth: 1, borderTopColor: "#6C6B68", flexDirection: "row", alignItems: "center", justifyContent: "space-between" },
  openingActionText: { color: colors.background, fontSize: 13 }, allPaths: { minHeight: 87, borderTopWidth: 1, borderBottomWidth: 1, borderColor: colors.line, flexDirection: "row", alignItems: "center", justifyContent: "space-between" }, pathsTitle: { fontSize: 18, color: colors.ink, fontWeight: "600", marginTop: 7 }, pathsArrow: { color: colors.ink, fontSize: 21 },
});
