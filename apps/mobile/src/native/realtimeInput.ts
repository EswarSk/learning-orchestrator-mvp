type InputEvent = { type?: string; item_id?: string; transcript?: string };

export function completedLearnerTranscript(event: InputEvent, heardSpeech: Set<string>, listening: boolean): string | null {
  if (event.type === "input_audio_buffer.speech_started") {
    if (listening && event.item_id) heardSpeech.add(event.item_id);
    return null;
  }
  if (event.type !== "conversation.item.input_audio_transcription.completed" || !event.item_id || !heardSpeech.delete(event.item_id)) return null;
  return event.transcript?.trim() || null;
}

export async function finishAfterPendingTurns(pending: Set<Promise<boolean>>, completed: boolean, finish: () => Promise<void>) {
  const saved = await Promise.all([...pending]);
  if (completed && saved.includes(false)) throw new Error("A voice response was not assessed. Reconnect and try again before completing practice.");
  await finish();
}
