package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	llm *LLM
}

func validate(p *UserProfile) string {
	if p.Age < 14 || p.Age > 60 {
		return "중학교 ~ 정년 전까지의 나이만 추천해요."
	}

	if strings.TrimSpace(p.Goal) == "" {
		return "목표를 적어주세요."
	}

	if len(p.Goal) > 1000 || len(p.Education) > 1000 || len(p.Experience) > 1000 {
		return "1000자 이내로 작성해주세요."
	}

	if len(p.Skills) > 20 {
		return "강점(스킬)은 20개까지 작성 가능해요."
	}

	return ""
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) handleRecommend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST만 지원해요."})
		return
	}

	// 1MB 초과되는 크기의 body는 제한
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var p UserProfile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "요청 형식이 올바르지 않아요."})
		return
	}

	if msg := validate(&p); msg != "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	rec, err := s.llm.Recommend(ctx, p)
	if err != nil {
		log.Printf("추천 실패했어요 : %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "추천 생성에 실패했어요. 잠시 후 다시 시도해주세요."})
		return
	}

	writeJSON(w, http.StatusOK, rec)
}
