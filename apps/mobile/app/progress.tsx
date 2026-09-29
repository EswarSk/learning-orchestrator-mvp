import { useQuery } from "@tanstack/react-query";
import { router } from "expo-router";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { api, cached } from "../src/api/client";
import { Display, Kicker, RuleRow } from "../src/ui/Orbit";
import { QueryState, Screen } from "../src/ui/Screen";
import { colors, serif } from "../src/ui/theme";

type Progress = { xp: number; streak: number; completedNodes: number; totalNodes: number; appliedSkills: number; practicedSkills: number };

export default function Progress() {
  const query = useQuery({ queryKey: ["progress"], queryFn: () => cached("progress", () => api<Progress>("/v1/progress")) });
  const data = query.data;
  const fraction = data?.totalNodes ? Math.min(1, data.completedNodes / data.totalNodes) : 0;
  return <QueryState loading={query.isLoading} error={query.error} retry={() => void query.refetch()}><Screen>
    <View style={styles.header}><Pressable accessibilityRole="button" accessibilityLabel="Back to You" onPress={() => router.replace("/(tabs)/settings")} style={styles.back}><Text style={styles.backText}>←  You</Text></Pressable></View>
    <View style={styles.intro}><Kicker>ALL LEARNINGS</Kicker><Display size={49}>Progress.</Display></View>
    <View style={styles.feature}><Kicker style={styles.reverseKicker}>ACTIVITIES COMPLETE</Kicker><Text style={styles.bigNumber}>{data?.completedNodes ?? 0}<Text style={styles.total}> / {data?.totalNodes ?? 0}</Text></Text><View style={styles.track}><View style={[styles.fill, { width: `${fraction * 100}%` }]}/></View></View>
    <View style={styles.metrics}><View style={styles.metric}><Text style={styles.metricNumber}>{data?.xp ?? 0}</Text><Kicker>XP EARNED</Kicker></View><View style={[styles.metric, styles.metricSecond]}><Text style={styles.metricNumber}>{data?.streak ?? 0}</Text><Kicker>PRACTICE DAYS</Kicker></View></View>
    <View><Kicker>SKILL EVIDENCE</Kicker><View style={styles.list}><RuleRow title="Practiced" action={String(data?.practicedSkills ?? 0)}/><RuleRow title="Applied" detail="Self-reported" action={String(data?.appliedSkills ?? 0)}/></View></View>
  </Screen></QueryState>;
}

const styles = StyleSheet.create({
  header: { minHeight: 48, flexDirection: "row", alignItems: "center" }, back: { minHeight: 44, justifyContent: "center" }, backText: { color: colors.ink, fontSize: 16 }, intro: { gap: 10, marginTop: 12 },
  feature: { backgroundColor: colors.ink, padding: 18, gap: 7, marginTop: 8 }, reverseKicker: { color: colors.background }, bigNumber: { fontFamily: serif, fontSize: 67, lineHeight: 74, color: colors.background }, total: { fontSize: 25, opacity: .65 }, track: { height: 3, backgroundColor: "#676663", marginTop: 14 }, fill: { height: 3, backgroundColor: colors.background },
  metrics: { flexDirection: "row", borderBottomWidth: 1, borderBottomColor: colors.line, paddingBottom: 15 }, metric: { flex: 1, gap: 3 }, metricSecond: { borderLeftWidth: 1, borderLeftColor: colors.line, paddingLeft: 20 }, metricNumber: { fontFamily: serif, fontSize: 31, color: colors.ink }, list: { marginTop: 8 },
});
