package main

// Address represents a geocoded Norwegian address from Kartverket.
type Address struct {
	Text           string  `json:"text"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	Kommunenummer  string  `json:"kommunenummer"`
	Kommunenavn    string  `json:"kommunenavn"`
	Postnummer     string  `json:"postnummer"`
	Poststed       string  `json:"poststed"`
}

// HazardResult holds the outcome of a single hazard check.
type HazardResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Score       int    `json:"score"`     // 0-100
	Level       string `json:"level"`     // low, medium, high, very_high
	Details     string `json:"details"`   // Norwegian human-readable detail
	Error       string `json:"error,omitempty"`
}

// WeatherAlert represents an active MET weather warning.
type WeatherAlert struct {
	Event       string `json:"event"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Instruction string `json:"instruction"`
	Area        string `json:"area"`
}

// RiskResponse is the full response for a risk assessment.
type RiskResponse struct {
	Address       Address        `json:"address"`
	OverallScore  int            `json:"overall_score"`
	OverallLevel  string         `json:"overall_level"`
	Summary       string         `json:"summary"`
	Elevation     *float64       `json:"elevation,omitempty"`
	Hazards       []HazardResult `json:"hazards"`
	WeatherAlerts []WeatherAlert `json:"weather_alerts"`
}

// scoreLevel returns the risk level string for a given score.
func scoreLevel(score int) string {
	switch {
	case score <= 15:
		return "low"
	case score <= 40:
		return "medium"
	case score <= 70:
		return "high"
	default:
		return "very_high"
	}
}
