import { useMutation, useQuery } from "@tanstack/react-query";
import { router, useLocalSearchParams } from "expo-router";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { api, cached } from "../src/api/client";
import { Display, Kicker, SolidAction } from "../src/ui/Orbit";
import { QueryState, Screen } from "../src/ui/Screen";
import { colors, serif } from "../src/ui/theme";
import { reviewStatusLabel } from "../src/ui/reviewStatus";

type Node = { id: string; title: string; objective: string };
type Milestone = { id: string; title: string; nodes: Node[] };
type Curriculum = { trackId: string; reviewStatus: string; subject: { title: string; category: string }; milestones: Milestone[]; completedNodeIds: string[] };
type Bootstrap = { nextNode: Node | null; nextReason: "new_activity" | "review_due" | ""; dueReviewCount: number; tutorAvailable: boolean; voiceAvailable: boolean };
type Started = { id: string; prompt: string; hint: string | null; understandingMet: boolean };

export default function Course() {
  const { trackId } = useLocalSearchParams<{ trackId: string }>();
  const curriculum = useQuery({ queryKey: ["curriculum", trackId], enabled: Boolean(trackId), queryFn: () => cached(`curriculum:${trackId}`, () => api<Curriculum>(`/v1/curriculum?trackId=${encodeURIComponent(trackId!)}`)) });
  const bootstrap = useQuery({ queryKey: ["home", trackId], enabled: Boolean(trackId), queryFn: () => cached(`bootstrap:${trackId}`, () => api<Bootstrap>(`/v1/bootstrap?trackId=${encodeURIComponent(trackId!)}`)) });
  const data = curriculum.data;
  const done = new Set(data?.completedNodeIds ?? []);
  const milestones = data?.milestones ?? [];
  const currentMilestoneIndex = milestones.findIndex(group => group.nodes.some(node => node.id === bootstrap.data?.nextNode?.id));
  const total = milestones.reduce((count, group) => count + group.nodes.length, 0);
  const current = bootstrap.data?.nextNode;
  const start = useMutation({ mutationFn: ({ node, mode }: { node: Node; mode: "text" | "voice" }) => api<Started>("/v1/sessions", { method: "POST", idempotencyKey: `mobile-${mode}-${Date.now()}`, body: JSON.stringify({ source: { type: "curriculum", nodeId: node.id, trackId }, mode }) }), onSuccess: (session, { node, mode }) => router.push({ pathname: "/practice", params: { id: session.id, prompt: session.prompt, hint: session.hint ?? "", title: node.title, mode, understandingMet: String(session.understandingMet) } }) });

  return <QueryState loading={curriculum.isLoading || bootstrap.isLoading} error={curriculum.error ?? bootstrap.error} retry={() => { void curriculum.refetch(); void bootstrap.refetch(); }}><Screen>
    <View style={styles.header}><Pressable accessibilityRole="button" accessibilityLabel="All learnings" onPress={() => router.replace("/(tabs)/learn")} style={styles.back}><Text style={styles.backText}>←</Text><Text style={styles.backLabel}>Learnings</Text></Pressable></View>
    <View style={styles.headline}><Kicker>YOUR COURSE</Kicker><Display size={48}>{data?.subject.title ?? "Your path"}</Display>{reviewStatusLabel(data?.reviewStatus) && <Text style={styles.disclosure}>{reviewStatusLabel(data?.reviewStatus)}</Text>}</View>
    {current ? <View style={styles.next}><Kicker>{bootstrap.data?.nextReason === "review_due" ? "REVIEW DUE" : "NEXT ACTIVITY"}</Kicker><Text style={styles.nextTitle}>{current.title}</Text><Text style={styles.nextCopy}>{current.objective}</Text><SolidAction label={start.isPending ? "Opening…" : "Start practice"} disabled={start.isPending || !bootstrap.data?.tutorAvailable} onPress={() => start.mutate({ node: current, mode: "text" })}/>{bootstrap.data?.voiceAvailable && <Pressable accessibilityRole="button" disabled={start.isPending} onPress={() => start.mutate({ node: current, mode: "voice" })} style={styles.preview}><Text style={styles.previewText}>Practice with voice →</Text></Pressable>}{!bootstrap.data?.tutorAvailable && <Text style={styles.error}>Practice needs the OpenAI server key. The course remains available to browse.</Text>}{start.error && <Text style={styles.error}>{start.error.message}</Text>}</View> : <View style={styles.next}><Kicker>COURSE COMPLETE</Kicker><Text style={styles.nextTitle}>You’ve finished this path.</Text></View>}
    <View style={styles.metrics}><View style={styles.metric}><Text style={styles.metricValue}>{done.size}/{total}</Text><Kicker>COMPLETE</Kicker></View><View style={[styles.metric, styles.metricSecond]}><Text style={styles.metricValue}>{bootstrap.data?.dueReviewCount ?? 0}</Text><Kicker>REVIEWS DUE</Kicker></View></View>
    <View><View style={styles.mapHeading}><Kicker>LEVELS</Kicker><Kicker>{milestones.length} TOTAL</Kicker></View>{milestones.map((group, index) => {
      const complete = group.nodes.every(node => done.has(node.id));
      const selected = index === currentMilestoneIndex;
      return <View key={group.id} style={styles.level}><View style={[styles.levelNode, complete && styles.levelDone, selected && !complete && styles.levelCurrent]}><Text style={[styles.levelNumber, complete && styles.levelDoneText]}>{complete ? "✓" : String(index + 1).padStart(2, "0")}</Text></View><View style={styles.levelBody}><Text style={styles.levelTitle}>{group.title}</Text><Text style={styles.levelDetail}>{group.nodes.filter(node => done.has(node.id)).length}/{group.nodes.length} complete</Text></View><Text style={styles.levelState}>{complete ? "DONE" : selected ? "CURRENT" : "LATER"}</Text></View>;
    })}</View>
  </Screen></QueryState>;
}

