import type { ReactNode } from "react";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { colors, serif } from "./theme";

export function Brand({ right, onRight }: { right?: string; onRight?: () => void }) {
  return <View style={styles.brand}><Text style={styles.brandText}>orbit.</Text>{right && <Pressable accessibilityRole="button" onPress={onRight} style={styles.brandAction}><Text style={styles.brandActionText}>{right} ↗</Text></Pressable>}</View>;
}

export function Kicker({ children, style }: { children: ReactNode; style?: object }) {
  return <Text style={[styles.kicker, style]}>{children}</Text>;
}

export function Display({ children, size = 46 }: { children: ReactNode; size?: number }) {
  return <Text style={[styles.display, { fontSize: size, lineHeight: size * 1.04 }]}>{children}</Text>;
}

export function SolidAction({ label, onPress, disabled = false }: { label: string; onPress: () => void; disabled?: boolean }) {
  return <Pressable accessibilityRole="button" accessibilityState={{ disabled }} disabled={disabled} onPress={onPress} style={({ pressed }) => [styles.solid, (pressed || disabled) && styles.dim]}><Text style={styles.solidText}>{label}</Text><Text style={styles.solidArrow}>→</Text></Pressable>;
}

export function RuleRow({ index, title, detail, action, onPress }: { index?: string; title: string; detail?: string; action?: string; onPress?: () => void }) {
  return <Pressable accessibilityRole={onPress ? "button" : undefined} disabled={!onPress} onPress={onPress} style={styles.row}>{index && <Text style={styles.rowIndex}>{index}</Text>}<View style={styles.rowBody}><Text style={styles.rowTitle}>{title}</Text>{detail && <Text style={styles.rowDetail}>{detail}</Text>}</View><Text style={styles.rowAction}>{action ?? (onPress ? "↗" : "")}</Text></Pressable>;
}

const styles = StyleSheet.create({
  brand: { minHeight: 48, flexDirection: "row", alignItems: "center", justifyContent: "space-between" },
  brandText: { fontSize: 25, letterSpacing: -1.5, fontWeight: "600", color: colors.ink },
  brandAction: { minHeight: 44, justifyContent: "center", paddingLeft: 12 }, brandActionText: { fontSize: 12, color: colors.ink },
  kicker: { fontSize: 11, letterSpacing: 1.25, fontWeight: "600", color: colors.muted },
  display: { fontFamily: serif, fontWeight: "400", letterSpacing: -1.8, color: colors.ink },
  solid: { minHeight: 55, backgroundColor: colors.ink, flexDirection: "row", alignItems: "center", justifyContent: "space-between", paddingHorizontal: 17 },
  solidText: { color: colors.background, fontSize: 14, fontWeight: "600" }, solidArrow: { color: colors.background, fontSize: 21 }, dim: { opacity: .55 },
  row: { minHeight: 60, borderTopWidth: 1, borderTopColor: colors.line, flexDirection: "row", alignItems: "center", gap: 12, paddingVertical: 10 },
  rowIndex: { width: 25, fontSize: 11, color: colors.muted }, rowBody: { flex: 1 },
  rowTitle: { fontSize: 16, color: colors.ink, fontWeight: "600" }, rowDetail: { fontSize: 12, color: colors.muted, marginTop: 4, lineHeight: 17 },
  rowAction: { fontSize: 12, color: colors.muted },
});
