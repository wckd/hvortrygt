package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const elevationURL = "https://ws.geonorge.no/hoydedata/v1/punkt"
const elevationCacheTTL = 24 * time.Hour

type elevationResponse struct {
	Points []elevationPoint `json:"punkter"`
}

type elevationPoint struct {
	Z     *float64 `json:"z"`
	Datakilde string `json:"datakilde"`
}

// getElevation returns the elevation in meters at the given coordinates.
func getElevation(ctx context.Context, cache *Cache, lat, lon float64) (*float64, error) {
	u := fmt.Sprintf("%s?nord=%f&ost=%f&koordsys=4326&geession=false", elevationURL, lat, lon)

	data, err := cachedGet(ctx, cache, u, elevationCacheTTL)
	if err != nil {
		return nil, fmt.Errorf("elevation: %w", err)
	}

	var result elevationResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("elevation decode: %w", err)
	}

	if len(result.Points) > 0 && result.Points[0].Z != nil {
		return result.Points[0].Z, nil
	}
	return nil, nil
}
