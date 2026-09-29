import { useState } from "react";
import { Redirect, router } from "expo-router";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useAuth } from "../src/api/auth";
import { Display, Kicker, SolidAction } from "../src/ui/Orbit";
import { colors, serif } from "../src/ui/theme";

export default function SignIn() {
  const { session, configured, signIn } = useAuth();
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const start = async () => { setBusy(true); setMessage(""); try { await signIn(); } catch (error) { setMessage(error instanceof Error ? error.message : "Sign-in is unavailable right now."); } finally { setBusy(false); } };
  if (session) return <Redirect href="/"/>;
  return <SafeAreaView style={styles.safe}><View style={styles.page}>
    <View style={styles.header}><Pressable accessibilityRole="button" accessibilityLabel="Back to welcome" onPress={() => router.back()} style={styles.back}><Text style={styles.backText}>←</Text></Pressable><Kicker>01 / ACCOUNT</Kicker></View>
    <Text style={styles.mark}>o.</Text><View style={styles.intro}><Kicker>A HOME FOR EVERY PATH</Kicker><Display size={49}>Make it{"\n"}<Text style={styles.italic}>yours.</Text></Display><Text style={styles.copy}>Save what you’re learning and pick up where you left off.</Text></View>
    <View style={styles.bottom}>{!configured && <Text style={styles.error}>Sign-in services are not configured for this build.</Text>}<SolidAction label={busy ? "Opening account…" : "Create account or sign in"} disabled={!configured || busy} onPress={() => void start()}/>{message ? <Text accessibilityLiveRegion="polite" style={styles.error}>{message}</Text> : null}<Text style={styles.note}>Your identity provider will open securely. Your learning context is not shared with it.</Text></View>
  </View></SafeAreaView>;
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background }, page: { flex: 1, paddingHorizontal: 25, paddingTop: 10, paddingBottom: 20 }, header: { height: 48, flexDirection: "row", justifyContent: "space-between", alignItems: "center" }, back: { minHeight: 44, minWidth: 44, justifyContent: "center" }, backText: { fontSize: 26, color: colors.ink }, mark: { fontFamily: serif, color: colors.ink, fontSize: 145, letterSpacing: -19, lineHeight: 155, marginTop: 58, marginBottom: 33 }, intro: { gap: 14 }, italic: { fontStyle: "italic" }, copy: { fontSize: 14, color: colors.muted, lineHeight: 21 }, bottom: { marginTop: "auto", gap: 13 }, note: { color: colors.muted, fontSize: 11, lineHeight: 16, textAlign: "center" }, error: { color: "#8A2929", fontSize: 12, lineHeight: 18 },
});
