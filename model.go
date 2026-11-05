package main

// 사용자 입력 부분
type UserProfile struct {
	Age        int      `json:"age"`        // 나이
	Education  string   `json:"education"`  // 학력 (중졸, 고졸, 대졸 모두 포용한다.)
	Experience string   `json:"experience"` // 경력 (인턴 기간 포함하여 관련 업종에서 종사한 이력)
	Skills     []string `json:"skills"`     // 선택 항목
	Goal       string   `json:"goal"`       // 목표
}

// LLM 모델 응답
type Recommendation struct {
	Summary string       `json:"summary"`
	Careers []CareerPath `json:"careers"`
	Notice  string       `json:"notice"`
}

// 커리어 방향 출력
type CareerPath struct {
	Role       string   `json:"role"`       // 추천 직무
	Fit        int      `json:"fit"`        // 적합도 1~5
	Reason     string   `json:"reason"`     // 추천 사유
	Skills     []string `json:"skills"`     // 필요 역량
	TotalWeeks int      `json:"totalWeeks"` // 전체 기간(주)
	Roadmap    []Phase  `json:"roadmap"`    // 로드맵
	Risks      string   `json:"risks"`      // 위험도
}

// 로드맵 단계
type Phase struct {
	Title        string   `json:"title"`        // 로드맵 제목
	StartWeek    int      `json:"startWeek"`    // 시작 주차
	EndWeek      int      `json:"endWeek"`      // 종료 주차
	HoursPerWeek int      `json:"hoursPerWeek"` // 주당 학습 시간
	Tasks        []string `json:"tasks"`        // 로드맵 단계마다 할 일
	Deliverable  string   `json:"deliverable"`  // 결과물
}
