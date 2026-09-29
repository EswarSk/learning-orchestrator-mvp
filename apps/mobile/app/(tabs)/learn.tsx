import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { router } from "expo-router";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { api } from "../../src/api/client";
import { Brand, Display, Kicker } from "../../src/ui/Orbit";
import { QueryState, Screen } from "../../src/ui/Screen";
import { colors, serif } from "../../src/ui/theme";
import { reviewStatusLabel } from "../../src/ui/reviewStatus";

type Track = { id: string; subjectId: string; subjectTitle?: string; reviewStatus?: string; goal: string };
type Subject = { id: string; title: string };

export default function Learn() {
  const [selectedId, setSelectedId] = useState<string | null>(null);
  useEffect(() => { void AsyncStorage.getItem("selectedTrackId").then(setSelectedId); }, []);
  const tracks = useQuery({ queryKey: ["tracks"], queryFn: () => api<{ items: Track[] }>("/v1/tracks") });
  const subjects = useQuery({ queryKey: ["subjects"], queryFn: () => api<{ items: Subject[] }>("/v1/subjects") });
  const names = new Map(subjects.data?.items.map(subject => [subject.id, subject.title]));
  const open = async (id: string) => { await AsyncStorage.setItem("selectedTrackId", id); setSelectedId(id); router.push({ pathname: "/course", params: { trackId: id } }); };
  return <QueryState loading={tracks.isLoading || subjects.isLoading} error={tracks.error ?? subjects.error} retry={() => { void tracks.refetch(); void subjects.refetch(); }}><Screen>
    <Brand/>
    <View style={styles.intro}><Kicker>YOUR LEARNINGS</Kicker><Display>Keep growing.</Display></View>
    <View style={styles.list}>{tracks.data?.items.map((track, index) => <Pressable key={track.id} accessibilityRole="button" onPress={() => void open(track.id)} style={styles.card}>
      <View style={styles.cardTop}><Kicker>PATH {String(index + 1).padStart(2, "0")}</Kicker>{(selectedId ?? tracks.data?.items[0]?.id) === track.id && <Text style={styles.current}>CURRENT</Text>}</View>
      <Text style={styles.cardTitle}>{track.subjectTitle ?? names.get(track.subjectId) ?? "Learning path"}</Text>{reviewStatusLabel(track.reviewStatus) && <Text style={styles.generated}>{reviewStatusLabel(track.reviewStatus)}</Text>}<Text style={styles.goal} numberOfLines={2}>{track.goal}</Text><Text style={styles.open}>Open course <Text>→</Text></Text>
    </Pressable>)}</View>
    <Pressable accessibilityRole="button" onPress={() => router.push("/new-track")} style={styles.add}><Text style={styles.addPlus}>＋</Text><Text style={styles.addText}>Add a learning</Text><Text style={styles.addArrow}>→</Text></Pressable>
  </Screen></QueryState>;
}

const styles = StyleSheet.create({
  intro: { gap: 10, marginTop: 12 }, list: { gap: 12 }, card: { borderWidth: 1, borderColor: colors.ink, padding: 18, gap: 10, minHeight: 173 }, cardTop: { flexDirection: "row", justifyContent: "space-between", alignItems: "center" }, current: { fontSize: 10, color: colors.ink, letterSpacing: 1.1 }, cardTitle: { fontFamily: serif, color: colors.ink, fontSize: 32 }, generated: { color: colors.muted, fontSize: 11 }, goal: { color: colors.muted, fontSize: 13, lineHeight: 19 }, open: { color: colors.ink, fontSize: 12, fontWeight: "600", marginTop: "auto" }, add: { borderWidth: 1, borderColor: colors.line, minHeight: 66, flexDirection: "row", alignItems: "center", gap: 13, paddingHorizontal: 18 }, addPlus: { color: colors.ink, fontSize: 23 }, addText: { flex: 1, color: colors.ink, fontSize: 16, fontWeight: "600" }, addArrow: { color: colors.ink, fontSize: 18 },
});
