package translator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/artschekoff/video-transcriber/internal/whisper"
	openai "github.com/sashabaranov/go-openai"
)

func TestTranslate_PreservesTimestamps(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Content: "Привет"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	segs := []whisper.Segment{
		{Start: 0, End: 2 * time.Second, Text: "Hello"},
	}
	result, err := translateWithBaseURL(segs, "Russian", "sk-test", srv.URL+"/v1")
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if result[0].Start != 0 || result[0].End != 2*time.Second {
		t.Errorf("timestamps changed: %v", result[0])
	}
	if result[0].Text != "Привет" {
		t.Errorf("text: want 'Привет', got %q", result[0].Text)
	}
}

func TestTranslate_OneAPICallPerSegment(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Content: "x"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	segs := []whisper.Segment{
		{Text: "Hello"}, {Text: "World"},
	}
	translateWithBaseURL(segs, "Russian", "sk-test", srv.URL+"/v1")
	if calls != 2 {
		t.Errorf("want 2 API calls, got %d", calls)
	}
}
