package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("set INFRAI_API_KEY")
	}
	c := client{http: &http.Client{Timeout: 30 * time.Second}, key: key}
	http.HandleFunc("POST /review", func(w http.ResponseWriter, r *http.Request) {
		var in ReviewInput
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(in.RequestID) == "" || len(in.SellerAssets)+len(in.BuyerUpdates)+len(in.OrderHandoff) == 0 {
			http.Error(w, "request_id and at least one activity required", http.StatusBadRequest)
			return
		}
		keys, err := c.keys(r.Context())
		if err != nil {
			respondError(w, err)
			return
		}
		rows := evaluate(in, keys)
		date := time.Now().UTC().Format("2006-01-02")
		pdf, err := c.pdf(r.Context(), in.RequestID, document(date, rows))
		if err != nil {
			respondError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Date     string          `json:"date"`
			Reach    []Reach         `json:"reach"`
			Document json.RawMessage `json:"document"`
		}{date, rows, pdf})
	})
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Fatal(http.ListenAndServe(addr, nil))
}

func respondError(w http.ResponseWriter, err error) {
	var api apiError
	if errors.As(err, &api) && api.Status >= 400 && api.Status < 500 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(api.Status)
		json.NewEncoder(w).Encode(map[string]any{"error": api.Detail})
		return
	}
	http.Error(w, "upstream request failed", http.StatusBadGateway)
}
