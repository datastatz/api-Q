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
	// LAUNDRY INSTALLATION CHECK ROUTES
	// ========================================

	// POST /api/laundry/silver/v1/waterFeedAttachedToTap - Check water supply connection
	http.HandleFunc("/api/laundry/silver/v1/waterFeedAttachedToTap", func(w http.ResponseWriter, r *http.Request) {
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

		// System prompt specifiek voor water supply check
		systemPrompt := `You are a quality control expert for washing machine installations.

	Check if the water supply hose is properly connected to the tap/faucet.
	Look for: water inlet hose connected to water supply tap, secure connection, no leaks visible.

	RESPONSE FORMAT - FOLLOW EXACTLY:
	- Respond with ONLY "PASS" or "FAIL"
	- PASS: Water supply hose is properly connected to tap
	- FAIL: Water supply hose is not connected or connection is faulty
	- No explanations needed`

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

	// POST /api/laundry/silver/v1/drainHoseInDrain - Check drain hose connection
	http.HandleFunc("/api/laundry/silver/v1/drainHoseInDrain", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "ONLY POST REQUESTS ARE ALLOWED"})
			return
		}

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid form data"})
			return
		}

		file, header, err := r.FormFile("photo")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "No photo found"})
			return
		}
		defer file.Close()

		photoBytes, err := io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Could not read photo"})
			return
		}

		photoBase64 := base64.StdEncoding.EncodeToString(photoBytes)
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg"
		}

		systemPrompt := `You are a quality control expert for washing machine installations.

	Check if the drain hose is properly connected to the drain pipe.
	Look for: drain hose inserted into drain pipe, secure connection, proper positioning.

	RESPONSE FORMAT - FOLLOW EXACTLY:
	- Respond with ONLY "PASS" or "FAIL"
	- PASS: Drain hose is properly connected to drain pipe
	- FAIL: Drain hose is not connected or connection is faulty
	- No explanations needed`

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: "gpt-5-nano-2025-08-07",
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: systemPrompt,
					},
					{
						Role: openai.ChatMessageRoleUser,
						MultiContent: []openai.ChatMessagePart{
							{
								Type: openai.ChatMessagePartTypeText,
								Text: "Analyze this drain hose connection.",
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

		aiResponse := resp.Choices[0].Message.Content
		var result string
		if strings.TrimSpace(aiResponse) == "PASS" {
			result = "PASS"
		} else {
			result = "FAIL"
		}

		json.NewEncoder(w).Encode(QualityResponse{Result: result})
	})

	// POST /api/laundry/silver/v1/powerCordInSocket - Check power cord connection
	http.HandleFunc("/api/laundry/silver/v1/powerCordInSocket", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "ONLY POST REQUESTS ARE ALLOWED"})
			return
		}

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid form data"})
			return
		}

		file, header, err := r.FormFile("photo")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "No photo found"})
			return
		}
		defer file.Close()

		photoBytes, err := io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Could not read photo"})
			return
		}

		photoBase64 := base64.StdEncoding.EncodeToString(photoBytes)
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg"
		}

		systemPrompt := `You are a quality control expert for washing machine installations.

	Check if the power cord is properly plugged into the electrical socket.
	Look for: power cord plugged into wall socket, secure connection, no loose connections.

	RESPONSE FORMAT - FOLLOW EXACTLY:
	- Respond with ONLY "PASS" or "FAIL"
	- PASS: Power cord is properly plugged into socket
	- FAIL: Power cord is not plugged in or connection is faulty
	- No explanations needed`

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: "gpt-5-nano-2025-08-07",
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: systemPrompt,
					},
					{
						Role: openai.ChatMessageRoleUser,
						MultiContent: []openai.ChatMessagePart{
							{
								Type: openai.ChatMessagePartTypeText,
								Text: "Analyze this power cord connection.",
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

		aiResponse := resp.Choices[0].Message.Content
		var result string
		if strings.TrimSpace(aiResponse) == "PASS" {
			result = "PASS"
		} else {
			result = "FAIL"
		}

		json.NewEncoder(w).Encode(QualityResponse{Result: result})
	})

	// POST /api/laundry/silver/v1/rinseCycleMachineIsOn - Check if machine is running rinse cycle
	http.HandleFunc("/api/laundry/silver/v1/rinseCycleMachineIsOn", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "ONLY POST REQUESTS ARE ALLOWED"})
			return
		}

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid form data"})
			return
		}

		file, header, err := r.FormFile("photo")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "No photo found"})
			return
		}
		defer file.Close()

		photoBytes, err := io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Could not read photo"})
			return
		}

		photoBase64 := base64.StdEncoding.EncodeToString(photoBytes)
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg"
		}

		systemPrompt := `You are a quality control expert for washing machine installations.

	Check if the washing machine is running a rinse cycle (machine is on and operating).
	Look for: machine display showing active cycle, water movement, machine running, rinse cycle indicators.

	RESPONSE FORMAT - FOLLOW EXACTLY:
	- Respond with ONLY "PASS" or "FAIL"
	- PASS: Machine is running rinse cycle
	- FAIL: Machine is not running or not in rinse cycle
	- No explanations needed`

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: "gpt-5-nano-2025-08-07",
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: systemPrompt,
					},
					{
						Role: openai.ChatMessageRoleUser,
						MultiContent: []openai.ChatMessagePart{
							{
								Type: openai.ChatMessagePartTypeText,
								Text: "Analyze if the machine is running rinse cycle.",
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

		aiResponse := resp.Choices[0].Message.Content
		var result string
		if strings.TrimSpace(aiResponse) == "PASS" {
			result = "PASS"
		} else {
			result = "FAIL"
		}

		json.NewEncoder(w).Encode(QualityResponse{Result: result})
	})

	// POST /api/laundry/silver/v1/shippingBoltsRemoved - Check if shipping bolts are removed
	http.HandleFunc("/api/laundry/silver/v1/shippingBoltsRemoved", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "ONLY POST REQUESTS ARE ALLOWED"})
			return
		}

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid form data"})
			return
		}

		file, header, err := r.FormFile("photo")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "No photo found"})
			return
		}
		defer file.Close()

		photoBytes, err := io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Could not read photo"})
			return
		}

		photoBase64 := base64.StdEncoding.EncodeToString(photoBytes)
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg"
		}

		systemPrompt := `You are a quality control expert for washing machine installations.

	Check if the shipping bolts/transit bolts have been removed from the washing machine.
	Look for: no shipping bolts visible, bolt holes empty, machine properly positioned without transport locks.

	RESPONSE FORMAT - FOLLOW EXACTLY:
	- Respond with ONLY "PASS" or "FAIL"
	- PASS: Shipping bolts have been removed
	- FAIL: Shipping bolts are still present or visible
	- No explanations needed`

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: "gpt-5-nano-2025-08-07",
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: systemPrompt,
					},
					{
						Role: openai.ChatMessageRoleUser,
						MultiContent: []openai.ChatMessagePart{
							{
								Type: openai.ChatMessagePartTypeText,
								Text: "Analyze if shipping bolts have been removed.",
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

		aiResponse := resp.Choices[0].Message.Content
		var result string
		if strings.TrimSpace(aiResponse) == "PASS" {
			result = "PASS"
		} else {
			result = "FAIL"
		}

		json.NewEncoder(w).Encode(QualityResponse{Result: result})
	})

	// POST /api/laundry/silver/v1/levelIndicatorPresent - Check if spirit level is present
	http.HandleFunc("/api/laundry/silver/v1/levelIndicatorPresent", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "ONLY POST REQUESTS ARE ALLOWED"})
			return
		}

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid form data"})
			return
		}

		file, header, err := r.FormFile("photo")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "No photo found"})
			return
		}
		defer file.Close()

		photoBytes, err := io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Could not read photo"})
			return
		}

		photoBase64 := base64.StdEncoding.EncodeToString(photoBytes)
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg"
		}

		systemPrompt := `You are a quality control expert for washing machine installations.

	Check if a spirit level/level indicator is present on the washing machine.
	Look for: spirit level tool visible on or near the machine, level indicator present, measuring tool for leveling.

	RESPONSE FORMAT - FOLLOW EXACTLY:
	- Respond with ONLY "PASS" or "FAIL"
	- PASS: Spirit level/level indicator is present
	- FAIL: Spirit level/level indicator is not visible
	- No explanations needed`

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: "gpt-5-nano-2025-08-07",
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: systemPrompt,
					},
					{
						Role: openai.ChatMessageRoleUser,
						MultiContent: []openai.ChatMessagePart{
							{
								Type: openai.ChatMessagePartTypeText,
								Text: "Analyze if spirit level is present.",
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

		aiResponse := resp.Choices[0].Message.Content
		var result string
		if strings.TrimSpace(aiResponse) == "PASS" {
			result = "PASS"
		} else {
			result = "FAIL"
		}

		json.NewEncoder(w).Encode(QualityResponse{Result: result})
	})

	log.Println("Server start op :8080")
	http.ListenAndServe(":8080", nil)

}
