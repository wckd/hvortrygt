package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var knrPattern = regexp.MustCompile(`^\d{4}$`)

func handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" || len(q) < 2 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query too short"})
		return
	}
	if len(q) > 200 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query too long"})
		return
	}

	addresses, err := searchAddresses(r.Context(), q)
	if err != nil {
		log.Printf("search error: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "search failed"})
		return
	}

	writeJSON(w, http.StatusOK, addresses)
}

func handleRisk(cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		lat, err := strconv.ParseFloat(q.Get("lat"), 64)
		if err != nil || lat < 57 || lat > 72 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid latitude"})
			return
		}

		lon, err := strconv.ParseFloat(q.Get("lon"), 64)
		if err != nil || lon < 4 || lon > 32 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid longitude"})
			return
		}

		knr := strings.TrimSpace(q.Get("knr"))
		if !knrPattern.MatchString(knr) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid kommunenummer"})
			return
		}

		text := q.Get("text")
		if len(text) > 500 {
			text = text[:500]
		}
		kommune := q.Get("kommune")
		if len(kommune) > 200 {
			kommune = kommune[:200]
		}

		addr := Address{
			Text:          text,
			Latitude:      lat,
			Longitude:     lon,
			Kommunenummer: knr,
			Kommunenavn:   kommune,
		}

		hazards, elevation, alerts := assessHazards(r.Context(), cache, addr)
		overallScore, overallLevel, summary := calculateRisk(hazards, elevation, knr)

		resp := RiskResponse{
			Address:       addr,
			OverallScore:  overallScore,
			OverallLevel:  overallLevel,
			Summary:       summary,
			Elevation:     elevation,
			Hazards:       hazards,
			WeatherAlerts: alerts,
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("json encode error: %v", err)
	}
}
