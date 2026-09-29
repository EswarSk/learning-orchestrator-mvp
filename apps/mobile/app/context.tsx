import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { router } from "expo-router";
import { AppState, Platform, Pressable, StyleSheet, Text, View } from "react-native";
import { api } from "../src/api/client";
import { disconnectLocation, enableCurrentPlace, locationStatus, selectedLocationTrack } from "../src/native/geofences";
import { disableNotifications, enableNotifications, notificationStatus } from "../src/native/push";
import { iosCalendarStatus, reconcileIOSCalendarPermission, requestIOSCalendarAccess, selectIOSCalendars, type IOSCalendarStatus } from "../src/native/iosCalendar";
import { connectGoogleCalendar, disconnectGoogleCalendar, googleCalendarStatus, reconcileGoogleCalendar, selectGoogleCalendars, type GoogleCalendarStatus } from "../src/native/googleCalendar";
import { Display, Kicker, SolidAction } from "../src/ui/Orbit";
import { Screen } from "../src/ui/Screen";
import { colors } from "../src/ui/theme";

const sources = [
  { name: "iPhone Calendar", purpose: "Choose calendars on this phone", uses: "Only calendars you select after iOS grants access.", limit: "Orbit will not create or edit events." },
  { name: "Google Calendar", purpose: "Keep learning in sync", uses: "Calendars you connect with Google.", limit: "No email access or event changes." },
  { name: "Location", purpose: "Fit practice to a place", uses: "Places you choose for a learning reminder.", limit: "No continuous location history in your learning profile." },
  { name: "Email", purpose: "Notice learning commitments", uses: "Only the messages you explicitly choose to share.", limit: "No automatic inbox-wide reading." },
  { name: "Messages", purpose: "Bring a real question into practice", uses: "A message you intentionally share with Orbit.", limit: "No background message access." },
  { name: "Notifications", purpose: "Offer one timely invitation", uses: "An eligible learning moment and your quiet-hour preference.", limit: "No automatic microphone or calendar action." },
] as const;

