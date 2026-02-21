package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const metalertsURL = "https://api.met.no/weatherapi/metalerts/2.0/current.json"
const metalertsCacheTTL = 5 * time.Minute

type metalertsResponse struct {
	Features []metalertsFeature `json:"features"`
}

type metalertsFeature struct {
	Properties metalertsProps `json:"properties"`
}

type metalertsProps struct {
	Event       string `json:"event"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Instruction string `json:"instruction"`
	Area        string `json:"area"`
}

// getWeatherAlerts fetches active weather warnings near a point.
func getWeatherAlerts(ctx context.Context, cache *Cache, lat, lon float64) ([]WeatherAlert, error) {
	u := fmt.Sprintf("%s?lat=%f&lon=%f", metalertsURL, lat, lon)

	data, err := cachedGet(ctx, cache, u, metalertsCacheTTL)
	if err != nil {
		return nil, fmt.Errorf("metalerts: %w", err)
	}

	var result metalertsResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("metalerts decode: %w", err)
	}

	alerts := make([]WeatherAlert, 0, len(result.Features))
	for _, f := range result.Features {
		alerts = append(alerts, WeatherAlert{
			Event:       f.Properties.Event,
			Severity:    f.Properties.Severity,
			Description: f.Properties.Description,
			Instruction: f.Properties.Instruction,
			Area:        f.Properties.Area,
		})
	}
	return alerts, nil
}
