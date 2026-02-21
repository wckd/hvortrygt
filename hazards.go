package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
)

// Named NVE service references to avoid magic indices.
var (
	svcFlood10yr         = nveServices[0]
	svcFlood20yr         = nveServices[1]
	svcFlood50yr         = nveServices[2]
	svcFlood100yr        = nveServices[3]
	svcFlood200yr        = nveServices[4]
	svcFloodAwareness    = nveServices[5]
	svcLandslide         = nveServices[6]
	svcQuickClayDetail   = nveServices[7]
	svcQuickClayOverview = nveServices[8]
	svcAvalanche         = nveServices[9]
	svcRockFall          = nveServices[10]
	svcCombinedHazard    = nveServices[11]
)

// assessHazards runs all hazard checks in parallel and returns results.
// Elevation is fetched first (needed by storm surge), then the rest fan out.
func assessHazards(ctx context.Context, cache *Cache, addr Address) ([]HazardResult, *float64, []WeatherAlert) {
	lat, lon := addr.Latitude, addr.Longitude

	// Fetch elevation first — storm surge depends on it.
	elev, err := getElevation(ctx, cache, lat, lon)
	if err != nil {
		log.Printf("elevation error: %v", err)
	}

	var (
		mu      sync.Mutex
		hazards []HazardResult
		alerts  []WeatherAlert
		wg      sync.WaitGroup
	)

	addHazard := func(h HazardResult) {
		mu.Lock()
		hazards = append(hazards, h)
		mu.Unlock()
	}

	// Flood zone queries (10, 20, 50, 100, 200 year)
	wg.Add(1)
	go func() {
		defer wg.Done()
		addHazard(checkFloodZones(ctx, cache, lat, lon))
	}()

	// Flood awareness
	wg.Add(1)
	go func() {
		defer wg.Done()
		addHazard(checkSingleNVE(ctx, cache, lat, lon, svcFloodAwareness, "flood_awareness", "Flomaktsomhet", "Flomaktsomhetsområde", 35))
	}()

	// Landslide
	wg.Add(1)
	go func() {
		defer wg.Done()
		addHazard(checkSingleNVE(ctx, cache, lat, lon, svcLandslide, "landslide", "Jord- og flomskred", "Aktsomhetsområde for jord- og flomskred", 60))
	}()

	// Quick clay
	wg.Add(1)
	go func() {
		defer wg.Done()
		addHazard(checkQuickClay(ctx, cache, lat, lon))
	}()

	// Avalanche
	wg.Add(1)
	go func() {
		defer wg.Done()
		addHazard(checkSingleNVE(ctx, cache, lat, lon, svcAvalanche, "avalanche", "Snøskred", "Aktsomhetsområde for snøskred", 70))
	}()

	// Rock fall
	wg.Add(1)
	go func() {
		defer wg.Done()
		addHazard(checkSingleNVE(ctx, cache, lat, lon, svcRockFall, "rock_fall", "Steinsprang", "Aktsomhetsområde for steinsprang", 65))
	}()

	// Combined hazard zones
	wg.Add(1)
	go func() {
		defer wg.Done()
		addHazard(checkSingleNVE(ctx, cache, lat, lon, svcCombinedHazard, "combined_hazard", "Skredfaresoner", "Faresone for skred", 75))
	}()

	// Storm surge (uses elevation — now safe, fetched above)
	wg.Add(1)
	go func() {
		defer wg.Done()
		addHazard(checkStormSurge(ctx, cache, addr.Kommunenummer, elev))
	}()

	// Weather alerts
	wg.Add(1)
	go func() {
		defer wg.Done()
		a, err := getWeatherAlerts(ctx, cache, lat, lon)
		if err != nil {
			log.Printf("metalerts error: %v", err)
			return
		}
		mu.Lock()
		alerts = a
		mu.Unlock()
	}()

	wg.Wait()
	return hazards, elev, alerts
}

// checkFloodZones queries all flood return period layers and returns
// a single hazard result for the worst match.
func checkFloodZones(ctx context.Context, cache *Cache, lat, lon float64) HazardResult {
	type floodLevel struct {
		svc   nveService
		label string
		score int
	}

	levels := []floodLevel{
		{svcFlood10yr, "10-årsflom", 90},
		{svcFlood20yr, "20-årsflom", 75},
		{svcFlood50yr, "50-årsflom", 55},
		{svcFlood100yr, "100-årsflom", 40},
		{svcFlood200yr, "200-årsflom", 25},
	}

	bestScore := 0
	bestLabel := ""

	for _, fl := range levels {
		resp, err := queryNVE(ctx, cache, fl.svc, lat, lon)
		if err != nil {
			log.Printf("flood query %s: %v", fl.label, err)
			continue
		}
		if len(resp.Features) > 0 && fl.score > bestScore {
			bestScore = fl.score
			bestLabel = fl.label
		}
	}

	h := HazardResult{
		ID:    "flood_zones",
		Name:  "Flomsoner",
		Score: bestScore,
		Level: scoreLevel(bestScore),
	}

	if bestScore > 0 {
		h.Description = fmt.Sprintf("Innenfor %s-sone", bestLabel)
		h.Details = fmt.Sprintf("Adressen ligger i en kartlagt flomsone (%s). Risiko for oversvømmelse ved ekstremvær.", bestLabel)
	} else {
		h.Description = "Ikke i kartlagt flomsone"
		h.Details = "Ingen registrerte flomsoner på dette punktet."
	}

	return h
}

