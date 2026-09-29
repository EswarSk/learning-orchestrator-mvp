package learning

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// Run explicitly with synthetic content; this calls paid providers.
func TestLiveProviders(t *testing.T) {
	if os.Getenv("RUN_LIVE_AI_TEST") != "1" {
		t.Skip("set RUN_LIVE_AI_TEST=1 to call live providers")
	}
	a := AIFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	reply, err := a.reply(ctx, "Spanish", Node{Objective: "Introduce yourself in Spanish", Prompt: "Say your name in Spanish", AcceptedVariation: []string{"Me llamo Ana"}}, nil, "Me llamo Ana.", "none")
	if err != nil || strings.TrimSpace(reply.Tutor) == "" {
		t.Fatalf("live tutor reply failed: %v", err)
	}
	t.Logf("OpenAI reply valid; understandingMet=%t", reply.UnderstandingMet)
	audio, err := a.speech(ctx, "Hola.")
	if err != nil {
		t.Fatalf("live speech failed: %v", err)
	}
	transcript, err := a.transcribe(ctx, "probe.mp3", audio)
	if err != nil || !strings.Contains(strings.ToLower(transcript), "hola") {
		t.Fatalf("live voice round trip failed: %v", err)
	}
	t.Log("ElevenLabs speech and transcription round trip passed")
}
