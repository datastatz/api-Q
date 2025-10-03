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
	"regexp"
	"strings"

	// Import database package
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

// Validatie functie voor projectNumber (letters, cijfers, underscore, hyphen)
func isValidProjectNumber(projectNumber string) bool {
	if projectNumber == "" || len(projectNumber) > 50 {
		return false
	}
	// Regex: alleen letters, cijfers, underscore en hyphen
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, projectNumber)
	return matched
}

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

	// Eenvoudige response struct voor alleen result (Silver tier)
	type QualityResponse struct {
		Result string `json:"result"` // PASS of FAIL
	}

	// Uitgebreide response struct voor Gold tier
	type GoldResponse struct {
		Result        string `json:"result"`        // PASS of FAIL
		ProjectNumber string `json:"projectNumber"` // Project identifier
		Reason        string `json:"reason"`        // Uitleg waarom PASS/FAIL
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
		systemPrompt := `You are a quality control expert for home appliance water connections.

		PHOTO QUALITY CHECK FIRST:
		- First check if the photo is clear enough for proper analysis
		- If the image is too blurry, unclear, or has poor quality that prevents proper evaluation, respond with: FAIL
		- Only proceed with the main check if photo quality is acceptable

		Evaluate if the water supply system is properly connected and functional.

		WHAT TO LOOK FOR:
		- Water inlet hose(s) present and connected (may include gray/silver flexible hoses)
		- Connection to water supply point (tap, valve, or wall outlet)
		- Leak detection device (aquastop) if present - should be connected
		- No visible water leaks or loose connections
		- Hoses are not kinked or damaged

		RESPONSE FORMAT - FOLLOW EXACTLY:
		- Respond with ONLY "PASS" or "FAIL"
		- PASS: Water supply system is properly connected AND photo quality is good
		- FAIL: Missing connection, visible leaks, damaged components OR photo is too blurry for analysis
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

		systemPrompt := `You are a quality control expert for appliance installations.

			PHOTO QUALITY CHECK FIRST:
			- First check if the photo is clear enough for proper analysis
			- If the image is too blurry, unclear, or has poor quality that prevents proper evaluation, respond with: FAIL
			- Only proceed with the main check if photo quality is acceptable

			CHECK: Is the drain hose connected to drainage?

			DRAIN HOSE: Large ribbed gray/blue corrugated hose (NOT the smooth water supply hose)

			PASS CONDITIONS:
			- Drain hose goes downward toward floor/wall
			- Hose appears to enter a drain, pipe, or opening
			- Hose is positioned for proper drainage (even if full connection not visible)

			FAIL CONDITIONS ONLY:
			- Drain hose is completely loose and hanging in the air
			- Hose is lying flat on the floor disconnected
			- No drain hose visible at all in the image
			- Only water supply hose visible (smooth, not ribbed)
			- Photo is too blurry for proper analysis

			IMPORTANT: If the drain hose goes downward and appears connected to drainage (even if you cannot see the exact connection point), respond PASS.

			Respond with ONLY "PASS" or "FAIL"`

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

		systemPrompt := `You are a quality control expert for appliance installations.

		PHOTO QUALITY CHECK FIRST:
		- First check if the photo is clear enough for proper analysis
		- If the image is too blurry, unclear, or has poor quality that prevents proper evaluation, respond with: FAIL
		- Only proceed with the main check if photo quality is acceptable

		CHECK: Is the power plug connected to an electrical outlet?

		WHAT TO LOOK FOR:
		- A power plug inserted into any type of electrical socket/outlet
		- This can be: wall socket, power strip, junction box, or any electrical connection point
		- The plug should be inserted (even if partially visible or in corner of image)

		PASS = Power plug is connected to ANY electrical outlet (wall, strip, box, etc.) AND photo quality is good
		FAIL = Plug clearly not connected, hanging loose, no electrical connection visible OR photo is too blurry for analysis

		Even if the connection is small or in corner of image, if you can see a plug connected to power, respond PASS.

		Respond with ONLY "PASS" or "FAIL"`

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

		systemPrompt := `You are a quality control expert for appliance installations.

		PHOTO QUALITY CHECK FIRST:
		- First check if the photo is clear enough for proper analysis
		- If the image is too blurry, unclear, or has poor quality that prevents proper evaluation, respond with: FAIL
		- Only proceed with the main check if photo quality is acceptable

		CHECK: Is the machine is powered on?
		
		PASS = Machine display is active/lit up showing time or cycle information AND photo quality is good
		FAIL = Display is off/dark, no machine visible OR photo is too blurry for analysis
		
		Respond with ONLY "PASS" or "FAIL"`

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

		systemPrompt := `You are a quality control expert for home appliance installations (washing machines, dryers, dishwashers, etc.).

	PHOTO QUALITY CHECK FIRST:
	- First check if the photo is clear enough for proper analysis
	- If the image is too blurry, unclear, or has poor quality that prevents proper evaluation, respond with: FAIL
	- Only proceed with the main check if photo quality is acceptable

	Check if the shipping bolts/transit bolts have been removed from the appliance.

	PASS CONDITIONS:
	- Shipping bolts have been removed from their original mounting positions in the appliance
	- If shipping bolts are visible on top of the machine or next to it, this means they were successfully REMOVED and should be counted as PASS
	- Bolt holes in the appliance are empty (no bolts screwed into the appliance itself)
	- Appliance is properly positioned without transport locks

	FAIL CONDITIONS:
	- Shipping bolts are still screwed into the appliance in their original positions
	- Appliance is still locked in transport position with bolts in place

	RESPONSE FORMAT - FOLLOW EXACTLY:
	- Respond with ONLY "PASS" or "FAIL"
	- PASS: Shipping bolts have been removed from the appliance (even if visible on top/side) AND photo quality is good
	- FAIL: Shipping bolts are still installed in the appliance OR photo is too blurry for analysis
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

		systemPrompt := `You are a quality control expert for home appliance installations (washing machines, dryers, dishwashers, etc.).

	PHOTO QUALITY CHECK FIRST:
	- First check if the photo is clear enough for proper analysis
	- If the image is too blurry, unclear, or has poor quality that prevents proper evaluation, respond with: FAIL
	- Only proceed with the main check if photo quality is acceptable

	Check if a spirit level/level indicator is present on the appliance.
	Look for: spirit level tool visible on or near the appliance, level indicator present, measuring tool for leveling.

	RESPONSE FORMAT - FOLLOW EXACTLY:
	- Respond with ONLY "PASS" or "FAIL"
	- PASS: Spirit level/level indicator is present AND photo quality is good
	- FAIL: Spirit level/level indicator is not visible OR photo is too blurry for analysis
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

	// ========================================
	// GOLD TIER ROUTES (met projectNumber en reasoning)
	// ========================================

	// POST /api/laundry/gold/v1/{projectNumber}/waterFeedAttachedToTap - Check water supply connection
	http.HandleFunc("/api/laundry/gold/v1/", func(w http.ResponseWriter, r *http.Request) {
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

		// Parse URL path om projectNumber en endpoint te extraheren
		path := strings.TrimPrefix(r.URL.Path, "/api/laundry/gold/v1/")
		pathParts := strings.Split(path, "/")

		if len(pathParts) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid URL format. Expected: /api/laundry/gold/v1/{projectNumber}/{endpoint}"})
			return
		}

		projectNumber := pathParts[0]
		endpoint := pathParts[1]

		// Valideer projectNumber
		if !isValidProjectNumber(projectNumber) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid projectNumber. Only letters, numbers, underscores and hyphens allowed (max 50 chars)"})
			return
		}

		// Controleer of endpoint geldig is
		validEndpoints := map[string]bool{
			"waterFeedAttachedToTap": true,
			"drainHoseInDrain":       true,
			"powerCordInSocket":      true,
			"rinseCycleMachineIsOn":  true,
			"shippingBoltsRemoved":   true,
			"levelIndicatorPresent":  true,
		}

		if !validEndpoints[endpoint] {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid endpoint. Valid endpoints: waterFeedAttachedToTap, drainHoseInDrain, powerCordInSocket, rinseCycleMachineIsOn, shippingBoltsRemoved, levelIndicatorPresent"})
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

		// Bepaal het juiste MIME type van de foto
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg" // Default
		}

		// Bepaal system prompt op basis van endpoint
		var systemPrompt string
		switch endpoint {
		case "waterFeedAttachedToTap":
			systemPrompt = `You are a quality control expert for home appliance water connections.

PHOTO QUALITY CHECK:
- First check if the photo is clear enough for proper analysis
- If the image is too blurry, unclear, or has poor quality that prevents proper evaluation, respond with:
FAIL
Photo too blurry - please retake with better focus

- Only proceed with the main check if photo quality is acceptable

Evaluate if the water supply system is properly connected and functional.

WHAT TO LOOK FOR:
- Water inlet hose(s) present and connected (may include gray/silver flexible hoses)
- Connection to water supply point (tap, valve, or wall outlet)
- Leak detection device (aquastop) if present - should be connected
- No visible water leaks or loose connections
- Hoses are not kinked or damaged

RESPONSE FORMAT - FOLLOW EXACTLY:
- First line: "PASS" or "FAIL"
- Second line: Brief explanation (max 100 characters) why it passed or failed
- Example:
PASS
Water supply properly connected with no visible leaks

or

FAIL
No water connection visible or loose hoses detected`

		case "drainHoseInDrain":
			systemPrompt = `You are a quality control expert for appliance installations.

PHOTO QUALITY CHECK:
- First check if the photo is clear enough for proper analysis
- If the image is too blurry, unclear, or has poor quality that prevents proper evaluation, respond with:
FAIL
Photo too blurry - please retake with better focus

- Only proceed with the main check if photo quality is acceptable

CHECK: Is the drain hose connected to drainage?

DRAIN HOSE: Large ribbed gray/blue corrugated hose (NOT the smooth water supply hose)

PASS CONDITIONS:
- Drain hose goes downward toward floor/wall
- Hose appears to enter a drain, pipe, or opening
- Hose is positioned for proper drainage (even if full connection not visible)

FAIL CONDITIONS ONLY:
- Drain hose is completely loose and hanging in the air
- Hose is lying flat on the floor disconnected
- No drain hose visible at all in the image
- Only water supply hose visible (smooth or ribbed)

RESPONSE FORMAT - FOLLOW EXACTLY:
- First line: "PASS" or "FAIL"
- Second line: Brief explanation (max 100 characters) why it passed or failed`

		case "powerCordInSocket":
			systemPrompt = `You are a quality control expert for appliance installations.

PHOTO QUALITY CHECK:
- First check if the photo is clear enough for proper analysis
- If the image is too blurry, unclear, or has poor quality that prevents proper evaluation, respond with:
FAIL
Photo too blurry - please retake with better focus

- Only proceed with the main check if photo quality is acceptable

CHECK: Is the power plug connected to an electrical outlet?

WHAT TO LOOK FOR:
- A power plug inserted into any type of electrical socket/outlet
- This can be: wall socket, power strip, junction box, or any electrical connection point
- The plug should be inserted (even if partially visible or in corner of image)

PASS = Power plug is connected to ANY electrical outlet (wall, strip, box, etc.)
FAIL = Plug clearly not connected, hanging loose, or no electrical connection visible

RESPONSE FORMAT - FOLLOW EXACTLY:
- First line: "PASS" or "FAIL"
- Second line: Brief explanation (max 100 characters) why it passed or failed`

		case "rinseCycleMachineIsOn":
			systemPrompt = `You are a quality control expert for appliance installations.

PHOTO QUALITY CHECK:
- First check if the photo is clear enough for proper analysis
- If the image is too blurry, unclear, or has poor quality that prevents proper evaluation, respond with:
FAIL
Photo too blurry - please retake with better focus

- Only proceed with the main check if photo quality is acceptable

CHECK: Is the machine powered on?

PASS = Machine display is active/lit up showing time or cycle information
FAIL = Display is off/dark, or no machine visible

RESPONSE FORMAT - FOLLOW EXACTLY:
- First line: "PASS" or "FAIL"
- Second line: Brief explanation (max 100 characters) why it passed or failed`

		case "shippingBoltsRemoved":
			systemPrompt = `You are a quality control expert for home appliance installations (washing machines, dryers, dishwashers, etc.).

PHOTO QUALITY CHECK:
- First check if the photo is clear enough for proper analysis
- If the image is too blurry, unclear, or has poor quality that prevents proper evaluation, respond with:
FAIL
Photo too blurry - please retake with better focus

- Only proceed with the main check if photo quality is acceptable

Check if the shipping bolts/transit bolts have been removed from the appliance.

PASS CONDITIONS:
- Shipping bolts have been removed from their original mounting positions in the appliance
- If shipping bolts are visible on top of the machine or next to it, this means they were successfully REMOVED and should be counted as PASS
- Bolt holes in the appliance are empty (no bolts screwed into the appliance itself)
- Appliance is properly positioned without transport locks

FAIL CONDITIONS:
- Shipping bolts are still screwed into the appliance in their original positions
- Appliance is still locked in transport position with bolts in place

RESPONSE FORMAT - FOLLOW EXACTLY:
- First line: "PASS" or "FAIL"
- Second line: Brief explanation (max 100 characters) why it passed or failed`

		case "levelIndicatorPresent":
			systemPrompt = `You are a quality control expert for home appliance installations (washing machines, dryers, dishwashers, etc.).

PHOTO QUALITY CHECK:
- First check if the photo is clear enough for proper analysis
- If the image is too blurry, unclear, or has poor quality that prevents proper evaluation, respond with:
FAIL
Photo too blurry - please retake with better focus

- Only proceed with the main check if photo quality is acceptable

Check if a spirit level/level indicator is present on the appliance.
Look for: spirit level tool visible on or near the appliance, level indicator present, measuring tool for leveling.

RESPONSE FORMAT - FOLLOW EXACTLY:
- First line: "PASS" or "FAIL"
- Second line: Brief explanation (max 100 characters) why it passed or failed`

		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unknown endpoint"})
			return
		}

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

		// Parse AI response (verwacht 2 regels: result en reason)
		aiResponse := resp.Choices[0].Message.Content
		lines := strings.Split(strings.TrimSpace(aiResponse), "\n")

		var result, reason string
		if len(lines) >= 2 {
			result = strings.TrimSpace(lines[0])
			reason = strings.TrimSpace(lines[1])
		} else if len(lines) == 1 {
			// Fallback voor oude format
			result = strings.TrimSpace(lines[0])
			reason = "No detailed reason provided"
		} else {
			result = "FAIL"
			reason = "Invalid AI response format"
		}

		// Normaliseer result naar PASS/FAIL
		if strings.ToUpper(result) == "PASS" {
			result = "PASS"
		} else {
			result = "FAIL"
		}

		// Stuur Gold response terug
		json.NewEncoder(w).Encode(GoldResponse{
			Result:        result,
			ProjectNumber: projectNumber,
			Reason:        reason,
		})
	})

	log.Println("Server start op :8080")
	http.ListenAndServe(":8080", nil)

}
