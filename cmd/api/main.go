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
		Reason      string `json:"reason"`      // Reden waarom pass of fail
		Filesize    string `json:"filesize"`    // Grootte van het bestand
	}

	// Country code to language code mapping
	countryToLanguage := map[string]string{
		"AF": "Pashto", "AL": "Albanian", "DZ": "Arabic", "AD": "Catalan",
		"AO": "Portuguese", "AR": "Spanish", "AM": "Armenian", "AT": "German",
		"AZ": "Azerbaijani", "BS": "English", "BH": "Arabic", "BD": "Bengali",
		"BB": "English", "BY": "Belarusian", "BE": "French", "BZ": "English",
		"BJ": "French", "BT": "Dzongkha", "BO": "Spanish", "BA": "Bosnian",
		"BW": "English", "BR": "Portuguese", "BN": "Malay", "BG": "Bulgarian",
		"BF": "French", "BI": "French", "CV": "Portuguese", "KH": "Khmer",
		"CM": "French", "CA": "English", "CF": "French", "TD": "French",
		"CL": "Spanish", "CN": "Chinese", "CO": "Spanish", "KM": "Arabic",
		"CG": "French", "CR": "Spanish", "HR": "Croatian", "CU": "Spanish",
		"CY": "Greek", "CZ": "Czech", "DK": "Danish", "DJ": "French",
		"DM": "English", "DO": "Spanish", "EC": "Spanish", "EG": "Arabic",
		"SV": "Spanish", "GQ": "Spanish", "ER": "Tigrinya", "EE": "Estonian",
		"SZ": "English", "ET": "Amharic", "FJ": "English", "FI": "Finnish",
		"FR": "French", "GA": "French", "GM": "English", "GE": "Georgian",
		"DE": "German", "GH": "English", "GR": "Greek", "GD": "English",
		"GT": "Spanish", "GN": "French", "GW": "Portuguese", "GY": "English",
		"HT": "French", "HN": "Spanish", "HU": "Hungarian", "IS": "Icelandic",
		"IN": "Hindi", "ID": "Indonesian", "IR": "Persian", "IQ": "Arabic",
		"IE": "English", "IT": "Italian", "JM": "English",
		"JP": "Japanese", "JO": "Arabic", "KZ": "Kazakh", "KE": "Swahili",
		"KI": "English", "KP": "Korean", "KR": "Korean", "KW": "Arabic",
		"KG": "Kyrgyz", "LA": "Lao", "LV": "Latvian", "LB": "Arabic",
		"LS": "English", "LR": "English", "LY": "Arabic", "LI": "German",
		"LT": "Lithuanian", "LU": "French", "MG": "French", "MW": "English",
		"MY": "Malay", "MV": "Dhivehi", "ML": "French", "MT": "Maltese",
		"MH": "English", "MR": "Arabic", "MU": "English", "MX": "Spanish",
		"FM": "English", "MD": "Romanian", "MC": "French", "MN": "Mongolian",
		"ME": "Serbian", "MA": "Arabic", "MZ": "Portuguese", "MM": "Burmese",
		"NA": "English", "NR": "English", "NP": "Nepali", "NL": "Dutch",
		"NZ": "English", "NI": "Spanish", "NE": "French", "NG": "English",
		"NO": "Norwegian", "OM": "Arabic", "PK": "Urdu", "PW": "English",
		"PA": "Spanish", "PG": "English", "PY": "Spanish", "PE": "Spanish",
		"PH": "Tagalog", "PL": "Polish", "PT": "Portuguese", "QA": "Arabic",
		"RO": "Romanian", "RU": "Russian", "RW": "French", "KN": "English",
		"LC": "English", "VC": "English", "WS": "English", "SM": "Italian",
		"ST": "Portuguese", "SA": "Arabic", "SN": "French", "RS": "Serbian",
		"SC": "French", "SL": "English", "SG": "English", "SK": "Slovak",
		"SI": "Slovenian", "SB": "English", "SO": "Somali", "ZA": "English",
		"SS": "English", "ES": "Spanish", "LK": "Sinhala", "SD": "Arabic",
		"SR": "Dutch", "SE": "Swedish", "CH": "German", "SY": "Arabic",
		"TW": "Chinese", "TJ": "Tajik", "TZ": "Swahili", "TH": "Thai",
		"TL": "Portuguese", "TG": "French", "TO": "English", "TT": "English",
		"TN": "Arabic", "TR": "Turkish", "TM": "Turkmen", "TV": "English",
		"UG": "English", "UA": "Ukrainian", "AE": "Arabic", "GB": "English",
		"US": "English", "UY": "Spanish", "UZ": "Uzbek", "VU": "French",
		"VA": "Italian", "VE": "Spanish", "VN": "Vietnamese", "YE": "Arabic",
		"ZM": "English", "ZW": "English",
	}

	// Route voor kwaliteitscontrole (alleen post requests)
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

		// Haal language parameter op, default is "en"
		language := r.FormValue("language")
		if language != "" {

			// Check of het een landcode is (2 letters)
			// Zo ja, converteer naar taalnaam

			if langName, exists := countryToLanguage[strings.ToUpper(language)]; exists {
				language = langName
			}
		}

		// Initialiseer een array om alle foto analyses in op te slaan
		var analyses []PhotoAnalysis

		// Loop doot mogelijke fotos (photo1 t/m photo5)
		// Maximaal 5 fotos per request om performance te behouden
		for i := 1; i <= 5; i++ {

			// Genereer fieldnames: photo1, photo2, ..., photo5
			photoField := fmt.Sprintf("photo%d", i)
			descriptionField := fmt.Sprintf("description%d", i)

			// Probeer foto op te halen uit form data
			file, header, err := r.FormFile(photoField)
			if err != nil {
				// Geen foto gevonden met dit nummer, stop met zoeken
				// Dit is normaal als minder dan 5 fotos zijn geupload
				break
			}
			defer file.Close() // Zorg dat bestand gesloten wordt na gebruik

			// Haal bijbehorende beschrijving op
			description := r.FormValue(descriptionField)

			if description == "" {
				w.WriteHeader(http.StatusBadRequest)
				response := map[string]string{
					"error": fmt.Sprintf("Missing description%d for photo%d", i, i),
				}
				json.NewEncoder(w).Encode(response)
				return
			}

			filename := header.Filename
			filesize := header.Size

			// Lees de foto inhoud naar memory
			photoBytes, err := io.ReadAll(file)
			if err != nil {

				// Als foto niet gelezen kan worden, sla error op en ga door naar volgende foto
				analyses = append(analyses, PhotoAnalysis{
					Filename:    filename,
					Description: description,
					Result:      "error",
					Reason:      "Could not read photo file",
					Filesize:    fmt.Sprintf("%d bytes", filesize),
				})
				continue // Ga door naar volgende foto
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

			// Bouw slimme system prompt die alle talen ondersteunt
			systemPrompt := fmt.Sprintf(`You are a quality control expert for installations. 

			CRITICAL INSTRUCTION: You MUST respond ENTIRELY in %s language. This is MANDATORY.

			Requirements to analyze: %s

			RESPONSE FORMAT - FOLLOW EXACTLY:
			- Start with either "PASS:" or "FAIL:"
			- Follow with your reason in %s language
			- Example: "FAIL: La machine est démontée et non sécurisée"
			- Example: "PASS: L'installation semble correcte et sécurisée"

			DO NOT respond in English if %s was specified.`, language, description, language, language)

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
				// OpenAI API error - voeg error toe maar ga door met andere foto's
				analyses = append(analyses, PhotoAnalysis{
					Filename:    filename,
					Description: description,
					Result:      "error",
					Reason:      "OpenAI API error: " + err.Error(),
					Filesize:    fmt.Sprintf("%d bytes", filesize),
				})
				continue // Ga door naar volgende foto
			}

			// Parse OpenAI response
			aiResponse := resp.Choices[0].Message.Content

			// Initialiseer variabelen voor resultaat en reden
			var aiResult, aiReason string

			// Check if response begint met "PASS"
			if len(aiResponse) > 5 && aiResponse[:4] == "PASS" {
				aiResult = "pass"
				aiReason = aiResponse[6:] // SKIP "PASS: " deel

				// Check if response begint met "FAIL"
			} else if len(aiResponse) > 5 && aiResponse[:4] == "FAIL" {
				aiResult = "fail"
				aiReason = aiResponse[6:] // SKIP "FAIL: " deel

				// Onverwacht antwoord van de AI
			} else {
				aiResult = "unknown"
				aiReason = aiResponse // Gebruik vollegide response
			}

			// Voeg analyse resultaat toe aan array
			analyses = append(analyses, PhotoAnalysis{
				Filename:    filename,
				Description: description,
				Result:      aiResult,
				Reason:      aiReason,
				Filesize:    fmt.Sprintf("%d bytes", filesize),
			})
		}

		// Controleer of er uberhaupt fotos zijn gevonden
		if len(analyses) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			response := map[string]string{
				"error": "No photos found. Use photo1, photo2, etc. with corresponding description1, description2, etc.",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Bouw finale response met alle analyses
		response := map[string]interface{}{
			"results":      analyses,      // Array van alle foto analyses
			"total_photos": len(analyses), // Totaal aantal geanalyseerde foto's
			"language":     language,      // Gebruikte taal (kan leeg zijn)
		}

		// Encodeer de response als JSON en stuur terug naar de client

		json.NewEncoder(w).Encode(response)
	})

	//Start een HTTP server op poort 8080, nil geef aan dat er nog geen routes zijn gedefinieerd
	log.Println("Server start op :8080")
	http.ListenAndServe(":8080", nil)

}
