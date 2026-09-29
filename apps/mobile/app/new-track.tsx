import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { router, useLocalSearchParams } from "expo-router";
import { Pressable, StyleSheet, Text, TextInput, View } from "react-native";
import { api } from "../src/api/client";
import { Display, Kicker, SolidAction } from "../src/ui/Orbit";
import { QueryState, Screen } from "../src/ui/Screen";
import { colors } from "../src/ui/theme";
import { reviewStatusLabel } from "../src/ui/reviewStatus";

type Subject = { id: string; title: string; category: string; reviewStatus: string };
type Track = { id: string };

export default function NewTrack() {
  const { onboarding } = useLocalSearchParams<{ onboarding?: string }>();
  const queryClient = useQueryClient();
  const [subjectId, setSubjectId] = useState("");
  const [goal, setGoal] = useState("");
  const subjects = useQuery({ queryKey: ["subjects"], queryFn: () => api<{ items: Subject[] }>("/v1/subjects") });
  const create = useMutation({ mutationFn: () => api<Track>("/v1/tracks", { method: "POST", body: JSON.stringify({ subjectId, goal: goal.trim() }) }), onSuccess: async track => { await AsyncStorage.setItem("selectedTrackId", track.id); await queryClient.invalidateQueries({ queryKey: ["tracks"] }); router.replace(onboarding === "1" ? "/context" : { pathname: "/course", params: { trackId: track.id } }); } });
  return <QueryState loading={subjects.isLoading} error={subjects.error} retry={() => void subjects.refetch()}><Screen>
    <View style={styles.header}><Pressable accessibilityRole="button" accessibilityLabel="Go back" onPress={() => router.canGoBack() ? router.back() : router.replace("/(tabs)")} style={styles.back}><Text style={styles.backText}>←</Text></Pressable><Kicker>NEW LEARNING</Kicker></View>
    <View style={styles.intro}><Kicker>START A PATH</Kicker><Display>What will you{"\n"}<Text style={styles.italic}>learn?</Text></Display></View>
    <View><Kicker>COURSES</Kicker><View style={styles.subjects}>{subjects.data?.items.map(subject => <Pressable key={subject.id} accessibilityRole="radio" accessibilityState={{ selected: subjectId === subject.id }} onPress={() => setSubjectId(subject.id)} style={[styles.subject, subjectId === subject.id && styles.subjectSelected]}><Text style={[styles.subjectText, subjectId === subject.id && styles.subjectTextSelected]}>{subject.title}</Text><Text style={[styles.category, subjectId === subject.id && styles.categorySelected]}>{reviewStatusLabel(subject.reviewStatus) ?? subject.category}</Text></Pressable>)}</View><Text style={styles.honest}>Activities are fixed. AI uses your practice and progress to choose what comes next.</Text></View>
    <View style={styles.goalBox}><Kicker>YOUR GOAL</Kicker><TextInput accessibilityLabel="Specific learning goal" placeholder="What would getting better look like?" placeholderTextColor={colors.muted} value={goal} onChangeText={setGoal} maxLength={160} multiline style={styles.goalInput}/></View>
    {create.error && <Text accessibilityLiveRegion="polite" style={styles.error}>{create.error.message}</Text>}
    <SolidAction label={create.isPending ? "Creating your path…" : "Create learning path"} disabled={!subjectId || goal.trim().length < 3 || create.isPending} onPress={() => create.mutate()}/>
  </Screen></QueryState>;
}

const styles = StyleSheet.create({
  header: { minHeight: 48, flexDirection: "row", alignItems: "center", justifyContent: "space-between" }, back: { minWidth: 44, minHeight: 44, justifyContent: "center" }, backText: { fontSize: 27, color: colors.ink }, intro: { gap: 14, marginTop: 19 }, italic: { fontStyle: "italic" },
  subjects: { marginTop: 10, borderBottomWidth: 1, borderBottomColor: colors.line }, subject: { minHeight: 68, borderTopWidth: 1, borderTopColor: colors.line, flexDirection: "row", alignItems: "center", justifyContent: "space-between", paddingHorizontal: 12 }, subjectSelected: { backgroundColor: colors.ink }, subjectText: { color: colors.ink, fontSize: 16 }, subjectTextSelected: { color: colors.background }, category: { color: colors.muted, fontSize: 11 }, categorySelected: { color: colors.background }, honest: { color: colors.muted, fontSize: 11, lineHeight: 16, marginTop: 10 },
  goalBox: { borderTopWidth: 1, borderTopColor: colors.ink, gap: 9, paddingTop: 15 }, goalInput: { minHeight: 120, borderWidth: 1, borderColor: colors.line, padding: 14, fontSize: 16, lineHeight: 23, color: colors.ink, textAlignVertical: "top" }, error: { fontSize: 12, color: "#8A2929" },
});
