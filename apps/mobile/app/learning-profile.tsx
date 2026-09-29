import { useQuery } from "@tanstack/react-query";
import { router } from "expo-router";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { api } from "../src/api/client";
import { Display, Kicker, RuleRow, SolidAction } from "../src/ui/Orbit";
import { QueryState, Screen } from "../src/ui/Screen";
import { colors, serif } from "../src/ui/theme";
import { reviewStatusLabel } from "../src/ui/reviewStatus";

type Track = { id: string; subjectId: string; subjectTitle?: string; reviewStatus?: string; goal: string };
type Subject = { id: string; title: string };

export default function LearningProfile() {
  const tracks = useQuery({ queryKey: ["tracks"], queryFn: () => api<{ items: Track[] }>("/v1/tracks") });
  const subjects = useQuery({ queryKey: ["subjects"], queryFn: () => api<{ items: Subject[] }>("/v1/subjects") });
  const names = new Map(subjects.data?.items.map(item => [item.id, item.title]));
  return <QueryState loading={tracks.isLoading || subjects.isLoading} error={tracks.error ?? subjects.error} retry={() => { void tracks.refetch(); void subjects.refetch(); }}><Screen>
    <View style={styles.header}><Pressable accessibilityRole="button" accessibilityLabel="Back to You" onPress={() => router.replace("/(tabs)/settings")} style={styles.back}><Text style={styles.backText}>←</Text></Pressable><Kicker>YOU</Kicker></View>
    <View style={styles.intro}><Kicker>LEARNING PROFILE</Kicker><Display size={49}>Your goals.</Display></View>
    <View style={styles.sheet}><Kicker style={styles.reverse}>BASED ON YOUR WORDS</Kicker><Text style={styles.sheetTitle}>{tracks.data?.items.length ?? 0} {tracks.data?.items.length === 1 ? "path" : "paths"}.{"\n"}<Text style={styles.italic}>Your pace.</Text></Text></View>
    <View><Kicker>WHAT YOU WANT TO LEARN</Kicker><View style={styles.paths}>{tracks.data?.items.map((track, index) => <RuleRow key={track.id} index={String(index + 1).padStart(2, "0")} title={track.subjectTitle ?? names.get(track.subjectId) ?? "Learning path"} detail={`${track.goal}${reviewStatusLabel(track.reviewStatus) ? ` · ${reviewStatusLabel(track.reviewStatus)}` : ""}`} onPress={() => router.push({ pathname: "/course", params: { trackId: track.id } })}/>)}</View></View>
    <View style={styles.receipt}><Text style={styles.disclosure}>No AI portrait or connected sources yet. This reflects only the goals you saved.</Text></View>
    <SolidAction label="Go to Today" onPress={() => router.replace("/(tabs)")}/>
  </Screen></QueryState>;
}

const styles = StyleSheet.create({
  header: { minHeight: 48, flexDirection: "row", alignItems: "center", justifyContent: "space-between" }, back: { minWidth: 44, minHeight: 44, justifyContent: "center" }, backText: { color: colors.ink, fontSize: 27 }, intro: { marginTop: 19, gap: 10 }, italic: { fontStyle: "italic" }, sheet: { backgroundColor: colors.ink, padding: 18, gap: 17 }, reverse: { color: colors.background }, sheetTitle: { color: colors.background, fontFamily: serif, fontSize: 35, lineHeight: 39 }, paths: { marginTop: 10, borderBottomWidth: 1, borderBottomColor: colors.line }, receipt: { borderTopWidth: 1, borderTopColor: colors.ink, paddingTop: 14 }, disclosure: { fontSize: 11, lineHeight: 17, color: colors.muted },
});
