import Purchases from "react-native-purchases";
import RevenueCatUI from "react-native-purchases-ui";

export async function configurePurchases(userId: string) {
  const apiKey = process.env.EXPO_PUBLIC_REVENUECAT_IOS_KEY;
  if (!apiKey) return false;
  Purchases.configure({ apiKey, appUserID: userId });
  return true;
}

export async function showPaywall() {
  if (!process.env.EXPO_PUBLIC_REVENUECAT_IOS_KEY) return false;
  return RevenueCatUI.presentPaywallIfNeeded({ requiredEntitlementIdentifier: "learning_plus" });
}
