import { router } from "expo-router";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useAuth } from "../src/api/auth";
import { Brand, Display, Kicker, SolidAction } from "../src/ui/Orbit";
import { colors, serif } from "../src/ui/theme";

export default function Welcome() {
  const { session } = useAuth();
  return <SafeAreaView style={styles.safe}><View style={styles.page}><Brand/><View style={styles.art}><View style={styles.outer}/><View style={styles.inner}/><Text style={styles.mark}>o.</Text><View style={[styles.dot, styles.dotOne]}/><View style={[styles.dot, styles.dotTwo]}/><View style={[styles.dot, styles.dotThree]}/></View><View style={styles.intro}><Kicker>ONE LIFE. MANY THINGS TO LEARN.</Kicker><Display size={49}>Your curiosity{"\n"}<Text style={styles.italic}>has a place.</Text></Display><Text style={styles.copy}>A new language. An instrument. A career leap. Orbit helps you take the next useful step in the life you already have.</Text></View><View style={styles.bottom}><SolidAction label={session ? "Back to Today" : "Create your space"} onPress={() => router.replace(session ? "/(tabs)" : "/sign-in")}/>{!session && <Pressable accessibilityRole="button" onPress={() => router.push("/sign-in")} style={styles.link}><Text style={styles.linkText}>I already have an account</Text></Pressable>}</View></View></SafeAreaView>;
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background }, page: { flex: 1, paddingHorizontal: 25, paddingTop: 10, paddingBottom: 20 },
  art: { flex: 1, maxHeight: 310, minHeight: 220, alignItems: "center", justifyContent: "center", marginTop: 28 }, outer: { position: "absolute", width: 270, height: 270, borderRadius: 135, borderWidth: 1, borderColor: colors.line }, inner: { position: "absolute", width: 162, height: 162, borderRadius: 81, borderWidth: 1, borderColor: colors.line }, mark: { fontFamily: serif, fontSize: 99, letterSpacing: -11, color: colors.ink, marginTop: -10 }, dot: { position: "absolute", width: 10, height: 10, borderRadius: 5, backgroundColor: colors.ink }, dotOne: { left: "10%", top: "35%" }, dotTwo: { right: "10%", top: "55%" }, dotThree: { bottom: 8 },
  intro: { gap: 15 }, italic: { fontStyle: "italic" }, copy: { color: colors.muted, fontSize: 14, lineHeight: 21, maxWidth: 320 }, bottom: { marginTop: "auto", paddingTop: 25 }, link: { minHeight: 44, justifyContent: "center", alignItems: "center" }, linkText: { color: colors.ink, fontSize: 12, textDecorationLine: "underline" },
});
