package gemini

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// flow
// create client for handeling gemini related tasks
// need model for sending payload for getting a desired res

// @ types
type GeminiAIService struct {
	Client      *genai.Client
	AIModelName string
}

// func that returns the instance of type client of type "genai.client"
func NewGeminiAIService(geminiapikey string) (*GeminiAIService, error) {

	ctx := context.Background()

	// create geminiai client that holds all the operations for doing ai calls
	genAIClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: geminiapikey,
	})
	if err != nil {
		return nil, err
	}

	// returning client from the created method by serving the api key
	return &GeminiAIService{
		Client: genAIClient,
		// storing which model to use
		AIModelName: "gemini-2.5-flash",
	}, nil
}

// method that belongs to the GeminiAiservice type struct -> sends payload to gemini ai to process res
func (g *GeminiAIService) PromptGemini(prompt string) (AIRes string, AIResErr error) {

	// ctx for timeout based prompt sessions
	ctx := context.Background()

	// call GenCon method to send payload
	res, err := g.Client.Models.GenerateContent(ctx, g.AIModelName, genai.Text(prompt), nil)
	if err != nil {
		return "", err
	}

	// res validation - canditate is []*Candidates which are -> basically metadata of res in multi part
	// candidate holds all the data responded in multi-parts,.Content holds the responded content in multi-parts
	// each candidate {if valid} has parts slice which holds data of type *part in slice ofparts -> which holds the actual response
	// % both holds actual data in first element of each slice of respestive models
	// if len(res.Candidates) > 0 && res.Candidates[0].Content != nil {
	// 	//* if that exists <- res exists => send this response as part contain the texr
	// 	promptResponse := res.Candidates[0].Content.Parts[0].Text
	// 	// sending res as
	// 	return promptResponse, nil
	// }

	// otherwise send nothing but empty"" string
	promptResponse := res.Text() // this Text method from res does all the work automatically
	if promptResponse == "" {
		return "", fmt.Errorf("gemini returned an empty structured candidate")
	}

	// otherwise returned res
	return promptResponse, nil
}
