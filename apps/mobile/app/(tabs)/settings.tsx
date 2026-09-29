import { useEffect, useState } from "react";
import { Alert, Pressable, StyleSheet, Text, TextInput, View } from "react-native";
import { router } from "expo-router";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { localAuth, useAuth } from "../../src/api/auth";
import { api, cached } from "../../src/api/client";
import { disconnectLocation } from "../../src/native/geofences";
import { shareDataExport } from "../../src/native/dataExport";
import { Brand, Display, Kicker, RuleRow, SolidAction } from "../../src/ui/Orbit";
import { Screen } from "../../src/ui/Screen";
import { colors } from "../../src/ui/theme";

type Profile = { learner: { timezone: string; proactivePaused: boolean; quietStart: string; quietEnd: string } };
type Progress = { completedNodes: number; totalNodes: number; xp: number };

export default function Controls() {
  const queryClient = useQueryClient();
  const { signOut } = useAuth();
  const [profile, setProfile] = useState<Profile["learner"] | null>(null);
  const [updating, setUpdating] = useState(false);
  const [editingQuiet, setEditingQuiet] = useState(false);
  const [quietStart, setQuietStart] = useState("");
  const [quietEnd, setQuietEnd] = useState("");
  const [exporting, setExporting] = useState(false);
  const progress = useQuery({ queryKey: ["progress"], queryFn: () => cached("progress", () => api<Progress>("/v1/progress")) });
  useEffect(() => { let mounted = true; void api<Profile>("/v1/bootstrap").then(data => { if (mounted) setProfile(data.learner); }).catch(() => {}); return () => { mounted = false; }; }, []);
  const updatePause = async (paused: boolean) => {
    if (!profile) return;
    const before = profile;
    setProfile({ ...profile, proactivePaused: paused }); setUpdating(true);
    try { await api("/v1/me", { method: "PATCH", body: JSON.stringify({ proactivePaused: paused }) }); await queryClient.invalidateQueries({ queryKey: ["opportunities"] }); }
    catch { setProfile(before); Alert.alert("Couldn’t update", "Please try that setting again."); }
    finally { setUpdating(false); }
  };
  const saveQuiet = async (start = quietStart, end = quietEnd) => {
    if (updating) return;
    setUpdating(true);
    try {
      const updated = await api<Profile["learner"]>("/v1/me", { method: "PATCH", body: JSON.stringify({ quietHours: { start, end } }) });
      setProfile(updated);
      setEditingQuiet(false);
    } catch {
      Alert.alert("Couldn’t save quiet hours", "Check the 24-hour times and your connection, then try again.");
    } finally {
      setUpdating(false);
    }
  };
  const disableLocation = async () => { try { await disconnectLocation(); Alert.alert("Place removed", "Orbit will no longer watch this practice place. You can also change iOS location access in Settings."); } catch { Alert.alert("Couldn’t finish removing access", "Reopen the app online to retry the account update, then check Connections to confirm removal."); } };
  const leave = () => { void signOut().then(() => router.replace("/sign-in")).catch(() => Alert.alert("Couldn’t sign out", "Connect to the internet and try again so this device can be removed from your account.")); };
  const exportData = async () => {
    if (exporting) return;
    setExporting(true);
    try {
      const data = await api<object>("/v1/me/export");
      await shareDataExport(data);
    } catch {
      Alert.alert("Couldn’t export your data", "Check your connection and try again.");
    } finally {
      setExporting(false);
    }
  };
  const confirmExport = () => Alert.alert("Export your data", "This JSON file includes practice transcripts and saved-place coordinates. Share it only with a destination you trust.", [
    { text: "Cancel", style: "cancel" },
    { text: "Continue", onPress: () => void exportData() },
  ]);

  return <Screen>
    <Brand/>
    <View style={styles.intro}><Kicker>ACCOUNT & SETTINGS</Kicker><Display size={49}>Your space.</Display></View>
    <Pressable accessibilityRole="button" onPress={() => router.push("/progress")} style={styles.progress}><Kicker style={styles.reverse}>YOUR PROGRESS</Kicker><Text style={styles.progressValue}>{progress.data ? `${progress.data.completedNodes}/${progress.data.totalNodes}` : "—"}</Text><Text style={styles.progressLabel}>activities complete <Text>→</Text></Text></Pressable>
    <View style={styles.group}><Kicker>LEARNING</Kicker><RuleRow title="Your profile" onPress={() => router.push("/learning-profile")}/></View>
    <View style={styles.group}>
      <Kicker>SMART MOMENTS</Kicker>
      <Pressable
        accessibilityRole="switch"
        accessibilityLabel="Pause invitations"
        accessibilityState={{ checked: profile?.proactivePaused ?? false, disabled: !profile || updating }}
        disabled={!profile || updating}
        onPress={() => void updatePause(!profile?.proactivePaused)}
        style={({ pressed }) => [styles.toggle, pressed && styles.pressed]}
      >
        <Text style={styles.rowTitle}>Pause invitations</Text>
        <View style={[styles.switchTrack, profile?.proactivePaused && styles.switchTrackOn]}>
          <View style={styles.switchThumb}/>
        </View>
      </Pressable>
      <RuleRow title="Quiet hours" detail={profile ? profile.quietStart === profile.quietEnd ? "Off" : `${profile.quietStart}–${profile.quietEnd}` : "Loading…"} onPress={profile && !updating ? () => { setQuietStart(profile.quietStart); setQuietEnd(profile.quietEnd); setEditingQuiet(!editingQuiet); } : undefined}/>
      {editingQuiet && <View style={styles.quietEditor}>
        <Text style={styles.quietHelp}>No invitations are sent between these times in {profile?.timezone ?? "your time zone"}.</Text>
        <View style={styles.timeFields}>
          <View style={styles.timeField}><Text style={styles.timeLabel}>FROM</Text><TextInput accessibilityLabel="Quiet hours start time" value={quietStart} onChangeText={setQuietStart} maxLength={5} keyboardType="numbers-and-punctuation" selectTextOnFocus style={styles.timeInput}/></View>
          <View style={styles.timeField}><Text style={styles.timeLabel}>UNTIL</Text><TextInput accessibilityLabel="Quiet hours end time" value={quietEnd} onChangeText={setQuietEnd} maxLength={5} keyboardType="numbers-and-punctuation" selectTextOnFocus style={styles.timeInput}/></View>
        </View>
        <SolidAction label={updating ? "Saving…" : "Save quiet hours"} disabled={updating} onPress={() => void saveQuiet()}/>
        {profile?.quietStart !== profile?.quietEnd && <Pressable accessibilityRole="button" disabled={updating} onPress={() => void saveQuiet("00:00", "00:00")} style={styles.quietOff}><Text style={styles.quietOffText}>Turn off quiet hours</Text></Pressable>}
      </View>}
    </View>
    <View style={styles.group}><Kicker>DATA & ACCESS</Kicker><RuleRow title="Connections" detail="Review access" onPress={() => router.push("/context")}/><RuleRow title="Remove saved places" onPress={() => void disableLocation()}/><RuleRow title="Export your data" detail={exporting ? "Preparing…" : "Save a copy"} onPress={confirmExport}/><RuleRow title="Delete account" detail="Not yet available" onPress={() => Alert.alert("Account deletion isn’t ready", "We can’t delete your sign-in identity until an account provider is connected. This app is not ready for public release.")}/>{!localAuth && <RuleRow title="Sign out" onPress={leave}/>}</View>
    {localAuth && <Text style={styles.version}>LOCAL PREVIEW ACCOUNT</Text>}
  </Screen>;
}

