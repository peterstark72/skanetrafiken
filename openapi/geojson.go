/*

Methods and structs for creating GeoJSON objects
for the Open API methods that return any of the following types:

- GetStartEndPointResult
- GetNearestStopAreaResult
- GetJourneyPathResult

*/

package openapi

import (
	"encoding/json"
	"io"
)

const (
	FeatureCollectionType = "FeatureCollection"
	FeatureType           = "Feature"
)

type Properties struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Role     string `json:"role"`
	Distance int    `json:"distance"`
}

type GeometryMultiPoint struct {
	Type        string       `json:"type"`
	Coordinates [][2]float64 `json:"coordinates"`
}

type GeometryPoint struct {
	Type        string     `json:"type"`
	Coordinates [2]float64 `json:"coordinates"`
}

type Feature struct {
	Type       string      `json:"type"`
	Geometry   interface{} `json:"geometry"`
	Properties interface{} `json:"properties"`
	Id         int         `json:"id"`
}

type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type GeoJsonEncoder struct {
	writer io.Writer
}

type GeoJsonObject interface {
	GeoJsonFeatures() (features []Feature)
}

func NewFeatureCollection(features []Feature) FeatureCollection {
	return FeatureCollection{FeatureCollectionType, features}
}

//Pos converts Coord to GeoJSON Position object
func (c Coord) Pos() [2]float64 {
	lat, lon := GridToGeodetic(c.X, c.Y)
	return [2]float64{lon, lat}
}

//WriteJSON writes GetStartEndPointResult as a GeoJSON object
func (res GetStartEndPointResult) WriteJSON(w io.Writer) error {

	var features []Feature

	for _, p := range res.StartPoints {
		feature := Feature{"Feature", GeometryPoint{"Point", p.Pos()},
			Properties{p.Name, p.Type, "START", 0}, p.Id}

		features = append(features, feature)
	}

	for _, p := range res.EndPoints {
		feature := Feature{"Feature", GeometryPoint{"Point", p.Pos()},
			Properties{p.Name, p.Type, "END", 0}, p.Id}

		features = append(features, feature)
	}

	return json.NewEncoder(w).Encode(FeatureCollection{FeatureCollectionType, features})

}

//WriteJSON writes GetNearestStopAreaResult as a GeoJSON object
func (res GetNearestStopAreaResult) WriteJSON(w io.Writer) error {

	var features []Feature

	for _, p := range res.NearestStopAreas {

		feature := Feature{"Feature", GeometryPoint{"Point", p.Pos()},
			Properties{p.Name, "STOP_AREA", "NEARBY", p.Distance}, p.Id}

		features = append(features, feature)
	}

	return json.NewEncoder(w).Encode(FeatureCollection{FeatureCollectionType, features})
}

//WriteJSON writes GetJourneyPathResult as a GeoJSON object
func (res GetJourneyPathResult) WriteJSON(w io.Writer) error {

	var features []Feature

	parts, _ := res.Parts()

	for n, p := range parts {

		var coords [][2]float64
		for c := range p.Coords {
			if p.Coords[c].X > 0 && p.Coords[c].Y > 0 {
				coords = append(coords, p.Coords[c].Pos())
			}
		}

		start := Feature{"Feature", GeometryPoint{"Point", p.From.Pos()}, p.From, p.From.Id}
		end := Feature{"Feature", GeometryPoint{"Point", p.To.Pos()}, p.To, p.To.Id}

		p.Line.Distance = GridDistance(p.From.X, p.From.Y, p.To.X, p.To.Y)

		line := Feature{"Feature",
			GeometryMultiPoint{"LineString", coords}, p.Line, n}

		features = append(features, start)
		features = append(features, line)
		features = append(features, end)

	}

	return json.NewEncoder(w).Encode(FeatureCollection{FeatureCollectionType, features})
}

//WriteJSON writes GetDepartureArrivalResult as a JSON object
func (res GetDepartureArrivalResult) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(res.Lines)
}

//WriteJSON writes GetJourneyResult as a JSON object
func (res GetJourneyResult) WriteJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(res)
}
