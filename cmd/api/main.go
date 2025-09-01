package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {

	//OpenAI client initialiseren
	// Laad .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Lees API key uit environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY not set")
	}

	client := openai.NewClient(apiKey)

	// Nieuwe route voor kwaliteitscontrole (alleen post requests)
	http.HandleFunc("/quality-check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Controleer of het een POST request is
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			response := map[string]string{
				"error": "ONLY POST REQUESTS ARE ALLOWED",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Parse multi-part form data (max 10MB)
		err := r.ParseMultipartForm(10 << 20) // 10MB LIMIT PER IMAGE
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			// Specifieke error messages
			var errorMessage string
			if err.Error() == "http: request body too large" {
				errorMessage = "File too large, maximum 10MB allowed"
			} else if err.Error() == "request Content-Type isn't multipart/form-data" {
				errorMessage = "Must send multipart/form-data, not regular JSON"
			} else {
				errorMessage = "Invalid form data: " + err.Error()
			}

			response := map[string]string{"error": errorMessage}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Haal de afbeelding op uit het formulier
		description := r.FormValue("description")

		file, header, err := r.FormFile("photo")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response := map[string]string{"error": "No photo found in request"}
			json.NewEncoder(w).Encode(response)
			return
		}
		defer file.Close()

		filename := header.Filename
		filesize := header.Size

		// Lees de foto inhoud naar memory
		photoBytes, err := io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := map[string]string{"error": "Could not read photo"}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Converteer de foto naar een base64 string
		photoBase64 := base64.StdEncoding.EncodeToString(photoBytes)

		// Bepaal het juiste MIME type van de foto
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			// Fallback gebaseerd op bestandsextensie
			if strings.HasSuffix(strings.ToLower(filename), ".png") {
				contentType = "image/png"
			} else if strings.HasSuffix(strings.ToLower(filename), ".webp") {
				contentType = "image/webp"
			} else {
				contentType = "image/jpeg" // Default
			}
		}

		// Maak OpenAI Vision request
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: "gpt-5-nano-2025-08-07",
				Messages: []openai.ChatCompletionMessage{
					{
						Role: openai.ChatMessageRoleUser,
						MultiContent: []openai.ChatMessagePart{
							{
								Type: openai.ChatMessagePartTypeText,
								Text: fmt.Sprintf("Analyze this installation photo. Requirements: %s\n\nRespond with exactly: 'PASS: [reason]' or 'FAIL: [reason]'", description),
							},
							{
								Type: openai.ChatMessagePartTypeImageURL,
								ImageURL: &openai.ChatMessageImageURL{
									URL: fmt.Sprintf("data:%s;base64,%s", contentType, photoBase64),
								},
							},
						},
					},
				},
			},
		)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := map[string]string{"error": "OpenAI API error: " + err.Error()}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Parse OpenAI response
		aiResponse := resp.Choices[0].Message.Content

		// Initialiseer variabelen voor resultaat en reden
		var aiResult, aiReason string

		// Check if response begint met "PASS"
		if len(aiResponse) > 5 && aiResponse[:4] == "PASS" {
			aiResult = "pass"
			aiReason = aiResponse[6:]

			// Check if response begint met "FAIL"
		} else if len(aiResponse) > 5 && aiResponse[:4] == "FAIL" {
			aiResult = "fail"
			aiReason = aiResponse[6:]

			// Onverwacht antwoord van de AI
		} else {
			aiResult = "unknown"
			aiReason = aiResponse
		}

		// Response met echte AI resultaten
		response := map[string]string{
			"result":      aiResult,
			"reason":      aiReason,
			"description": description,
			"filename":    filename,
			"filesize":    fmt.Sprintf("%d bytes", filesize),
		}

		// Encodeer de response als JSON en stuur terug naar de client

		json.NewEncoder(w).Encode(response)
	})

	//Start een HTTP server op poort 8080, nil geef aan dat er nog geen routes zijn gedefinieerd
	http.ListenAndServe(":8080", nil)
	log.Println("Server start op :8080")

}
