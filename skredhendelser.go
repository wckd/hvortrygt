package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/url"
	"sort"
	"strings"
	"time"
)

const skredHendelserURL = "https://gis3.nve.no/map/rest/services/Mapservices/SkredHendelser/MapServer/0/query"
const skredSearchRadiusKm = 1.0
const skredMaxResults = 50

// skredHendelserResponse is the ArcGIS JSON response for the SkredHendelser layer.
type skredHendelserResponse struct {
	Features []skredHendelserFeature `json:"features"`
}

type skredHendelserFeature struct {
	Attributes map[string]any     `json:"attributes"`
	Geometry   *skredPointGeometry `json:"geometry"`
}

type skredPointGeometry struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// getSkredHendelser queries NVE's historical landslide event database
// within a bounding box, filters to a circular radius, scores, and returns
// both a HazardResult and the individual events for display.
func getSkredHendelser(ctx context.Context, cache *Cache, lat, lon float64) (HazardResult, []HistoricalEvent) {
	h := HazardResult{
		ID:   "historical_landslides",
		Name: "Historiske skredhendelser",
	}

	events, err := fetchSkredHendelser(ctx, cache, lat, lon)
	if err != nil {
		log.Printf("skredhendelser error: %v", err)
		h.Error = "Kunne ikke hente data"
		h.Level = "unknown"
		return h, nil
	}

	if len(events) == 0 {
		h.Score = 0
		h.Level = scoreLevel(0)
		h.Description = "Ingen registrerte skredhendelser"
		h.Details = "Ingen historiske skredhendelser er registrert innenfor 1 km."
		return h, nil
	}

	score := scoreHistoricalEvents(events)
	h.Score = score
	h.Level = scoreLevel(score)

	if len(events) == 1 {
		h.Description = "1 historisk skredhendelse innenfor 1 km"
	} else {
		h.Description = fmt.Sprintf("%d historiske skredhendelser innenfor 1 km", len(events))
	}

	// Build details summary
	var parts []string
	for i, e := range events {
		if i >= 3 {
			parts = append(parts, fmt.Sprintf("og %d til.", len(events)-3))
			break
		}
		detail := e.Type
		if e.Date != "" {
			detail += " (" + e.Date + ")"
		}
		detail += fmt.Sprintf(", %d m unna", e.DistanceMeters)
		parts = append(parts, detail)
	}
	h.Details = strings.Join(parts, ". ")

	return h, events
}

// fetchSkredHendelser queries the NVE SkredHendelser API with a bounding box,
// then filters to a circular radius and deduplicates by skredID.
func fetchSkredHendelser(ctx context.Context, cache *Cache, lat, lon float64) ([]HistoricalEvent, error) {
	dLat := skredSearchRadiusKm / 111.0
	dLon := skredSearchRadiusKm / (111.0 * math.Cos(lat*math.Pi/180.0))

	bbox := fmt.Sprintf("%f,%f,%f,%f", lon-dLon, lat-dLat, lon+dLon, lat+dLat)
	params := url.Values{
		"geometry":          {bbox},
		"geometryType":      {"esriGeometryEnvelope"},
		"inSR":              {"4326"},
		"outSR":             {"4326"},
		"spatialRel":        {"esriSpatialRelIntersects"},
		"outFields":         {"*"},
		"returnGeometry":    {"true"},
		"resultRecordCount": {fmt.Sprintf("%d", skredMaxResults)},
		"f":                 {"json"},
	}
	u := skredHendelserURL + "?" + params.Encode()

	data, err := cachedGet(ctx, cache, u, nveCacheTTL)
	if err != nil {
		return nil, fmt.Errorf("skredhendelser fetch: %w", err)
	}

	var result skredHendelserResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("skredhendelser decode: %w", err)
	}

	seen := make(map[int64]bool)
	seenCoord := make(map[string]bool)
	var events []HistoricalEvent

	for _, f := range result.Features {
		if f.Geometry == nil {
			continue
		}

		// Deduplicate: use skredID when present, otherwise fall back to
		// a coordinate-based key (5 decimal places ~ 1 m precision).
		if id, ok := attrInt(f.Attributes, "skredID"); ok {
			if seen[id] {
				continue
			}
			seen[id] = true
		} else {
			coordKey := fmt.Sprintf("%.5f,%.5f", f.Geometry.X, f.Geometry.Y)
			if seenCoord[coordKey] {
				continue
			}
			seenCoord[coordKey] = true
		}

		evtLat := f.Geometry.Y
		evtLon := f.Geometry.X

		dist := haversineMeters(lat, lon, evtLat, evtLon)
		if dist > skredSearchRadiusKm*1000 {
			continue
		}

		e := HistoricalEvent{
			Type:           attrString(f.Attributes, "skredtype", "typeNavn", "skredtypeNavn"),
			Date:           parseSkredDate(f.Attributes),
			Location:       attrString(f.Attributes, "sted", "stedsnavn"),
			BuildingDamage: attrString(f.Attributes, "bygnSkadet") == "Ja",
			RoadDamage:     attrString(f.Attributes, "vegSkadet") == "Ja",
			Fatalities:     attrIntVal(f.Attributes, "dodsfall", "antallOmkommet"),
			Description:    attrString(f.Attributes, "beskrivelse", "hendelseBeskrivelse"),
			Latitude:       evtLat,
			Longitude:      evtLon,
			DistanceMeters: int(dist),
		}

		if e.Type == "" {
			e.Type = "Skred (ukjent type)"
		}

		events = append(events, e)
	}

	// Sort by distance (closest first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].DistanceMeters < events[j].DistanceMeters
	})

	return events, nil
}

