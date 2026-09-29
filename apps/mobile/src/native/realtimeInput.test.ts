import { expect, test } from "vitest";
import { completedLearnerTranscript, finishAfterPendingTurns } from "./realtimeInput";

test("a learner transcript arriving during a coach response is assessed once; muted speech is ignored", () => {
  const heard = new Set<string>();
  const started = (item_id: string) => ({ type: "input_audio_buffer.speech_started", item_id });
  const completed = (item_id: string) => ({ type: "conversation.item.input_audio_transcription.completed", item_id, transcript: "  Hola  " });

  expect(completedLearnerTranscript(started("learner"), heard, true)).toBeNull();
  expect(completedLearnerTranscript(completed("learner"), heard, false)).toBe("Hola");
  expect(completedLearnerTranscript(completed("learner"), heard, true)).toBeNull();
  expect(completedLearnerTranscript(started("coach-echo"), heard, false)).toBeNull();
  expect(completedLearnerTranscript(completed("coach-echo"), heard, true)).toBeNull();
  expect(completedLearnerTranscript(completed("unheard"), heard, true)).toBeNull();
});

test("finishing practice waits and blocks completion when a pending voice assessment fails", async () => {
  let settle!: (saved: boolean) => void;
  const pending = new Set([new Promise<boolean>(resolve => { settle = resolve; })]);
  let finished = false;
  const finish = async () => { finished = true; };
  const attempt = finishAfterPendingTurns(pending, true, finish);
  await Promise.resolve();
  expect(finished).toBe(false);
  settle(false);
  await expect(attempt).rejects.toThrow("not assessed");
  expect(finished).toBe(false);
  await finishAfterPendingTurns(new Set([Promise.resolve(true)]), true, finish);
  expect(finished).toBe(true);
  finished = false;
  await finishAfterPendingTurns(new Set([Promise.resolve(false)]), false, finish);
  expect(finished).toBe(true);
});
