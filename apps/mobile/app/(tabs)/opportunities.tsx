import { useMutation, useQuery } from "@tanstack/react-query";
import { router } from "expo-router";
import { Alert, Pressable, StyleSheet, Text, View } from "react-native";
import { api } from "../../src/api/client";
import { Brand, Display, Kicker, SolidAction } from "../../src/ui/Orbit";
import { QueryState, Screen } from "../../src/ui/Screen";
import { colors, serif } from "../../src/ui/theme";

type Opportunity = { id: string; contextLabel: string; reasonCode: string; validUntil: string; status: string };
type Started = { id: string; prompt: string; hint: string | null; understandingMet: boolean };
const reasonText: Record<string, string> = {
  due_review_at_saved_place: "This skill is due for another try at a place you chose for practice.",
  saved_practice_place: "This step in your learning path fits a place you chose for practice.",
  calendar_availability: "A free window in a calendar you connected may fit this learning step.",
  review_due_in_calendar_gap: "A skill due for review may fit this free window in your calendar.",
};

export default function Moments() {
  const query = useQuery({ queryKey: ["opportunities"], queryFn: () => api<{ items: Opportunity[] }>("/v1/opportunities"), refetchInterval: 60_000 });
  const availability = useQuery({ queryKey: ["practice-availability"], queryFn: () => api<{ tutorAvailable: boolean; voiceAvailable: boolean }>("/v1/bootstrap") });
  const start = useMutation({ mutationFn: ({ item, mode }: { item: Opportunity; mode: "text" | "voice" }) => api<Started>("/v1/sessions", { method: "POST", idempotencyKey: `opportunity-${item.id}-${mode}-${Date.now()}`, body: JSON.stringify({ source: { type: "opportunity", id: item.id }, mode }) }), onSuccess: (session, { item, mode }) => router.push({ pathname: "/practice", params: { id: session.id, prompt: session.prompt, hint: session.hint ?? "", title: item.contextLabel, mode, understandingMet: String(session.understandingMet) } }) });
  const report = useMutation({ mutationFn: ({ item, outcome }: { item: Opportunity; outcome: "applied" | "tried" }) => api(`/v1/opportunities/${item.id}/applications`, { method: "POST", body: JSON.stringify({ outcome }) }), onSuccess: () => void query.refetch() });
  const dismiss = useMutation({ mutationFn: (item: Opportunity) => api(`/v1/opportunities/${item.id}/actions`, { method: "POST", body: JSON.stringify({ action: "dismiss" }) }), onSuccess: () => void query.refetch() });
  const askOutcome = (item: Opportunity) => Alert.alert("Did you use this skill?", "Your answer adjusts your path. It is a self-report, not a scored test.", [{ text: "Not now", style: "cancel" }, { text: "I tried", onPress: () => report.mutate({ item, outcome: "tried" }) }, { text: "I used it", onPress: () => report.mutate({ item, outcome: "applied" }) }]);
  const items = query.data?.items.filter(item => item.status === "ready") ?? [];
  return <QueryState loading={query.isLoading} error={query.error} retry={() => void query.refetch()}><Screen>
    <Brand/>
    <View style={styles.intro}><Kicker>REAL-LIFE PRACTICE</Kicker><Display size={49}>Moments.</Display></View>
    {!items.length && <View style={styles.empty}><Text style={styles.emptyMark}>o.</Text><Text style={styles.emptyTitle}>No invitations right now.</Text><SolidAction label="Explore learnings" onPress={() => router.push("/(tabs)/learn")}/></View>}
    {items.map(item => <View key={item.id} style={styles.moment}><View style={styles.momentHead}><Kicker>AVAILABLE NOW</Kicker><Kicker>UNTIL {new Date(item.validUntil).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })}</Kicker></View><Text style={styles.momentTitle}>{item.contextLabel}</Text><View style={styles.receipt}><Kicker>WHY NOW</Kicker><Text style={styles.reason}>{reasonText[item.reasonCode] ?? item.reasonCode.replaceAll("_", " ")}</Text></View><SolidAction label={start.isPending && start.variables?.item.id === item.id ? "Opening…" : "Practice"} disabled={start.isPending || !availability.data?.tutorAvailable} onPress={() => start.mutate({ item, mode: "text" })}/>{availability.data?.voiceAvailable && <Pressable accessibilityRole="button" disabled={start.isPending} onPress={() => start.mutate({ item, mode: "voice" })} style={styles.textAction}><Text style={styles.textActionLabel}>Practice with voice →</Text></Pressable>}<Pressable accessibilityRole="button" onPress={() => askOutcome(item)} style={styles.textAction}><Text style={styles.textActionLabel}>I used this skill</Text></Pressable><Pressable accessibilityRole="button" disabled={dismiss.isPending} onPress={() => dismiss.mutate(item)} style={styles.textAction}><Text style={styles.textActionLabel}>Dismiss</Text></Pressable>{start.error && start.variables?.item.id === item.id && <Text style={styles.error}>{start.error.message}</Text>}{report.error && report.variables?.item.id === item.id && <Text style={styles.error}>{report.error.message}</Text>}{dismiss.error && dismiss.variables?.id === item.id && <Text style={styles.error}>{dismiss.error.message}</Text>}</View>)}
  </Screen></QueryState>;
}

const styles = StyleSheet.create({
  intro: { gap: 10, marginTop: 12 },
  empty: { borderTopWidth: 1, borderTopColor: colors.ink, gap: 22, paddingTop: 15, marginTop: 12 }, emptyMark: { fontFamily: serif, fontSize: 89, letterSpacing: -9, color: colors.ink, textAlign: "center", marginVertical: 15 }, emptyTitle: { fontFamily: serif, fontSize: 27, color: colors.ink, lineHeight: 30 },
  moment: { borderTopWidth: 1, borderTopColor: colors.ink, paddingTop: 15, gap: 16 }, momentHead: { flexDirection: "row", justifyContent: "space-between" }, momentTitle: { fontFamily: serif, fontSize: 31, lineHeight: 34, color: colors.ink }, receipt: { borderTopWidth: 1, borderBottomWidth: 1, borderColor: colors.line, paddingVertical: 12, gap: 7 }, reason: { color: colors.muted, fontSize: 13, textTransform: "capitalize" }, textAction: { minHeight: 44, alignItems: "center", justifyContent: "center" }, textActionLabel: { fontSize: 12, color: colors.ink, textDecorationLine: "underline" }, error: { color: "#8A2929", fontSize: 12 },
});
