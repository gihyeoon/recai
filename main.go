package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

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

	s := &Server{llm: &LLM{client: client}}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/recommend", s.handleRecommend)
	mux.Handle("/", http.FileServer(http.Dir("static")))

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      90 * time.Second,
	}

	log.Println("서버 시작 : localhost:8080")
	log.Fatal(srv.ListenAndServe())
}
