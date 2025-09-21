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

	// Struct definiëren voor response van elke foto, dit helpt georganiseerde data terug te geven
	type PhotoAnalysis struct {
		Filename    string `json:"filename"`    // Naam van het bestand
		Description string `json:"description"` // Beschrijving van wat gecontroleerd moet worden
		Result      string `json:"result"`      // pass/fail/error/unknown
		Filesize    string `json:"filesize"`    // Grootte van het bestand
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
		err := r.ParseMultipartForm(10 << 20) // 10MB LIMIT VOOR ALLE FOTOS SAMEN
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

		// Initialiseer een array om alle foto analyses in op te slaan
		var analyses []PhotoAnalysis

		// Haal foto op uit form data
		file, header, err := r.FormFile("photo")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response := map[string]string{
				"error": "No photo found. Use 'photo' field to upload an image.",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		defer file.Close() // Zorg dat bestand gesloten wordt na gebruik

		// Haal bijbehorende beschrijving op
		description := r.FormValue("description")

		if description == "" {
			w.WriteHeader(http.StatusBadRequest)
			response := map[string]string{
				"error": "Missing description for photo",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		filename := header.Filename
		filesize := header.Size

		// Lees de foto inhoud naar memory
		photoBytes, err := io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response := map[string]string{
				"error": "Could not read photo file",
			}
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
			response := map[string]string{
				"error": "OpenAI API error: " + err.Error(),
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Parse OpenAI response
		aiResponse := resp.Choices[0].Message.Content

		// Initialiseer variabele voor resultaat
		var aiResult string

		// Check if response is "PASS"
		if strings.TrimSpace(aiResponse) == "PASS" {
			aiResult = "pass"

			// Check if response is "FAIL"
		} else if strings.TrimSpace(aiResponse) == "FAIL" {
			aiResult = "fail"

			// Onverwacht antwoord van de AI
		} else {
			aiResult = "unknown"
		}

		// Voeg analyse resultaat toe aan array
		analyses = append(analyses, PhotoAnalysis{
			Filename:    filename,
			Description: description,
			Result:      aiResult,
			Filesize:    fmt.Sprintf("%d bytes", filesize),
		})

		// Bouw finale response met alle analyses
		response := map[string]interface{}{
			"results": analyses, // Array van alle foto analyses
		}

		// Encodeer de response als JSON en stuur terug naar de client
		json.NewEncoder(w).Encode(response)
	})

	//Start een HTTP server op poort 8080, nil geef aan dat er nog geen routes zijn gedefinieerd
	log.Println("Server start op :8080")
	http.ListenAndServe(":8080", nil)

}
