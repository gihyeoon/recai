package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
	"google.golang.org/genai/interactions/models/interactions"
	"google.golang.org/genai/interactions/models/operations"
)

type LLM struct {
	client *genai.Client
}

func (l *LLM) Generate(ctx context.Context, system, user string) (string, error) {
	input := interactions.NewInteractionsInput(user)
	res, err := l.client.Interactions.Create(ctx, operations.CreateInteractionRequest{
		Body: operations.NewCreateInteractionRequestBody(interactions.CreateModelInteraction{
			Model:             interactions.Model("gemini-3.5-flash"),
			Input:             &input,
			SystemInstruction: &system,
		}),
	})

	if err != nil {
		return "", fmt.Errorf("LLM 호출 실패 %w", err)
	}

	if res == nil || res.Interaction == nil || res.Interaction.OutputText == nil {
		return "", fmt.Errorf("LLM이 빈 응답을 반환함")
	}

	return *res.Interaction.OutputText, nil
}

func (l *LLM) Recommend(ctx context.Context, profile UserProfile) (*Recommendation, error) {
	// 사용자 입력을 JSON 변환 후 <profile> 태그로 감싸기
	b, err := json.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("프로필 직렬화 실패 %w", err)
	}
	user := "<profile>\n" + string(b) + "\n</profile>"

	// LLM 모델 호출
	out, err := l.Generate(ctx, systemPrompt, user)
	if err != nil {
		return nil, err
	}

	// 코드블록 파싱
	out = strings.TrimSpace(out)
	out = strings.TrimPrefix(out, "```json")
	out = strings.TrimPrefix(out, "```")
	out = strings.TrimSuffix(out, "```")
	out = strings.TrimSpace(out)

	var rec Recommendation
	if err := json.Unmarshal([]byte(out), &rec); err != nil {
		return nil, fmt.Errorf("응답 JSON 파싱 실패 : %w\n원문: %s", err, out)
	}

	if len(rec.Careers) == 0 {
		return nil, fmt.Errorf("추천 결과가 비어있어요.\n원문: %s", out)
	}

	return &rec, nil
}
