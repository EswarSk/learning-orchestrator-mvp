import { useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { setAudioModeAsync } from "expo-audio";
import { Redirect, router, useLocalSearchParams } from "expo-router";
import { AppState, KeyboardAvoidingView, Platform, Pressable, ScrollView, StyleSheet, Text, TextInput, View } from "react-native";
import { mediaDevices, MediaStream as RTCMediaStream, RTCPeerConnection } from "react-native-webrtc";
import { SafeAreaView, useSafeAreaInsets } from "react-native-safe-area-context";
import { api } from "../src/api/client";
import { useAuth } from "../src/api/auth";
import { completedLearnerTranscript, finishAfterPendingTurns } from "../src/native/realtimeInput";
import { colors, serif } from "../src/ui/theme";

type Turn = { messageId: string; tutor: string; understandingMet: boolean; remainingSeconds: number };
type Message = { role: "learner" | "tutor"; text: string };

export default function Practice() {
  const insets = useSafeAreaInsets();
  const { session, loading } = useAuth();
  const queryClient = useQueryClient();
  const { id, prompt, hint, title, mode, understandingMet: initialUnderstanding } = useLocalSearchParams<{ id: string; prompt: string; hint: string; title: string; mode?: "text" | "voice"; understandingMet?: string }>();
  const [text, setText] = useState("");
  const [messages, setMessages] = useState<Message[]>(prompt ? [{ role: "tutor", text: prompt }] : []);
  const [understandingMet, setUnderstandingMet] = useState(initialUnderstanding === "true");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [connecting, setConnecting] = useState(false);
  const [connected, setConnected] = useState(false);
  const [assistantResponding, setAssistantResponding] = useState(false);
  const mounted = useRef(true);
  const peer = useRef<RTCPeerConnection | null>(null);
  const microphone = useRef<RTCMediaStream | null>(null);
  const channel = useRef<ReturnType<RTCPeerConnection["createDataChannel"]> | null>(null);
  const turnNumber = useRef(0);
  const heardSpeech = useRef(new Set<string>());
  const pendingTurns = useRef(new Set<Promise<boolean>>());
  const microphoneResumeTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const setAssistantAudioPlaying = (playing: boolean) => {
    if (microphoneResumeTimer.current) clearTimeout(microphoneResumeTimer.current);
    microphoneResumeTimer.current = null;
    if (mounted.current) setAssistantResponding(playing);
  };
  const resumeMicrophoneAfterPlayback = () => {
    if (microphoneResumeTimer.current) clearTimeout(microphoneResumeTimer.current);
    microphoneResumeTimer.current = setTimeout(() => setAssistantAudioPlaying(false), 700);
  };
  const stopRealtime = () => {
    if (microphoneResumeTimer.current) clearTimeout(microphoneResumeTimer.current);
    microphoneResumeTimer.current = null;
    channel.current?.close(); channel.current = null;
    microphone.current?.getTracks().forEach(track => track.stop()); microphone.current = null;
    peer.current?.close(); peer.current = null;
    heardSpeech.current.clear();
    if (mounted.current) setAssistantResponding(false);
    void setAudioModeAsync({ allowsRecording: false, playsInSilentMode: true, shouldRouteThroughEarpiece: false }).catch(() => {});
    if (mounted.current) setConnected(false);
  };
  const connectRealtime = async () => {
    if (!id || mode !== "voice" || connecting || connected) return;
    setConnecting(true); setError("");
    let pc: RTCPeerConnection | null = null;
    try {
      await setAudioModeAsync({ allowsRecording: true, playsInSilentMode: true, shouldRouteThroughEarpiece: false });
      const permission = await mediaDevices.getUserMedia({ audio: true, video: false });
      if (!mounted.current) { permission.getTracks().forEach(track => track.stop()); return; }
      microphone.current = permission;
      pc = new RTCPeerConnection(); peer.current = pc;
      permission.getAudioTracks().forEach(track => pc?.addTrack(track, permission));
      pc.onconnectionstatechange = () => {
        if (pc?.connectionState === "connected" && mounted.current) { setConnected(true); setConnecting(false); }
        if ((pc?.connectionState === "failed" || pc?.connectionState === "disconnected") && mounted.current) {
          stopRealtime(); setError("Voice connection was interrupted. Tap reconnect to try again."); setConnecting(false);
        }
      };
      const events = pc.createDataChannel("oai-events"); channel.current = events;
      events.onopen = () => events.send(JSON.stringify({ type: "response.create" }));
      events.onmessage = (event: { data: string }) => {
        if (channel.current !== events || !mounted.current) return;
        try {
          const data = JSON.parse(event.data as string) as { type?: string; item_id?: string; transcript?: string; delta?: string; response?: { status?: string } };
          const transcript = data.transcript?.trim();
          const learnerTranscript = completedLearnerTranscript(data, heardSpeech.current, Boolean(microphone.current?.getAudioTracks().some(track => track.enabled)));
          if (data.type === "response.created") setAssistantAudioPlaying(true);
          if (data.type === "output_audio_buffer.started") setAssistantAudioPlaying(true);
          if (data.type === "output_audio_buffer.stopped") resumeMicrophoneAfterPlayback();
          if (data.type === "output_audio_buffer.cleared") setAssistantAudioPlaying(false);
          if (data.type === "response.done" && data.response?.status !== "completed") setAssistantAudioPlaying(false);
          if (data.type === "error") setAssistantAudioPlaying(false);
          if (learnerTranscript) {
            setMessages(items => [...items, { role: "learner", text: learnerTranscript }]);
            const messageId = `live-${Date.now()}-${++turnNumber.current}`;
            const request = new AbortController();
            const timeout = setTimeout(() => request.abort(), 35_000);
            const saved = api<Turn>(`/v1/sessions/${id}/turns`, { method: "POST", signal: request.signal, idempotencyKey: `turn-${id}-${messageId}`, body: JSON.stringify({ messageId, text: learnerTranscript, assistanceRequested: "none" }) })
              .then(turn => { if (mounted.current && turn.understandingMet) setUnderstandingMet(true); return true; })
              .catch(cause => { if (mounted.current) setError(request.signal.aborted ? "Voice assessment timed out. Reconnect and try again." : cause instanceof Error ? cause.message : "Couldn’t save the practice turn."); return false; });
            pendingTurns.current.add(saved);
            void saved.finally(() => { clearTimeout(timeout); pendingTurns.current.delete(saved); });
          }
          if ((data.type === "response.audio_transcript.done" || data.type === "response.output_audio_transcript.done") && transcript) {
            setMessages(items => [...items, { role: "tutor", text: transcript }]);
          }
        } catch { /* Ignore non-JSON provider events. */ }
      };
      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);
      if (pc.iceGatheringState !== "complete") await new Promise<void>((resolve, reject) => {
        const timeout = setTimeout(() => { if (pc) pc.onicegatheringstatechange = null; reject(new Error("Timed out connecting to the voice service.")); }, 10000);
        const check = () => { if (pc?.iceGatheringState === "complete") { clearTimeout(timeout); if (pc) pc.onicegatheringstatechange = null; resolve(); } };
        if (pc) pc.onicegatheringstatechange = check;
        check();
      });
      const client = await api<{ clientSecret: string }>(`/v1/sessions/${id}/realtime`, { method: "POST" });
      if (!mounted.current) { stopRealtime(); return; }
      if (!pc.localDescription?.sdp) throw new Error("Could not create a voice connection.");
      const response = await fetch("https://api.openai.com/v1/realtime/calls", { method: "POST", headers: { Authorization: `Bearer ${client.clientSecret}`, "Content-Type": "application/sdp" }, body: pc.localDescription.sdp });
      if (!response.ok) throw new Error("OpenAI couldn’t start live voice. Check the server key and Realtime access.");
      await pc.setRemoteDescription({ type: "answer", sdp: await response.text() });
    } catch (cause) {
      stopRealtime();
      if (mounted.current) setError(cause instanceof Error ? cause.message : "Couldn’t start live voice.");
    } finally { if (mounted.current) setConnecting(false); }
  };
  useEffect(() => {
    mounted.current = true;
    if (mode === "voice" && id) void connectRealtime();
    const appState = AppState.addEventListener("change", state => { if (state !== "active") stopRealtime(); });
    return () => {
      appState.remove();
      mounted.current = false;
      stopRealtime();
    };
  }, [id, mode]);
  const send = async () => {
    if (!id || !text.trim() || busy) return;
    const mine = text.trim(); setBusy(true); setError(""); setText("");
    try {
      const messageId = `${Date.now()}`;
      const turn = await api<Turn>(`/v1/sessions/${id}/turns`, { method: "POST", idempotencyKey: `turn-${id}-${messageId}`, body: JSON.stringify({ messageId, text: mine, assistanceRequested: "none" }) });
      setMessages(items => [...items, { role: "learner", text: mine }, { role: "tutor", text: turn.tutor }]);
      if (turn.understandingMet) setUnderstandingMet(true);
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Couldn’t send that response."); setText(mine); }
    finally { setBusy(false); }
  };
  const finish = async (reason: "completed" | "userEnded") => {
    if (!id || busy) return;
    stopRealtime();
    setBusy(true); setError("");
    try {
      await finishAfterPendingTurns(pendingTurns.current, reason === "completed", () => api(`/v1/sessions/${id}/finish`, { method: "POST", idempotencyKey: `finish-${id}-${reason}`, body: JSON.stringify({ reason }) }));
      await Promise.all([queryClient.invalidateQueries({ queryKey: ["home"] }), queryClient.invalidateQueries({ queryKey: ["curriculum"] }), queryClient.invalidateQueries({ queryKey: ["progress"] }), queryClient.invalidateQueries({ queryKey: ["opportunities"] })]);
      router.back();
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Couldn’t close this lesson."); setBusy(false); }
  };
  if (loading) return null;
  if (!session) return <Redirect href="/sign-in"/>;
  const latest = messages.at(-1);
  return <SafeAreaView edges={[]} style={[styles.safe, { paddingTop: Math.max(insets.top, Platform.OS === "ios" ? 59 : 0), paddingBottom: insets.bottom }]}><KeyboardAvoidingView behavior={Platform.OS === "ios" ? "padding" : undefined} style={styles.flex}>
    <View style={styles.header}><Pressable accessibilityRole="button" accessibilityLabel="End practice" disabled={busy} onPress={() => void finish("userEnded")} style={styles.close}><Text style={styles.closeText}>×</Text></Pressable><Text style={styles.headerText}>ORBIT / GUIDED PRACTICE</Text><Text style={styles.headerText}>{mode === "voice" ? "VOICE" : "TEXT"}</Text></View>
    <View style={styles.progress}><View style={[styles.progressFill, { width: `${Math.min(90, 18 + messages.filter(item => item.role === "learner").length * 24)}%` }]}/></View>
    <ScrollView contentContainerStyle={styles.content} keyboardShouldPersistTaps="handled" showsVerticalScrollIndicator={false}>
      <View style={styles.orbit}><View style={styles.outerRing}/><View style={styles.innerRing}/><Text style={styles.orbitMark}>o.</Text></View>
      <Text style={styles.kicker}>{title?.toUpperCase() ?? "YOUR LEARNING"}</Text><Text style={styles.prompt}>{latest?.role === "tutor" ? latest.text : prompt}</Text>
      <View style={styles.status}><View style={styles.statusDot}/><View><Text style={styles.statusMain}>{busy ? "Orbit is thinking" : connecting ? "Connecting voice…" : mode === "voice" && assistantResponding ? "Orbit is responding · you can speak" : mode === "voice" && connected ? "Listening · speak naturally" : understandingMet ? "Understanding check met" : "Your turn"}</Text><Text style={styles.statusSub}>{understandingMet ? "Ready for the next step · real-world practice is tracked separately" : mode === "voice" ? "Live voice · speak naturally" : "Text practice · microphone off"}</Text></View></View>
      {mode === "voice" && !connected && <Pressable accessibilityRole="button" disabled={connecting || busy} onPress={() => void connectRealtime()} style={styles.voiceControl}><Text style={styles.voiceControlText}>{connecting ? "Connecting…" : "Reconnect voice"}</Text></Pressable>}
      {messages.length > 1 && <View style={styles.transcript}><Text style={styles.transcriptHead}>CONVERSATION SO FAR</Text>{messages.slice(1).map((message, index) => <View key={`${message.role}-${index}`} style={styles.transcriptRow}><Text style={styles.transcriptRole}>{message.role === "learner" ? "YOU" : "ORBIT"}</Text><Text style={styles.transcriptText}>{message.text}</Text></View>)}</View>}
      {messages.length === 1 && hint ? <View style={styles.hint}><Text style={styles.hintLabel}>A POSSIBLE START</Text><Text style={styles.hintText}>{hint}</Text></View> : null}
      {error ? <Text accessibilityLiveRegion="polite" style={styles.error}>{error}</Text> : null}
    </ScrollView>
    {mode !== "voice" && <View style={styles.composer}><TextInput accessibilityLabel="Your response" placeholder="Write your response…" placeholderTextColor="#AAA9A4" value={text} onChangeText={setText} multiline style={styles.input}/><Pressable accessibilityRole="button" accessibilityLabel="Send response" disabled={!text.trim() || busy} onPress={() => void send()} style={[styles.send, (!text.trim() || busy) && styles.disabled]}><Text style={styles.sendText}>→</Text></Pressable></View>}
    {understandingMet && <View style={styles.composer}><Pressable accessibilityRole="button" onPress={() => void finish("completed")} style={styles.finish}><Text style={styles.finishText}>Continue learning</Text><Text style={styles.finishText}>✓</Text></Pressable></View>}
  </KeyboardAvoidingView></SafeAreaView>;
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.ink }, flex: { flex: 1 }, header: { minHeight: 60, paddingHorizontal: 25, flexDirection: "row", alignItems: "center", justifyContent: "space-between", gap: 8 }, close: { minWidth: 44, minHeight: 44, justifyContent: "center" }, closeText: { fontSize: 29, color: colors.background }, headerText: { color: colors.background, opacity: .66, fontSize: 10, letterSpacing: 1.1 }, progress: { height: 2, backgroundColor: "#555450", marginHorizontal: 25 }, progressFill: { height: 2, backgroundColor: colors.background },
  content: { paddingHorizontal: 25, paddingBottom: 30 }, orbit: { height: 185, alignItems: "center", justifyContent: "center", marginTop: 16, marginBottom: 9 }, outerRing: { position: "absolute", width: 168, height: 168, borderRadius: 84, borderWidth: 1, borderColor: "#777671" }, innerRing: { position: "absolute", width: 111, height: 111, borderRadius: 56, borderWidth: 1, borderColor: "#AAA9A4" }, orbitMark: { fontFamily: serif, fontSize: 73, letterSpacing: -8, color: colors.background, marginTop: -10 },
  kicker: { color: colors.background, opacity: .65, fontSize: 11, letterSpacing: 1.2 }, prompt: { color: colors.background, fontFamily: serif, fontSize: 29, lineHeight: 34, marginTop: 15 }, status: { borderTopWidth: 1, borderTopColor: "#777671", marginTop: 25, paddingTop: 13, flexDirection: "row", alignItems: "center", gap: 10 }, statusDot: { width: 8, height: 8, borderRadius: 4, backgroundColor: colors.background }, statusMain: { color: colors.background, fontSize: 12 }, statusSub: { color: colors.background, opacity: .62, fontSize: 11, marginTop: 3 },
  transcript: { marginTop: 27, borderTopWidth: 1, borderTopColor: "#777671" }, transcriptHead: { color: colors.background, opacity: .65, fontSize: 10, letterSpacing: 1.1, paddingVertical: 12 }, transcriptRow: { borderTopWidth: 1, borderTopColor: "#555450", paddingVertical: 11, gap: 5 }, transcriptRole: { color: colors.background, opacity: .65, fontSize: 10, letterSpacing: 1 }, transcriptText: { color: colors.background, fontSize: 14, lineHeight: 20 }, hint: { marginTop: 24, padding: 14, backgroundColor: "#30302E", gap: 8 }, hintLabel: { color: colors.background, opacity: .65, fontSize: 10, letterSpacing: 1 }, hintText: { color: colors.background, fontSize: 13, lineHeight: 19 }, error: { color: "#F0B8B8", marginTop: 13, fontSize: 12 },
  composer: { paddingHorizontal: 25, paddingTop: 13, paddingBottom: 8, borderTopWidth: 1, borderTopColor: "#66655F", flexDirection: "row", flexWrap: "wrap", gap: 8 }, input: { flex: 1, minHeight: 53, maxHeight: 100, borderWidth: 1, borderColor: "#777671", padding: 12, color: colors.background, fontSize: 16, textAlignVertical: "top" }, send: { width: 53, height: 53, backgroundColor: colors.background, alignItems: "center", justifyContent: "center" }, sendText: { fontSize: 23, color: colors.ink }, disabled: { opacity: .45 }, finish: { width: "100%", minHeight: 45, borderWidth: 1, borderColor: colors.background, paddingHorizontal: 13, flexDirection: "row", justifyContent: "space-between", alignItems: "center" }, finishText: { color: colors.background, fontSize: 12 }, voiceControl: { borderWidth: 1, borderColor: "#777671", minHeight: 44, alignItems: "center", justifyContent: "center" }, voiceControlText: { color: colors.background, fontSize: 12 },
});
