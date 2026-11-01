package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: os.Getenv("API_KEY"),
	})
	if err != nil {
		panic(err)
	}
	llm := &LLM{client: client}

	// 사용자 입력 부분
	profile := UserProfile{
		Age:        25,
		Education:  "AI 로봇공학과",
		Experience: "LH 인턴, 한국남부발전 인턴, 한국주택금융공사 ICT운영실 인턴",
		Skills:     []string{"정보처리기사", "빅데이터분석기사", "파이썬"},
		Goal:       "1년 안에 공공기관 IT직렬로 입사 목표",
	}

	// 사용자 입력을 JSON 변환 후 <profile> 태그로 감싸기
	b, err := json.Marshal(profile)
	if err != nil {
		panic(err)
	}
	user := "<profile>\n" + string(b) + "\n</profile>"

	// LLM 모델 호출
	out, err := llm.Generate(ctx, systemPrompt, user)
	if err != nil {
		panic(err)
	}

	// 코드블록 파싱
	out = strings.TrimSpace(out)
	out = strings.TrimPrefix(out, "```json")
	out = strings.TrimPrefix(out, "```")
	out = strings.TrimPrefix(out, "```")

	var rec Recommendation
	if err := json.Unmarshal([]byte(out), &rec); err != nil {
		panic(err)
	}

	// 출력
	fmt.Println(rec.Summary)
	for _, c := range rec.Careers {
		fmt.Printf("\n[%s] 적합도 %d/5, 약 %d개월 \n%s\n", c.Role, c.Fit, c.Months, c.Reason)
		for i, s := range c.Steps {
			fmt.Printf("	%d, %s\n", i+1, s)
		}
	}
	fmt.Printf("\n", rec.Notice)
}