// checkSingleNVE queries a single NVE service and returns present/absent.
func checkSingleNVE(ctx context.Context, cache *Cache, lat, lon float64, svc nveService, id, name, desc string, presentScore int) HazardResult {
	h := HazardResult{
		ID:   id,
		Name: name,
	}

	resp, err := queryNVE(ctx, cache, svc, lat, lon)
	if err != nil {
		h.Error = "Kunne ikke hente data"
		h.Level = "unknown"
		log.Printf("nve %s error: %v", id, err)
		return h
	}

	if len(resp.Features) > 0 {
		h.Score = presentScore
		h.Level = scoreLevel(presentScore)
		h.Description = desc
		h.Details = fmt.Sprintf("Adressen ligger i %s.", strings.ToLower(desc))
	} else {
		h.Score = 0
		h.Level = scoreLevel(0)
		h.Description = "Ikke i aktsomhetsområde"
		h.Details = fmt.Sprintf("Ingen registrert %s-fare på dette punktet.", strings.ToLower(name))
	}

	return h
}

// checkQuickClay queries both detailed and overview quick clay services.
func checkQuickClay(ctx context.Context, cache *Cache, lat, lon float64) HazardResult {
	h := HazardResult{
		ID:   "quick_clay",
		Name: "Kvikkleire",
	}

	// Try detailed first
	resp, err := queryNVE(ctx, cache, svcQuickClayDetail, lat, lon)
	if err == nil && len(resp.Features) > 0 {
		grade := extractFaregrad(resp.Features[0].Attributes)
		switch grade {
		case "Hoy", "Høy":
			h.Score = 80
		case "Middels":
			h.Score = 50
		default:
			h.Score = 25
		}
		h.Level = scoreLevel(h.Score)
		h.Description = fmt.Sprintf("Kvikkleiresone (faregrad: %s)", grade)
		h.Details = "Adressen ligger i område med kartlagt kvikkleirefare."
		return h
	}

	// Fallback to overview
	resp2, err := queryNVE(ctx, cache, svcQuickClayOverview, lat, lon)
	if err != nil {
		h.Error = "Kunne ikke hente data"
		h.Level = "unknown"
		return h
	}

	if len(resp2.Features) > 0 {
		h.Score = 40
		h.Level = scoreLevel(h.Score)
		h.Description = "Aktsomhetsområde for kvikkleire"
		h.Details = "Adressen ligger i et generelt aktsomhetsområde for kvikkleire."
	} else {
		h.Score = 0
		h.Level = scoreLevel(0)
		h.Description = "Ikke i kvikkleireområde"
		h.Details = "Ingen registrert kvikkleirefare på dette punktet."
	}

	return h
}

func extractFaregrad(attrs map[string]any) string {
	for _, key := range []string{"faregrad", "Faregrad", "FAREGRAD", "faregradTekst"} {
		if v, ok := attrs[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return "Ukjent"
}

// checkStormSurge evaluates storm surge risk based on municipality consequence data.
func checkStormSurge(ctx context.Context, cache *Cache, kommunenummer string, elevation *float64) HazardResult {
	h := HazardResult{
		ID:   "storm_surge",
		Name: "Stormflo",
	}

	entries, err := getStormSurge(ctx, cache, kommunenummer)
	if err != nil || len(entries) == 0 {
		h.Score = 0
		h.Level = scoreLevel(0)
		h.Description = "Ingen stormflodata"
		h.Details = "Ingen stormflodata tilgjengelig for denne kommunen."
		return h
	}

	if elevation != nil && *elevation < 3 {
		h.Score = 50
		h.Level = scoreLevel(h.Score)
		h.Description = fmt.Sprintf("Lav kystbeliggenhet (%.1f moh.)", *elevation)
		h.Details = fmt.Sprintf("Adressen ligger på bare %.1f moh. i en kystkommune med stormflorisiko. Kan bli berørt ved ekstreme stormflosituasjoner.", *elevation)
	} else if elevation != nil && *elevation < 10 {
		h.Score = 25
		h.Level = scoreLevel(h.Score)
		h.Description = fmt.Sprintf("Kystnær beliggenhet (%.1f moh.)", *elevation)
		h.Details = fmt.Sprintf("Adressen ligger på %.1f moh. i en kystkommune. Moderat risiko for stormflo.", *elevation)
	} else {
		h.Score = 0
		h.Level = scoreLevel(0)
		h.Description = "Over stormflonivå"
		h.Details = "Adressen ligger høyt nok til at stormflo neppe er en trussel."
	}

	return h
}
