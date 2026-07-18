package translator

import (
	"context"
	"fmt"
	"strings"

	"github.com/artschekoff/video-transcriber/internal/whisper"
	openai "github.com/sashabaranov/go-openai"
)

func translateWithBaseURL(
	segments []whisper.Segment,
	targetLanguage, apiKey, baseURL string,
) ([]whisper.Segment, error) {
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = baseURL
	client := openai.NewClientWithConfig(cfg)

	result := make([]whisper.Segment, len(segments))
	for i, seg := range segments {
		resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: fmt.Sprintf(
						"Translate the following subtitle text to %s. Return only the translated text.",
						targetLanguage,
					),
				},
				{Role: openai.ChatMessageRoleUser, Content: strings.TrimSpace(seg.Text)},
			},
			Temperature: 0.2,
		})
		if err != nil {
			return nil, fmt.Errorf("translate segment %d: %w", i, err)
		}
		result[i] = whisper.Segment{
			Start: seg.Start,
			End:   seg.End,
			Text:  resp.Choices[0].Message.Content,
		}
	}
	return result, nil
}

// Translate returns a new []Segment with translated Text; Start/End preserved.
func Translate(segments []whisper.Segment, targetLanguage, apiKey string) ([]whisper.Segment, error) {
	return translateWithBaseURL(segments, targetLanguage, apiKey, "https://api.openai.com/v1")
}
