package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	// Import database package
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
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

	// Eenvoudige response struct voor alleen result
	type QualityResponse struct {
		Result string `json:"result"` // PASS of FAIL
	}

	// ========================================
	// MAIN API ROUTE - Foto kwaliteitscontrole
	// ========================================

	// POST /quality-check - Analyseer foto's met AI
	// Body: multipart/form-data met foto1, description1, photo2, description2, etc.
	// Response: AI analyse resultaten (pass/fail) per foto
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
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid form data"})
			return
		}

		// Haal foto op uit form data
		file, header, err := r.FormFile("photo")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "No photo found"})
			return
		}
		defer file.Close()

		// Haal bijbehorende beschrijving op
		description := r.FormValue("description")

		if description == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing description"})
			return
		}

		// Lees de foto inhoud naar memory
		photoBytes, err := io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Could not read photo"})
			return
		}

		// Converteer de foto naar een base64 string
		photoBase64 := base64.StdEncoding.EncodeToString(photoBytes)

		// Bepaal het juiste MIME type van de foto (WEBP en avif nog toevoegen)
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg" // Default
		}

		// Bouw system prompt voor kwaliteitscontrole
		systemPrompt := fmt.Sprintf(`You are a quality control expert for installations.

	Requirements to analyze: %s

	RESPONSE FORMAT - FOLLOW EXACTLY:
	- Respond with ONLY "PASS" or "FAIL"
	- No explanations needed
	- Example: "PASS"
	- Example: "FAIL"`, description)

		// Maak OpenAI Vision request
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: "gpt-5-nano-2025-08-07",
				Messages: []openai.ChatCompletionMessage{
					// System prompt (instructies voor de AI)
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: systemPrompt,
					},
					{
						Role: openai.ChatMessageRoleUser,
						MultiContent: []openai.ChatMessagePart{
							{ // User message (gewone instructie)
								Type: openai.ChatMessagePartTypeText,
								Text: "Analyze this installation photo.",
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
			json.NewEncoder(w).Encode(map[string]string{"error": "AI analysis failed"})
			return
		}

		// Parse AI response
		aiResponse := resp.Choices[0].Message.Content
		var result string
		if strings.TrimSpace(aiResponse) == "PASS" {
			result = "PASS"
		} else {
			result = "FAIL" // Default voor alles wat niet PASS is
		}

		// Stuur response terug
		json.NewEncoder(w).Encode(QualityResponse{Result: result})
	})

	log.Println("Server start op :8080")
	http.ListenAndServe(":8080", nil)

}
