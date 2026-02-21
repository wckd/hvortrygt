package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const stormfloBaseURL = "https://stormflo-konsekvens.kartverket.no/public/api/v1"
const stormfloCacheTTL = 24 * time.Hour

// stormfloEntry is one scenario from the consequence API.
type stormfloEntry struct {
	Kommunenummer string `json:"kommunenummer"`
	Code          string `json:"code"`
	Year          string `json:"year"`
	BygningTotal  int    `json:"bygning_total"`
}

// getStormSurge fetches storm surge consequence data for a municipality.
// Returns the full list of scenario entries (e.g. 20y, 200y, 1000y).
func getStormSurge(ctx context.Context, cache *Cache, kommunenummer string) ([]stormfloEntry, error) {
	u := fmt.Sprintf("%s/%s.json", stormfloBaseURL, kommunenummer)

	data, err := cachedGet(ctx, cache, u, stormfloCacheTTL)
	if err != nil {
		// Many inland municipalities return 404 — not an error.
		return nil, nil
	}

	var entries []stormfloEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("stormflo decode: %w", err)
	}

	return entries, nil
}