export default function Context() {
  const [open, setOpen] = useState<string | null>(null);
  const [locationEnabled, setLocationEnabled] = useState(false);
  const [selectedTrackId, setSelectedTrackId] = useState("");
  const tracks = useQuery({ queryKey: ["tracks"], queryFn: () => api<{ items: { id: string; subjectTitle?: string; subjectId: string; goal: string }[] }>("/v1/tracks") });
  const [push, setPush] = useState<{ available: boolean; enabled: boolean; systemAllowed: boolean } | null>(null);
  const [iosCalendar, setIOSCalendar] = useState<IOSCalendarStatus | null>(null);
  const [googleCalendar, setGoogleCalendar] = useState<GoogleCalendarStatus | null>(null);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  useEffect(() => {
    const refresh = () => {
      void locationStatus().then(setLocationEnabled).catch(() => {});
      void selectedLocationTrack().then(async saved => setSelectedTrackId(saved || await AsyncStorage.getItem("selectedTrackId") || "")).catch(() => {});
      void notificationStatus().then(setPush).catch(() => {});
      if (Platform.OS === "ios") void iosCalendarStatus().then(setIOSCalendar).catch(() => {});
      if (Platform.OS === "ios") void googleCalendarStatus().then(setGoogleCalendar).catch(error => setMessage(error instanceof Error ? error.message : "Couldn’t read Google calendars."));
    };
    refresh();
    const subscription = AppState.addEventListener("change", state => { if (state === "active") refresh(); });
    return () => subscription.remove();
  }, []);
  const changeLocation = async () => {
    setBusy(true); setMessage(null);
    try {
      if (locationEnabled) { await disconnectLocation(); setLocationEnabled(false); }
      else { await enableCurrentPlace(selectedTrackId); setLocationEnabled(true); }
    } catch (error) { setLocationEnabled(await locationStatus().catch(() => false)); setMessage(error instanceof Error ? error.message : "Couldn’t change location access. Please try again."); }
    finally { setBusy(false); }
  };
  const changeNotifications = async () => {
    setBusy(true); setMessage(null);
    try {
      if (push?.enabled) await disableNotifications();
      else await enableNotifications();
      setPush(await notificationStatus());
    } catch (error) { setMessage(error instanceof Error ? error.message : "Couldn’t change notification access. Please try again."); }
    finally { setBusy(false); }
  };
  const connectIOSCalendar = async () => {
    setBusy(true); setMessage(null);
    try { setIOSCalendar(await requestIOSCalendarAccess()); }
    catch (error) { setMessage(error instanceof Error ? error.message : "Couldn’t read iPhone calendars."); }
    finally { setBusy(false); }
  };
  const changeIOSCalendars = async (selected: string[]) => {
    if (!iosCalendar) return;
    setBusy(true); setMessage(null);
    try {
      setIOSCalendar(await selectIOSCalendars(selected));
    } catch (error) { setIOSCalendar(await iosCalendarStatus().catch(() => iosCalendar)); setMessage(error instanceof Error ? error.message : "Couldn’t save your calendar choice."); }
    finally { setBusy(false); }
  };
  const retryCalendar = async (source: "ios" | "google") => {
    setBusy(true); setMessage(null);
    try {
      if (source === "ios") { await reconcileIOSCalendarPermission(); setIOSCalendar(await iosCalendarStatus()); }
      else { await reconcileGoogleCalendar(); setGoogleCalendar(await googleCalendarStatus()); }
    } catch (error) { setMessage(error instanceof Error ? error.message : "Couldn’t sync calendars. Try again later."); }
    finally { setBusy(false); }
  };
  const changeGoogleConnection = async () => {
    setBusy(true); setMessage(null);
    try {
      if (googleCalendar?.connected) await disconnectGoogleCalendar();
      else await connectGoogleCalendar();
      setGoogleCalendar(await googleCalendarStatus());
    } catch (error) {
      setGoogleCalendar(await googleCalendarStatus().catch(() => googleCalendar));
      setMessage(error instanceof Error ? error.message : "Couldn’t change Google Calendar access.");
    } finally { setBusy(false); }
  };
  const toggleGoogleCalendar = async (id: string) => {
    if (!googleCalendar) return;
    setBusy(true); setMessage(null);
    try {
      const selected = googleCalendar.selectedIds.includes(id) ? googleCalendar.selectedIds.filter(value => value !== id) : [...googleCalendar.selectedIds, id];
      setGoogleCalendar(await selectGoogleCalendars(selected));
    } catch (error) {
      setGoogleCalendar(await googleCalendarStatus().catch(() => googleCalendar));
      setMessage(error instanceof Error ? error.message : "Couldn’t save your Google calendar choice.");
    } finally { setBusy(false); }
  };
  return <Screen>
    <View style={styles.header}><Pressable accessibilityRole="button" accessibilityLabel="Back to You" onPress={() => router.replace("/(tabs)/settings")} style={styles.back}><Text style={styles.backText}>←</Text></Pressable><Kicker>CONTEXT / OPTIONAL</Kicker></View>
    <View style={styles.intro}><Kicker>ONE SIGNAL AT A TIME</Kicker><Display>Let learning{"\n"}<Text style={styles.italic}>find its moment.</Text></Display><Text style={styles.copy}>When a source is available, Orbit will explain what it adds before requesting access. None is required to use a course.</Text></View>
    <View>
      <View style={styles.listHead}><Kicker>POSSIBLE SOURCES</Kicker><Kicker>OPTIONAL</Kicker></View>
      {sources.map((source, index) => {
        const connected = source.name === "iPhone Calendar" ? Boolean(iosCalendar?.permissionGranted && iosCalendar.selectedIds.length && !iosCalendar.syncPending)
          : source.name === "Google Calendar" ? Boolean(googleCalendar?.connected && googleCalendar.selectedIds.length && !googleCalendar.syncPending && !googleCalendar.loadError)
          : source.name === "Location" ? locationEnabled
          : source.name === "Notifications" ? Boolean(push?.available && push.enabled && push.systemAllowed)
          : false;
        const approved = source.name === "iPhone Calendar" ? Boolean(iosCalendar?.permissionGranted)
          : source.name === "Google Calendar" ? Boolean(googleCalendar?.connected)
          : connected;
        const syncPending = source.name === "iPhone Calendar" ? iosCalendar?.syncPending : source.name === "Google Calendar" ? googleCalendar?.syncPending : false;
        const checked = source.name === "Google Calendar" ? Boolean(googleCalendar?.connected && !googleCalendar.syncPending) : connected;
        const status = connected ? "Connected to context" : syncPending ? "Sync pending"
          : source.name === "Google Calendar" && googleCalendar?.loadError ? "Connection needs attention"
          : approved ? "Access approved · Choose calendars" : source.purpose;
        return <View key={source.name} style={styles.rowWrap}>
        <Pressable accessibilityRole="button" accessibilityLabel={`${source.name}, ${connected ? "connected to context" : syncPending ? "sync pending" : approved ? "access approved" : "not connected"}`} accessibilityState={{ expanded: open === source.name }} onPress={() => { setOpen(open === source.name ? null : source.name); setMessage(null); }} style={styles.row}>
          <Text style={styles.index}>{String(index + 1).padStart(2, "0")}</Text>
          <View style={styles.rowBody}><Text style={styles.rowTitle}>{source.name}</Text><Text style={styles.rowDetail}>{status}</Text></View>
          <Text style={styles.chevron}>{checked ? "✓" : open === source.name ? "−" : "+"}</Text>
        </Pressable>
        {open === source.name && <View style={styles.detail}>
          <Kicker>WHAT IT WOULD USE</Kicker><Text style={styles.copy}>{source.uses}</Text>
          <Kicker>YOUR BOUNDARY</Kicker><Text style={styles.copy}>{source.limit}</Text>
          {source.name === "iPhone Calendar" && Platform.OS === "ios" ? <>
            <Text style={styles.copy}>Orbit checks selected calendars when you open the app and shares only a free window near now. No event titles or guests leave your phone.</Text>
            {!iosCalendar?.permissionGranted ? <SolidAction label={busy ? "Opening…" : "Choose iPhone calendars"} disabled={busy} onPress={() => void connectIOSCalendar()}/> : <>
              {iosCalendar.calendars.length === 0 && <Text style={styles.unavailable}>No event calendars were found on this iPhone.</Text>}
              <View style={styles.trackChoices}>{iosCalendar.calendars.map(calendar => <Pressable key={calendar.id} accessibilityRole="checkbox" accessibilityState={{ checked: iosCalendar.selectedIds.includes(calendar.id), disabled: busy }} disabled={busy} onPress={() => void changeIOSCalendars(iosCalendar.selectedIds.includes(calendar.id) ? iosCalendar.selectedIds.filter(value => value !== calendar.id) : [...iosCalendar.selectedIds, calendar.id])} style={[styles.trackChoice, iosCalendar.selectedIds.includes(calendar.id) && styles.trackSelected]}><Text style={[styles.trackTitle, iosCalendar.selectedIds.includes(calendar.id) && styles.trackTitleSelected]}>{calendar.title}</Text><Text style={[styles.copy, iosCalendar.selectedIds.includes(calendar.id) && styles.trackTitleSelected]}>{calendar.account}</Text></Pressable>)}</View>
              <Text style={styles.copy}>{iosCalendar.syncPending ? "Change waiting to sync. Orbit will retry when you open the app." : iosCalendar.selectedIds.length ? `${iosCalendar.selectedIds.length} selected · availability syncs when the app opens` : "Tap a calendar above to connect it again."}</Text>
              {iosCalendar.syncPending && <SolidAction label={busy ? "Retrying…" : "Retry iPhone Calendar sync"} disabled={busy} onPress={() => void retryCalendar("ios")}/>}
              {(iosCalendar.selectedIds.length > 0 || iosCalendar.syncPending) && <SolidAction label={busy ? "Removing…" : "Disconnect iPhone Calendar"} disabled={busy} onPress={() => void changeIOSCalendars([])}/>}
            </>}
          </> : source.name === "Google Calendar" && Platform.OS === "ios" ? <>
            <Text style={styles.copy}>Connect a Google account separately from iPhone Calendar. Orbit reads only busy times from calendars you select while the app is open, then shares a near-term free window. No event titles or guests reach Orbit.</Text>
            {!googleCalendar?.available ? <Text style={styles.unavailable}>Google Calendar needs an iOS Google client ID in a new app build.</Text> : <>
              <SolidAction label={busy ? "Saving…" : googleCalendar.connected ? "Disconnect Google Calendar" : "Connect Google Calendar"} disabled={busy} onPress={() => void changeGoogleConnection()}/>
              {googleCalendar.connected && <>
                {googleCalendar.loadError && <Text style={styles.unavailable}>{googleCalendar.loadError}</Text>}
                {!googleCalendar.calendars.length && !googleCalendar.syncPending && !googleCalendar.loadError && <Text style={styles.unavailable}>No calendars were found in this Google account.</Text>}
                <View style={styles.trackChoices}>{googleCalendar.calendars.map(calendar => <Pressable key={calendar.id} accessibilityRole="checkbox" accessibilityState={{ checked: googleCalendar.selectedIds.includes(calendar.id), disabled: busy || googleCalendar.syncPending }} disabled={busy || googleCalendar.syncPending} onPress={() => void toggleGoogleCalendar(calendar.id)} style={[styles.trackChoice, googleCalendar.selectedIds.includes(calendar.id) && styles.trackSelected]}><Text style={[styles.trackTitle, googleCalendar.selectedIds.includes(calendar.id) && styles.trackTitleSelected]}>{calendar.title}</Text></Pressable>)}</View>
                <Text style={styles.copy}>{googleCalendar.loadError ? "No new availability is trusted until Google can be read again." : googleCalendar.syncPending ? "Change waiting to sync. Reopen the app to retry." : googleCalendar.selectedIds.length ? `${googleCalendar.selectedIds.length} selected · availability syncs when the app opens` : "Select a Google calendar to use its availability."}</Text>
                {(googleCalendar.syncPending || googleCalendar.loadError) && <SolidAction label={busy ? "Retrying…" : "Retry Google Calendar sync"} disabled={busy} onPress={() => void retryCalendar("google")}/>}
              </>}
            </>}
          </> : source.name === "Location" && Platform.OS === "ios" ? <>
            <Text style={styles.copy}>{locationEnabled ? "Watching one place for this learning path." : "Choose a learning path, then your current place. iOS will ask before access is enabled."}</Text>
            {!locationEnabled && <View style={styles.trackChoices}><Kicker>FOR THIS LEARNING PATH</Kicker>{tracks.data?.items.map(track => <Pressable key={track.id} accessibilityRole="radio" accessibilityState={{ selected: selectedTrackId === track.id }} onPress={() => setSelectedTrackId(track.id)} style={[styles.trackChoice, selectedTrackId === track.id && styles.trackSelected]}><Text style={[styles.trackTitle, selectedTrackId === track.id && styles.trackTitleSelected]}>{track.subjectTitle || track.subjectId}</Text><Text style={[styles.copy, selectedTrackId === track.id && styles.trackTitleSelected]} numberOfLines={1}>{track.goal}</Text></Pressable>)}{tracks.isError && <Text style={styles.unavailable}>Couldn’t load learning paths. Try reopening Connections.</Text>}{tracks.data?.items.length === 0 && <Text style={styles.unavailable}>Create a learning path before saving a place.</Text>}</View>}
            {locationEnabled && <Text style={styles.copy}>{tracks.data?.items.find(track => track.id === selectedTrackId)?.subjectTitle ?? "Selected learning path"}</Text>}
            <SolidAction label={busy ? "Saving…" : locationEnabled ? "Remove this place" : "Use this place"} disabled={busy || (!locationEnabled && !tracks.data?.items.some(track => track.id === selectedTrackId))} onPress={() => void changeLocation()}/>
          </> : source.name === "Notifications" && Platform.OS === "ios" && push?.available ? <>
            {push.enabled && !push.systemAllowed && <Text style={styles.copy}>iOS notifications are off. You can turn off invitations here or allow notifications in Settings.</Text>}
            <SolidAction label={busy ? "Saving…" : push.enabled ? "Turn off invitations" : "Allow invitations"} disabled={busy} onPress={() => void changeNotifications()}/>
          </> : <Text style={styles.unavailable}>Connection is not available in this build. This screen requests no access.</Text>}
          {message && <Text accessibilityLiveRegion="polite" style={styles.unavailable}>{message}</Text>}
        </View>}
      </View>; })}
    </View>
    <View style={styles.note}><Kicker>YOUR CHOICE STAYS YOURS</Kicker><Text style={styles.copy}>Courses work without connections. Each source asks separately before access.</Text></View>
    <SolidAction label="Continue without connections" onPress={() => router.replace("/learning-profile")}/>
  </Screen>;
}

