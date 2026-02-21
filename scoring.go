package main

import "fmt"

// calculateRisk computes the overall risk score and Norwegian summary.
func calculateRisk(hazards []HazardResult, elevation *float64, kommunenummer string) (int, string, string) {
	maxScore := 0
	for _, h := range hazards {
		if h.Score > maxScore {
			maxScore = h.Score
		}
	}

	// Coastal low-elevation boost
	if elevation != nil && *elevation < 5 && isCoastalMunicipality(kommunenummer) {
		maxScore += 10
		if maxScore > 100 {
			maxScore = 100
		}
	}

	level := scoreLevel(maxScore)
	summary := scoreSummary(maxScore, level)

	return maxScore, level, summary
}

// scoreSummary returns a Norwegian human-readable summary for the score.
func scoreSummary(score int, level string) string {
	switch level {
	case "low":
		return "Adressen har lav risiko for naturfare. Ingen kjente faresoner er registrert her."
	case "medium":
		return "Adressen har moderat risiko. Det er registrert noen aktsomhetsområder i nærheten."
	case "high":
		return fmt.Sprintf("Adressen har høy risiko (score %d/100). Én eller flere naturfaresoner er registrert på dette punktet.", score)
	case "very_high":
		return fmt.Sprintf("Adressen har svært høy risiko (score %d/100). Flere alvorlige naturfaresoner er registrert. Vurder å innhente profesjonell vurdering.", score)
	default:
		return ""
	}
}

// isCoastalMunicipality checks if a municipality is coastal.
// This is a simplified check using the first two digits of kommunenummer (county).
func isCoastalMunicipality(knr string) bool {
	if len(knr) < 2 {
		return false
	}
	coastalCounties := map[string]bool{
		"03": true, // Oslo
		"11": true, // Rogaland
		"15": true, // Møre og Romsdal
		"18": true, // Nordland
		"30": true, // Viken (coastal parts)
		"33": true, // Buskerud (was Viken)
		"34": true, // Østfold (was Viken)
		"32": true, // Akershus (was Viken)
		"38": true, // Vestfold og Telemark
		"39": true, // Vestfold
		"40": true, // Telemark
		"42": true, // Agder
		"46": true, // Vestland
		"50": true, // Trøndelag
		"55": true, // Troms
		"56": true, // Finnmark
	}
	return coastalCounties[knr[:2]]
}
