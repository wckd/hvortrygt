package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// NVE ArcGIS REST endpoint base URLs and layers.
type nveService struct {
	Name    string
	BaseURL string
	Layer   int
}

// All services use nve.geodataonline.no with correct layer IDs.
var nveServices = []nveService{
	// Flood zones by return period
	{Name: "flood_10yr", BaseURL: "https://nve.geodataonline.no/arcgis/rest/services/Flomsoner1/MapServer", Layer: 11},
	{Name: "flood_20yr", BaseURL: "https://nve.geodataonline.no/arcgis/rest/services/Flomsoner1/MapServer", Layer: 12},
	{Name: "flood_50yr", BaseURL: "https://nve.geodataonline.no/arcgis/rest/services/Flomsoner1/MapServer", Layer: 13},
	{Name: "flood_100yr", BaseURL: "https://nve.geodataonline.no/arcgis/rest/services/Flomsoner1/MapServer", Layer: 14},
	{Name: "flood_200yr", BaseURL: "https://nve.geodataonline.no/arcgis/rest/services/Flomsoner1/MapServer", Layer: 15},
	// Flood awareness
	{Name: "flood_awareness", BaseURL: "https://nve.geodataonline.no/arcgis/rest/services/FlomAktsomhet/MapServer", Layer: 1},
	// Landslide/debris flow (combined snow+stone+debris awareness)
	{Name: "landslide", BaseURL: "https://nve.geodataonline.no/arcgis/rest/services/SkredSnoSteinAkt/MapServer", Layer: 0},
	// Quick clay detailed
	{Name: "quick_clay_detailed", BaseURL: "https://nve.geodataonline.no/arcgis/rest/services/SkredKvikkleire2/MapServer", Layer: 0},
	// Quick clay overview
	{Name: "quick_clay_overview", BaseURL: "https://nve.geodataonline.no/arcgis/rest/services/KvikkleireskredAktsomhet/MapServer", Layer: 0},
	// Avalanche
	{Name: "avalanche", BaseURL: "https://nve.geodataonline.no/arcgis/rest/services/SnoskredAktsomhet/MapServer", Layer: 1},
	// Rock fall
	{Name: "rock_fall", BaseURL: "https://nve.geodataonline.no/arcgis/rest/services/SkredSteinAktR/MapServer", Layer: 2},
	// Combined hazard zones (100-year)
	{Name: "combined_hazard", BaseURL: "https://nve.geodataonline.no/arcgis/rest/services/Skredfaresoner2/MapServer", Layer: 2},
}

type arcgisResponse struct {
	Features []arcgisFeature `json:"features"`
}

type arcgisFeature struct {
	Attributes map[string]any `json:"attributes"`
}

const nveCacheTTL = 1 * time.Hour

// queryNVE performs a point-intersect query against an NVE ArcGIS REST service.
func queryNVE(ctx context.Context, cache *Cache, svc nveService, lat, lon float64) (*arcgisResponse, error) {
	u := fmt.Sprintf(
		"%s/%d/query?geometry=%f,%f&geometryType=esriGeometryPoint&inSR=4326&spatialRel=esriSpatialRelIntersects&outFields=*&returnGeometry=false&f=json",
		svc.BaseURL, svc.Layer, lon, lat,
	)

	data, err := cachedGet(ctx, cache, u, nveCacheTTL)
	if err != nil {
		return nil, fmt.Errorf("nve %s: %w", svc.Name, err)
	}

	var result arcgisResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("nve %s decode: %w", svc.Name, err)
	}

	return &result, nil
}
