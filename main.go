package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
	"google.golang.org/genai/interactions/models/interactions"
	"google.golang.org/genai/interactions/models/operations"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	key := os.Getenv("API_KEY")

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: key,
	})
	if err != nil {
		panic(err)
	}

	input := interactions.NewInteractionsInput("너가 누군지 간단하게 설명해")
	res, err := client.Interactions.Create(ctx, operations.CreateInteractionRequest{
		Body: operations.NewCreateInteractionRequestBody(interactions.CreateModelInteraction{
			Model: interactions.Model("gemini-3.5-flash"),
			Input: &input,
		}),
	})

	if err != nil {
		panic(err)
	}

	if res.Interaction.OutputText != nil {
		fmt.Println(*res.Interaction.OutputText)
	}
}
