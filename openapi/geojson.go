package openapi

import (
	"github.com/peterstark72/golang-skanetrafiken/geo"
)

type PointProperties struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Role     string `json:"role"`
	Distance int    `json:"distance"`
}

type GeometryPoint struct {
	Type        string     `json:"type"`
	Coordinates [2]float64 `json:"coordinates"`
}

type Feature struct {
	Type       string          `json:"type"`
	Geometry   GeometryPoint   `json:"geometry"`
	Properties PointProperties `json:"properties"`
	Id         int             `json:"id"`
}

type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

//LatLng convert X, Y RT90 values to WGS84 lat lon
func (p Point) LatLng() [2]float64 {
	lat, lon := geo.GridToGeodetic(p.X, p.Y)
	return [2]float64{lon, lat}
}

//AsFeature converts a Point into a GeoJSON Feature Point
func (p Point) AsFeature(role string) Feature {
	return Feature{"Feature",
		GeometryPoint{"Point", p.LatLng()},
		PointProperties{p.Name, p.Type, role, 0}, p.Id}

}

//AsFeature converts a NearestStopArea into a GeoJSON Feature Point
func (p NearestStopArea) AsFeature() Feature {
	return Feature{"Feature",
		GeometryPoint{"Point", p.LatLng()},
		PointProperties{p.Name, "STOP_AREA", "NEARBY", p.Distance}, p.Id}

}

//AsFeatureCollection returns a GeoJSON FeatureCollection
func (res GetStartEndPointResult) AsFeatureCollection() FeatureCollection {

	var features []Feature

	for p := range res.StartPoints {
		features = append(features, res.StartPoints[p].AsFeature("START"))
	}

	for p := range res.EndPoints {
		features = append(features, res.EndPoints[p].AsFeature("END"))
	}

	return FeatureCollection{Type: "FeatureCollection", Features: features}
}

//AsFeatureCollection returns a GeoJSON FeatureCollection
func (res GetNearestStopAreaResult) AsFeatureCollection() FeatureCollection {
	var features []Feature

	for p := range res.NearestStopAreas {
		features = append(features, res.NearestStopAreas[p].AsFeature())
	}

	return FeatureCollection{Type: "FeatureCollection", Features: features}
}
