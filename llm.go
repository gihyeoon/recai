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
		panic(err)
	}

	if res.Interaction.OutputText != nil {
		fmt.Println(*res.Interaction.OutputText)
	}

	return *res.Interaction.OutputText, nil
}

func (l *LLM) Recommend(ctx context.Context, profile UserProfile) (*Recommendation, error) {
	// 사용자 입력을 JSON 변환 후 <profile> 태그로 감싸기
	b, err := json.Marshal(profile)
	if err != nil {
		panic(err)
	}
	user := "<profile>\n" + string(b) + "\n</profile>"

	// LLM 모델 호출
	out, err := l.Generate(ctx, systemPrompt, user)
	if err != nil {
		panic(err)
	}

	// 코드블록 파싱
	out = strings.TrimSpace(out)
	out = strings.TrimPrefix(out, "```json")
	out = strings.TrimPrefix(out, "```")
	out = strings.TrimSuffix(out, "```")
	out = strings.TrimSpace(out)

	var rec Recommendation
	if err := json.Unmarshal([]byte(out), &rec); err != nil {
		panic(err)
	}

	return &rec, nil
}
