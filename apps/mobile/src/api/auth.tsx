import { createContext, useContext, useEffect, useState, type PropsWithChildren } from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { useQueryClient } from "@tanstack/react-query";
import { AppState, Platform } from "react-native";
import * as AuthSession from "expo-auth-session";
import * as SecureStore from "expo-secure-store";
import { LOCAL_AUTH_TOKEN } from "@orchestrator/contracts";

const issuer = process.env.EXPO_PUBLIC_OIDC_ISSUER;
const clientId = process.env.EXPO_PUBLIC_OIDC_CLIENT_ID;
const redirectUri = AuthSession.makeRedirectUri({ scheme: "pausa", path: "auth", native: "pausa://auth" });
const storageKey = "orchestrator.auth.v1";
export const localAuth = __DEV__ && !(issuer && clientId);

type Session = { accessToken: string; refreshToken?: string; expiresAt: number; issuer?: string; clientId?: string };
type AuthValue = {
  session: Session | null;
  loading: boolean;
  configured: boolean;
  readyToSignIn: boolean;
  signIn: () => Promise<void>;
  signOut: () => Promise<void>;
};

let current: Session | null = localAuth ? { accessToken: LOCAL_AUTH_TOKEN, expiresAt: Infinity } : null;
let hydrated = localAuth;
let hydrating: Promise<void> | null = null;
let refreshing: Promise<string> | null = null;
let storageWrite: Promise<void> = Promise.resolve();
let authGeneration = 0;
const listeners = new Set<(session: Session | null) => void>();

function publish(session: Session | null) {
  current = session;
  for (const listener of listeners) listener(session);
}

async function hydrate() {
  if (hydrated) return;
  hydrating ??= (async () => {
    const raw = await SecureStore.getItemAsync(storageKey);
    if (raw && issuer && clientId) {
      try {
        const saved = JSON.parse(raw) as Session;
        if (typeof saved.accessToken === "string" && typeof saved.expiresAt === "number" && saved.issuer === issuer && saved.clientId === clientId) publish(saved);
        else await SecureStore.deleteItemAsync(storageKey);
      } catch { await SecureStore.deleteItemAsync(storageKey); }
    }
    hydrated = true;
  })();
  try { await hydrating; } finally { hydrating = null; }
}

async function clearPrivateCache() {
  const keys = (await AsyncStorage.getAllKeys()).filter(key => key.startsWith("cache:") || key === "selectedTrackId" || key === "calendar.ios.selection.v1");
  if (keys.length) await AsyncStorage.multiRemove(keys);
}

async function storeSession(session: Session | null) {
  const update = storageWrite.then(async () => {
    if (session) await SecureStore.setItemAsync(storageKey, JSON.stringify(session));
    else await SecureStore.deleteItemAsync(storageKey);
    publish(session);
  });
  storageWrite = update.catch(() => {});
  await update;
}

function sessionFromToken(token: AuthSession.TokenResponse, priorRefreshToken?: string): Session {
  if (!token.accessToken || !token.expiresIn || token.expiresIn <= 0 || token.accessToken.split(".").length !== 3) {
    throw new Error("Sign-in provider must issue an expiring JWT access token for this app.");
  }
  return {
    accessToken: token.accessToken,
    refreshToken: token.refreshToken ?? priorRefreshToken,
    expiresAt: token.issuedAt + token.expiresIn,
    issuer,
    clientId,
  };
}

export async function getAccessToken(): Promise<string> {
  await hydrate();
  if (!localAuth && (!issuer || !clientId)) throw new Error("Sign-in services are not configured");
  if (!current) throw new Error("Sign in required");
  if (localAuth || current.expiresAt > Date.now() / 1000 + 60) return current.accessToken;
  if (!issuer || !clientId || !current.refreshToken) {
    await storeSession(null);
    await clearPrivateCache();
    void import("../native/geofences").then(({ clearRegions }) => clearRegions()).catch(() => {});
    void import("../native/googleCalendar").then(({ clearGoogleCalendarOnSignOut }) => clearGoogleCalendarOnSignOut()).catch(() => {});
    throw new Error("Sign in required");
  }
  refreshing ??= (async () => {
    const prior = current!;
    const generation = authGeneration;
    const discovery = await AuthSession.fetchDiscoveryAsync(issuer);
    const refreshed = await AuthSession.refreshAsync({ clientId, refreshToken: prior.refreshToken }, discovery);
    if (current !== prior || authGeneration !== generation) throw new Error("Sign in required");
    const session = sessionFromToken(refreshed, prior.refreshToken);
    await storeSession(session);
    return session.accessToken;
  })();
  try { return await refreshing; } finally { refreshing = null; }
}