// scoreHistoricalEvents computes an aggregate score from individual events.
func scoreHistoricalEvents(events []HistoricalEvent) int {
	now := time.Now()

	var scores []int
	hasFatalitiesClose := false

	for _, e := range events {
		s := 10 // base: event exists within 1km

		if e.BuildingDamage {
			s += 25
		}
		if e.RoadDamage {
			s += 10
		}
		if e.Fatalities > 0 {
			s += 30
			if e.DistanceMeters < 200 {
				hasFatalitiesClose = true
			}
		}

		// Recency bonus
		if e.Date != "" {
			if t, err := time.Parse("2006-01-02", e.Date); err == nil {
				years := now.Sub(t).Hours() / (365.25 * 24)
				if years < 20 {
					s += 10
				} else if years < 50 {
					s += 5
				}
			}
		}

		// Proximity bonus
		if e.DistanceMeters < 200 {
			s += 15
		} else if e.DistanceMeters < 500 {
			s += 5
		}

		if s > 100 {
			s = 100
		}
		scores = append(scores, s)
	}

	// Sort individual scores descending for diminishing returns
	sort.Sort(sort.Reverse(sort.IntSlice(scores)))

	total := 0.0
	for i, s := range scores {
		weight := 1.0 / math.Pow(2, float64(i))
		total += float64(s) * weight
	}

	result := int(total)
	if result > 85 {
		result = 85
	}

	// Floor at 75 if fatalities within 200m
	if hasFatalitiesClose && result < 75 {
		result = 75
	}

	return result
}

// haversineMeters returns the distance in meters between two lat/lon points.
func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusM = 6_371_000.0
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	la1 := lat1 * math.Pi / 180.0
	la2 := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(la1)*math.Cos(la2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}

// parseSkredDate extracts and formats the event date from attributes.
// NVE uses epoch milliseconds or date strings.
func parseSkredDate(attrs map[string]any) string {
	for _, key := range []string{"skredTidspunkt", "dato", "skredDato"} {
		v, ok := attrs[key]
		if !ok || v == nil {
			continue
		}
		switch val := v.(type) {
		case float64:
			if val > 0 {
				t := time.UnixMilli(int64(val))
				return t.Format("2006-01-02")
			}
		case string:
			if val != "" {
				if t, err := time.Parse("2006-01-02", val); err == nil {
					return t.Format("2006-01-02")
				}
				if t, err := time.Parse(time.RFC3339, val); err == nil {
					return t.Format("2006-01-02")
				}
				// Unrecognized date format — discard rather than pass raw API data through.
				return ""
			}
		}
	}
	return ""
}

// attrString returns the first non-empty string value found for the given keys.
func attrString(attrs map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := attrs[k]; ok && v != nil {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

// attrInt returns the first integer value found for a key.
func attrInt(attrs map[string]any, keys ...string) (int64, bool) {
	for _, k := range keys {
		if v, ok := attrs[k]; ok && v != nil {
			switch val := v.(type) {
			case float64:
				return int64(val), true
			}
		}
	}
	return 0, false
}

// attrIntVal returns the first non-zero int value found for the given keys.
func attrIntVal(attrs map[string]any, keys ...string) int {
	for _, k := range keys {
		if v, ok := attrs[k]; ok && v != nil {
			if f, ok := v.(float64); ok && f > 0 {
				return int(f)
			}
		}
	}
	return 0
}
