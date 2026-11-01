package main

import (
	"context"
	"fmt"

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
