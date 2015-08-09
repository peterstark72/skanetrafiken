package main

import (
	"appengine"
	"appengine/urlfetch"
	"encoding/json"
	"github.com/peterstark72/golang-skanetrafiken/geo"
	"github.com/peterstark72/golang-skanetrafiken/openapi"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	DATETIME = "2006-01-02T15:04:05"
	LOCALE   = "Europe/Copenhagen"
)

func GetApiClient(req *http.Request) (api openapi.OpenApi) {

	c := appengine.NewContext(req)
	client := urlfetch.Client(c)

	api = openapi.NewOpenAPI()
	api.SetHTTPClient(client)
	return api
}

//WriteGeoJSONHeaders writes relevant HTTP Headers for GeoJSON
func WriteGeoJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-type", "application/vnd.geo+json")
}

//WriteJSONHeaders writes relevant HTTP Headers for GeoJSON
func WriteJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-type", "application/json")
}

//Returns departues for a given Station ID
func StationHandler(w http.ResponseWriter, req *http.Request) {

	var err error

	t := req.URL.Query().Get("t")

	//Get Locale
	loc, err := time.LoadLocation(LOCALE)
	if err != nil {
		http.Error(w, "Unknown location", 400)
		return
	}

	//Take date from Query OR default to today
	datetime, err := time.ParseInLocation(DATETIME, t, loc)
	if err != nil {
		datetime = time.Now().In(loc)
	}

	//Get the station ID
	segments := strings.Split(req.URL.Path, "/")
	resource := segments[len(segments)-1]
	stationId, err := strconv.ParseInt(resource, 0, 0)
	if err != nil {
		http.Error(w, "Missing ID", 400)
		return
	}

	api := GetApiClient(req)
	lines, err := api.StationResult(int(stationId), datetime)
	if err != nil {
		http.Error(w, "Could not load time table", 500)
		return
	}

	WriteJSONHeaders(w)
	json.NewEncoder(w).Encode(lines.Lines)
}

//SearchHandler searches for a station
func SearchStationHandler(w http.ResponseWriter, req *http.Request) {

	q := req.URL.Query().Get("q")

	if len(q) == 0 {
		http.Error(w, "Missing parameter q", 400)
		return
	}

	api := GetApiClient(req)

	result, err := api.QueryStation(q)
	if err != nil {
		http.Error(w, "Could not find matching stations", 500)
		return
	}

	WriteGeoJSONHeaders(w)
	result.WriteGeoJSON(w)

}

func SearchStartEndPointsHandler(w http.ResponseWriter, req *http.Request) {

	f := req.URL.Query().Get("from")
	t := req.URL.Query().Get("to")

	if len(f) == 0 || len(t) == 0 {
		http.Error(w, "Missing parameters", 400)
		return
	}

	api := GetApiClient(req)

	result, err := api.QueryPage(f, t)
	if err != nil {
		http.Error(w, "Could not find points", 500)
		return
	}

	WriteGeoJSONHeaders(w)
	result.WriteGeoJSON(w)
}

func NearestStationsHandler(w http.ResponseWriter, req *http.Request) {

	var lat, lon float64
	var err error

	path := req.URL.Path
	segments := strings.Split(path, "/")
	lastSegment := segments[len(segments)-1]
	latlon := strings.Split(lastSegment, ",")

	if len(latlon) != 2 {
		http.Error(w, "Coordinates should be <lat>,<lon>", 400)
		return
	}

	if lat, err = strconv.ParseFloat(latlon[0], 64); err != nil {
		http.Error(w, "Illegal coordinate", 400)
		return
	}
	if lon, err = strconv.ParseFloat(latlon[1], 64); err != nil {
		http.Error(w, "Illegal coordinate", 400)
		return
	}

	api := GetApiClient(req)

	x, y := geo.GeodeticToGrid(lat, lon)

	result, err := api.NearestStation(x, y, 500)
	if err != nil {
		http.Error(w, "Could not find stations nearby", 500)
		return
	}

	WriteGeoJSONHeaders(w)
	result.WriteGeoJSON(w)
}

/*

	/<name>,<id>,<type>/<name>,<id>,<type>

	Example:

	/Lund C,81216,0/Malmö C,80000,0

	LUNDASTIGEN 3 KÅGERÖD|280721|1

*/
func JourneysHandler(w http.ResponseWriter, req *http.Request) {

	loc, _ := time.LoadLocation(LOCALE)

	segments := strings.Split(req.URL.Path, "/")
	if len(segments) < 2 {
		http.Error(w, "Unknown path", 400)
		return
	}

	var points []openapi.Point
	for _, seg := range segments[len(segments)-2:] {
		p, err := openapi.NewPointFromURIParameter(seg)
		points = append(points, *p)
		if err != nil {
			http.Error(w, "Illegal Point", 500)
			return
		}
	}

	//t := openapi.Point{"Malmö C", 80000, "STOP_AREA", openapi.Coord{6167946, 1323245}}

	//t := openapi.Point{Name: "Lund C", Id: 81216, Type: "STOP_AREA"}
	//f := openapi.Point{Name: "Hjärnarp Kyrkan", Id: 92156, Type: "STOP_AREA"}
	//f := openapi.Point{Name: "Tygelsjö Laavägen", Id: 80421, Type: "STOP_AREA"}
	//t := openapi.Point{Name: "Höörs Station", Id: 67048, Type: "STOP_AREA"}

	api := GetApiClient(req)

	result, err := api.ResultsPage("next", points[0], points[1], time.Now().In(loc))
	if err != nil {
		http.Error(w, "Could not find journey", 500)
		return
	}

	WriteJSONHeaders(w)
	json.NewEncoder(w).Encode(result)
}

func JourneyPathsHandler(w http.ResponseWriter, req *http.Request) {

	segments := strings.Split(req.URL.Path, "/")
	if len(segments) < 2 {
		http.Error(w, "Missing Journey Key", 400)
		return
	}

	key := segments[len(segments)-2]
	seq, err := strconv.ParseInt(segments[len(segments)-1], 0, 0)
	if err != nil {
		http.Error(w, "Sequence must be an integer", 400)
		return
	}

	api := GetApiClient(req)

	result, err := api.JourneyPath(key, int(seq))
	if err != nil {
		http.Error(w, "Could not find journey path", 500)
		return
	}

	WriteGeoJSONHeaders(w)
	result.WriteGeoJSON(w)
}

func init() {
	http.HandleFunc("/journeypaths/", JourneyPathsHandler)
	http.HandleFunc("/journeys/", JourneysHandler)

	http.HandleFunc("/search", SearchStartEndPointsHandler)

	http.HandleFunc("/stations", SearchStationHandler)
	http.HandleFunc("/stations/", StationHandler)

	http.HandleFunc("/nearby/", NearestStationsHandler)

}