const styles = StyleSheet.create({
  intro: { gap: 10, marginTop: 12 }, progress: { minHeight: 170, backgroundColor: colors.ink, padding: 18, justifyContent: "space-between" }, reverse: { color: colors.background }, progressValue: { color: colors.background, fontSize: 52, lineHeight: 57, fontFamily: "Georgia" }, progressLabel: { color: colors.background, fontSize: 12 },
  group: { borderTopWidth: 1, borderTopColor: colors.ink, paddingTop: 12, gap: 4 },
  toggle: { minHeight: 60, borderTopWidth: 1, borderTopColor: colors.line, flexDirection: "row", alignItems: "center", justifyContent: "space-between", gap: 12 },
  pressed: { opacity: .65 }, rowTitle: { fontSize: 16, color: colors.ink, fontWeight: "600" },
  switchTrack: { width: 46, height: 28, borderRadius: 14, backgroundColor: colors.soft, borderWidth: 1, borderColor: colors.muted, justifyContent: "center", paddingHorizontal: 3 },
  switchTrackOn: { backgroundColor: colors.ink, borderColor: colors.ink, alignItems: "flex-end" },
  switchThumb: { width: 20, height: 20, borderRadius: 10, backgroundColor: colors.background },
  version: { fontSize: 10, letterSpacing: 1.1, color: colors.muted, textAlign: "center" },
  quietEditor: { backgroundColor: colors.soft, padding: 15, gap: 14 }, quietHelp: { color: colors.muted, fontSize: 13, lineHeight: 19 },
  timeFields: { flexDirection: "row", gap: 12 }, timeField: { flex: 1, gap: 6 }, timeLabel: { color: colors.muted, fontSize: 11, letterSpacing: 1 },
  timeInput: { minHeight: 50, borderWidth: 1, borderColor: colors.line, backgroundColor: colors.background, paddingHorizontal: 12, color: colors.ink, fontSize: 19 },
  quietOff: { minHeight: 44, alignItems: "center", justifyContent: "center" }, quietOffText: { color: colors.ink, fontSize: 13, textDecorationLine: "underline" },
});
