package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const geonorgeSearchURL = "https://ws.geonorge.no/adresser/v1/sok"

type geonorgeResponse struct {
	Adresser []geonorgeAddress `json:"adresser"`
}

type geonorgeAddress struct {
	Adressetekst         string        `json:"adressetekst"`
	Kommunenummer        string        `json:"kommunenummer"`
	Kommunenavn          string        `json:"kommunenavn"`
	Postnummer           string        `json:"postnummer"`
	Poststed             string        `json:"poststed"`
	Representasjonspunkt geonorgePoint `json:"representasjonspunkt"`
}

type geonorgePoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// searchAddresses queries Kartverket for address matches.
func searchAddresses(ctx context.Context, query string) ([]Address, error) {
	u := fmt.Sprintf("%s?sok=%s&treffPerSide=5&utkoordsys=4326",
		geonorgeSearchURL, url.QueryEscape(query))

	req, err := newGetRequest(ctx, u)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geocode request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocode: status %d", resp.StatusCode)
	}

	var result geonorgeResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&result); err != nil {
		return nil, fmt.Errorf("geocode decode: %w", err)
	}

	addresses := make([]Address, 0, len(result.Adresser))
	for _, a := range result.Adresser {
		addresses = append(addresses, Address{
			Text:          a.Adressetekst,
			Latitude:      a.Representasjonspunkt.Lat,
			Longitude:     a.Representasjonspunkt.Lon,
			Kommunenummer: a.Kommunenummer,
			Kommunenavn:   a.Kommunenavn,
			Postnummer:    a.Postnummer,
			Poststed:      a.Poststed,
		})
	}
	return addresses, nil
}

func newGetRequest(ctx context.Context, rawURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "hvortrygt/1.0 github.com/hvortrygt")
	return req, nil
}
