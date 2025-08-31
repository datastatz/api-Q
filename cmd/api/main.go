package main

import (
	"encoding/json"
	"net/http"
)

func main() {

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
			response := map[string]string{"error": "Could not parse multipart form"}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Haal de afbeelding op uit het formulier
		description := r.FormValue("description")

		// Voor nu gewoon de beschrijving terug geven
		response := map[string]string{
			"result":      "Pass",
			"reason":      "Description received: " + description,
			"description": description,
		}

		// Encodeer de response als JSON en stuur terug naar de client

		json.NewEncoder(w).Encode(response)
	})

	//Start een HTTP server op poort 8080, nil geef aan dat er nog geen routes zijn gedefinieerd
	http.ListenAndServe(":8080", nil)

}
