package main

import (
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func allowedOrigins() map[string]bool {
	m := map[string]bool{}
	for _, o := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			m[o] = true
		}
	}
	return m
}

func withCORS(next http.Handler) http.Handler {
	allowed := allowedOrigins() // 서버 시작할 때 한 번만 읽어 둡니다.

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin") // 브라우저가 "나는 이 출처에서 왔다"라고 보내는 헤더

		// 허용 목록에 있는 출처일 때만 허용 헤더를 붙입니다.
		// 목록에 없으면 헤더를 안 붙이고, 브라우저가 알아서 응답을 차단합니다.
		if origin != "" && allowed[origin] {
			// 요청한 출처를 그대로 돌려줘야 브라우저가 "허락받았다"고 판단합니다.
			w.Header().Set("Access-Control-Allow-Origin", origin)
			// 허용할 HTTP 메서드. 이 API는 POST만 쓰고, OPTIONS는 사전 확인용입니다.
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			// 프론트가 보내는 헤더. JSON을 보낼 때 Content-Type을 쓰므로 필요합니다.
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			// 사전 확인 결과를 브라우저가 600초 동안 기억 → 매번 OPTIONS를 안 보냅니다.
			w.Header().Set("Access-Control-Max-Age", "600")
			// 출처마다 응답이 달라진다고 캐시(프록시/CDN)에 알립니다.
			// 이게 없으면 A 출처용 응답이 B 출처에 잘못 재사용될 수 있습니다.
			w.Header().Add("Vary", "Origin")
		}

		// [preflight] JSON POST처럼 "단순하지 않은" 요청은 브라우저가 본 요청 전에
		// OPTIONS 요청을 먼저 보내서 "이 요청 보내도 돼?"라고 물어봅니다.
		// 이 요청은 핸들러까지 보낼 필요가 없으니 여기서 204로 끝냅니다.
		// (handleRecommend는 POST 외에는 405를 돌려주기 때문에, 여기서 안 끝내면 preflight가 실패합니다.)
		// 또한 이 요청은 레이트 리밋에 세지 않도록 CORS를 바깥쪽에 둡니다.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type visitor struct {
	limiter  *rate.Limiter // 이 IP의 토큰 양동이
	lastSeen time.Time     // 마지막 요청 시각 (오래된 IP를 정리할 때 씁니다)
}

type IPLimiter struct {
	mu       sync.Mutex // 요청은 동시에 들어오므로, map을 건드릴 땐 잠금이 필수입니다. 없으면 크래시가 납니다.
	visitors map[string]*visitor
	r        rate.Limit // 토큰이 채워지는 속도
	burst    int        // 양동이 크기
}

func NewIPLimiter(r rate.Limit, burst int) *IPLimiter {
	l := &IPLimiter{visitors: map[string]*visitor{}, r: r, burst: burst}
	go l.cleanup() // 백그라운드에서 오래된 기록을 주기적으로 정리합니다.
	return l
}

// IP에 해당하는 양동이를 돌려줍니다. 처음 보는 IP면 새로 만듭니다.
func (l *IPLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	v, ok := l.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(l.r, l.burst)}
		l.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

// [왜 필요한가] IP 기록을 map에 계속 쌓기만 하면 방문자가 늘수록 메모리가 한없이 늘어납니다.
// 10분 넘게 안 온 IP는 지워서 막습니다. (지워도 다시 오면 새 양동이로 시작하니 문제없음)
func (l *IPLimiter) cleanup() {
	for range time.Tick(5 * time.Minute) {
		l.mu.Lock()
		for ip, v := range l.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(l.visitors, ip)
			}
		}
		l.mu.Unlock()
	}
}

// 요청한 사람의 IP를 꺼냅니다. RemoteAddr은 "1.2.3.4:51234" 형태라 포트를 떼어냅니다.
//
// [주의] 지금은 서버에 직접 접속하는 구조라 RemoteAddr이 맞습니다.
//
//	하지만 Cloudflare, Nginx, Cloud Run 같은 프록시 뒤에 배포하면
//	RemoteAddr이 항상 "프록시의 IP"가 되어, 모든 사용자가 한 사람으로 취급되고
//	그 한 명이 제한에 걸리면 전원이 같이 차단됩니다.
//	그때는 프록시가 넣어주는 X-Forwarded-For 같은 헤더를 읽어야 합니다.
//	단, 신뢰할 수 있는 프록시가 넣은 값만 써야 합니다.
//	(아무 요청의 헤더나 믿으면 사용자가 헤더를 조작해 제한을 우회합니다.)
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// 미들웨어: 토큰이 있으면 통과, 없으면 429(Too Many Requests)로 거절합니다.
func (l *IPLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.get(clientIP(r)).Allow() {
			// Retry-After: "이만큼 뒤에 다시 시도해"라고 클라이언트에 알려주는 표준 헤더입니다.
			w.Header().Set("Retry-After", "60")
			// 기존 에러 응답과 같은 {"error": "..."} 형태로 보내서,
			// 프론트가 별도 처리 없이 data.error를 그대로 화면에 보여줄 수 있습니다.
			writeJSON(w, http.StatusTooManyRequests,
				map[string]string{"error": "요청이 너무 많아요. 잠시 후 다시 시도해주세요."})
			return
		}
		next.ServeHTTP(w, r) // 통과
	})
}
