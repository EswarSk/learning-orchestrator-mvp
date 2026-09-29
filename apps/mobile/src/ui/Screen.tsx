import type { PropsWithChildren, ReactNode } from "react";
import { ScrollView, StyleSheet, View } from "react-native";
import { ActivityIndicator, Button, Text } from "react-native-paper";
import { colors, spacing } from "./theme";

export function Screen({ children }: PropsWithChildren) {
  return <ScrollView style={styles.screen} contentInsetAdjustmentBehavior="automatic" showsVerticalScrollIndicator={false} contentContainerStyle={styles.content}>{children}</ScrollView>;
}
export function QueryState({ loading, error, retry, children }: { loading: boolean; error: Error | null; retry: () => void; children: ReactNode }) {
  if (loading) return <View style={styles.center}><ActivityIndicator size="large" color={colors.purple} accessibilityLabel="Loading"/><Text style={styles.loading}>Building your next step…</Text></View>;
  if (error) return <Screen><View style={styles.error}><Text variant="headlineSmall" style={styles.errorTitle}>That step didn’t load</Text><Text style={styles.errorText}>{error.message}</Text><Button mode="contained" onPress={retry}>Try again</Button></View></Screen>;
  return children;
}

const styles=StyleSheet.create({
  screen:{flex:1,backgroundColor:colors.background},content:{paddingHorizontal:25,paddingTop:14,gap:spacing.lg,paddingBottom:120},
  center:{flex:1,justifyContent:"center",alignItems:"center",gap:spacing.md,backgroundColor:colors.background},loading:{color:colors.muted,fontWeight:"700"},
  error:{marginTop:80,padding:spacing.lg,borderWidth:1,borderColor:colors.line,backgroundColor:colors.surface,gap:spacing.md},errorTitle:{color:colors.ink,fontWeight:"600"},errorText:{color:colors.muted,lineHeight:22},
});
