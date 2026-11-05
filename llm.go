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

// 로드맵이 그릴 수 있는 형태인지 검증 (모델이 숫자를 틀리게 응답할 수 있고, 이에 따라 그림이 깨질 수 있기 때문)
func validateRoadmap(c *CareerPath) error {
	if len(c.Roadmap) == 0 {
		return fmt.Errorf("%q: 로드맵이 비어 있어요", c.Role)
	}

	prevEnd := 0
	for i, p := range c.Roadmap {
		if p.StartWeek != prevEnd+1 {
			return fmt.Errorf("%q: %d번째 단계 시작 주차가 %d여야 하는데 %d임", c.Role, i+1, prevEnd+1, p.StartWeek)
		}

		if p.EndWeek < p.StartWeek {
			return fmt.Errorf("%q: %d번째 종료 주차가 시작 주차보다 앞섬", c.Role, i+1)
		}

		if p.HoursPerWeek < 1 || p.HoursPerWeek > 60 {
			return fmt.Errorf("%q: %d번째 단계 주당 시간이 비정상(%d)", c.Role, i+1, p.HoursPerWeek)
		}

		prevEnd = p.EndWeek
	}

	if prevEnd > 104 {
		return fmt.Errorf("%q: 전체 기간이 2년을 넘음(%d주)", c.Role, prevEnd)
	}

	c.TotalWeeks = prevEnd

	return nil
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

	for i := range rec.Careers {
		if err := validateRoadmap(&rec.Careers[i]); err != nil {
			return nil, fmt.Errorf("로드맵 검증 실패: %w\n원문: %s", err, out)
		}
	}

	return &rec, nil
}