const AuthContext = createContext<AuthValue>({
  session: null, loading: true, configured: false, readyToSignIn: false,
  signIn: async () => {}, signOut: async () => {},
});

export function AuthProvider({ children }: PropsWithChildren) {
  const queryClient = useQueryClient();
  const [session, setSession] = useState<Session | null>(current);
  const [loading, setLoading] = useState(!hydrated);

  useEffect(() => {
    listeners.add(setSession);
    void hydrate().catch(() => {}).finally(() => setLoading(false));
    return () => { listeners.delete(setSession); };
  }, []);
  useEffect(() => {
    const retry = () => { void import("../native/googleCalendar").then(({ retryGoogleCalendarRevocations }) => retryGoogleCalendarRevocations()).catch(() => {}); };
    retry();
    const subscription = AppState.addEventListener("change", state => { if (state === "active") retry(); });
    return () => subscription.remove();
  }, []);
  useEffect(() => { if (!loading && !session) queryClient.clear(); }, [loading, session, queryClient]);
  useEffect(() => {
    if (!session) return;
    let lastZone = "";
    let syncingZone = false;
    const flush = () => {
      void import("../native/geofences").then(async ({ reconcileLocationPermission, flushQueuedEvents }) => {
        await reconcileLocationPermission().catch(() => {});
        await flushQueuedEvents();
      }).catch(() => {});
      if (Platform.OS === "ios") void import("../native/iosCalendar").then(({ reconcileIOSCalendarPermission }) => reconcileIOSCalendarPermission()).then(() => queryClient.invalidateQueries({ queryKey: ["opportunities"] })).catch(() => {});
      if (Platform.OS === "ios") void import("../native/googleCalendar").then(({ reconcileGoogleCalendar }) => reconcileGoogleCalendar()).then(() => queryClient.invalidateQueries({ queryKey: ["opportunities"] })).catch(() => {});
      let zone = "";
      try { zone = Intl.DateTimeFormat().resolvedOptions().timeZone; } catch { return; }
      if (!zone || zone === lastZone || syncingZone) return;
      syncingZone = true;
      void import("./client").then(async ({ api }) => {
        const account = await api<{ learner: { timezone: string } }>("/v1/bootstrap");
        if (account.learner.timezone !== zone) await api("/v1/me", { method: "PATCH", body: JSON.stringify({ timezone: zone }) });
        lastZone = zone;
      }).catch(() => {}).finally(() => { syncingZone = false; });
    };
    flush();
    const subscription = AppState.addEventListener("change", state => { if (state === "active") flush(); });
    return () => subscription.remove();
  }, [session]);

  const signIn = async () => {
    if (!issuer || !clientId) throw new Error("Sign-in services are unavailable. Please try again.");
    authGeneration++;
    const discovery = await AuthSession.fetchDiscoveryAsync(issuer);
    const request = new AuthSession.AuthRequest({
      clientId, redirectUri, scopes: ["openid", "profile", "offline_access"],
      responseType: AuthSession.ResponseType.Code, usePKCE: true,
    });
    const result = await request.promptAsync(discovery);
    if (result.type === "cancel" || result.type === "dismiss") return;
    if (result.type !== "success" || !result.params.code || !request.codeVerifier) throw new Error("Sign-in could not be completed.");
    const token = await AuthSession.exchangeCodeAsync({
      clientId, code: result.params.code, redirectUri,
      extraParams: { code_verifier: request.codeVerifier },
    }, discovery);
    const next = sessionFromToken(token);
    await (await import("../native/geofences")).clearRegions();
    await (await import("../native/googleCalendar")).clearGoogleCalendarOnSignOut();
    await clearPrivateCache();
    queryClient.clear();
    await storeSession(next);
  };
  const signOut = async () => {
    authGeneration++;
    await hydrate();
    const refreshToken = current?.refreshToken;
    await (await import("../native/geofences")).disconnectLocation();
    await (await import("../native/device")).removeDevice();
    await (await import("../native/googleCalendar")).clearGoogleCalendarOnSignOut();
    await storeSession(null);
    await clearPrivateCache();
    if (refreshToken && issuer && clientId) {
      try {
        const discovery = await AuthSession.fetchDiscoveryAsync(issuer);
        if (discovery.revocationEndpoint) await AuthSession.revokeAsync({ clientId, token: refreshToken, tokenTypeHint: AuthSession.TokenTypeHint.RefreshToken }, discovery);
      } catch { /* Local sign-out still succeeds when provider is offline. */ }
    }
  };

  return <AuthContext.Provider value={{ session, loading, configured: localAuth || Boolean(issuer && clientId), readyToSignIn: Boolean(issuer && clientId), signIn, signOut }}>{children}</AuthContext.Provider>;
}

export const useAuth = () => useContext(AuthContext);