const styles = StyleSheet.create({
  header: { minHeight: 48, flexDirection: "row", alignItems: "center" }, back: { minWidth: 44, minHeight: 44, flexDirection: "row", alignItems: "center", gap: 9 }, backText: { fontSize: 27, color: colors.ink }, backLabel: { color: colors.ink, fontSize: 13 },
  headline: { gap: 10, marginTop: 13 },
  disclosure: { color: colors.muted, fontSize: 12 },
  metrics: { flexDirection: "row", borderBottomWidth: 1, borderColor: colors.line, paddingBottom: 15 }, metric: { flex: 1, gap: 2 }, metricSecond: { borderLeftWidth: 1, borderColor: colors.line, paddingLeft: 18 }, metricValue: { fontFamily: serif, fontSize: 29, color: colors.ink },
  mapHeading: { flexDirection: "row", justifyContent: "space-between", marginBottom: 8 }, level: { minHeight: 59, flexDirection: "row", alignItems: "center", gap: 12 }, levelNode: { width: 45, height: 45, borderRadius: 23, borderWidth: 1, borderColor: colors.line, alignItems: "center", justifyContent: "center", backgroundColor: colors.background }, levelCurrent: { borderWidth: 2, borderColor: colors.ink }, levelDone: { backgroundColor: colors.ink, borderColor: colors.ink }, levelNumber: { fontFamily: serif, color: colors.ink, fontSize: 18 }, levelDoneText: { color: colors.background, fontFamily: undefined }, levelBody: { flex: 1 }, levelTitle: { fontSize: 13, fontWeight: "600", color: colors.ink }, levelDetail: { fontSize: 11, color: colors.muted, marginTop: 3 }, levelState: { fontSize: 10, letterSpacing: .6, color: colors.muted },
  next: { padding: 17, backgroundColor: colors.soft, gap: 11 }, nextTitle: { fontFamily: serif, fontSize: 27, lineHeight: 30, color: colors.ink }, nextCopy: { color: colors.muted, fontSize: 12, lineHeight: 18 }, preview: { minHeight: 44, justifyContent: "center", alignItems: "center" }, previewText: { color: colors.ink, fontSize: 12, textDecorationLine: "underline" }, error: { color: "#8A2929", fontSize: 12 },
});
