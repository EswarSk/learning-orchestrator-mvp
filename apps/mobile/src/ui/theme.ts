import { MD3LightTheme } from "react-native-paper";

export const colors = {
  ink: "#171717", muted: "#696863", background: "#F7F6F2", surface: "#F7F6F2", line: "#D8D6D0", soft: "#EAE8E2",
  purple: "#171717", purpleDark: "#171717", purpleSoft: "#EAE8E2", teal: "#171717", tealSoft: "#EAE8E2",
  orange: "#171717", orangeSoft: "#EAE8E2", coral: "#171717",
} as const;

export const serif = "Georgia";

export const theme = {
  ...MD3LightTheme,
  roundness: 0,
  colors: {
    ...MD3LightTheme.colors,
    primary: colors.purple,
    onPrimary: "#FFFFFF",
    primaryContainer: colors.purpleSoft,
    onPrimaryContainer: colors.ink,
    secondary: colors.teal,
    secondaryContainer: colors.tealSoft,
    background: colors.background,
    surface: colors.surface,
    surfaceVariant: colors.soft,
    outline: colors.line,
    error: "#BA1A1A"
  }
};

export const spacing = { xs: 4, sm: 8, md: 16, lg: 24, xl: 32, xxl: 40 } as const;