const styles = StyleSheet.create({
  header: { minHeight: 48, flexDirection: "row", alignItems: "center", justifyContent: "space-between" }, back: { minWidth: 44, minHeight: 44, justifyContent: "center" }, backText: { color: colors.ink, fontSize: 27 }, intro: { marginTop: 19, gap: 14 }, italic: { fontStyle: "italic" }, copy: { color: colors.muted, fontSize: 13, lineHeight: 19 }, listHead: { flexDirection: "row", justifyContent: "space-between", marginBottom: 8 }, rowWrap: { borderTopWidth: 1, borderTopColor: colors.line }, row: { minHeight: 69, flexDirection: "row", alignItems: "center", gap: 12 }, index: { width: 24, color: colors.muted, fontSize: 11 }, rowBody: { flex: 1 }, rowTitle: { color: colors.ink, fontSize: 16, fontWeight: "600" }, rowDetail: { color: colors.muted, fontSize: 12, marginTop: 3 }, chevron: { color: colors.ink, fontSize: 22 }, detail: { backgroundColor: colors.soft, padding: 15, gap: 8, marginBottom: 10 }, unavailable: { color: colors.ink, fontSize: 11, lineHeight: 17, marginTop: 5 }, note: { borderTopWidth: 1, borderTopColor: colors.ink, paddingTop: 13, gap: 8 },
  trackChoices: { gap: 7, marginTop: 5 }, trackChoice: { padding: 11, borderWidth: 1, borderColor: colors.line, gap: 3 }, trackSelected: { backgroundColor: colors.ink, borderColor: colors.ink }, trackTitle: { color: colors.ink, fontSize: 14, fontWeight: "600" }, trackTitleSelected: { color: colors.background },
});
