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

//GetResourceIdFromPath return the ID from /some/path/{id}
func GetResourceIdFromPath(path string) (int, error) {

	segments := strings.Split(path, "/")

	id, err := strconv.ParseInt(segments[len(segments)-1], 0, 0)
	if err != nil {
		return 0, err
	}
	return int(id), nil
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
	stationId, err := GetResourceIdFromPath(req.URL.Path)
	if err != nil {
		http.Error(w, "Missing ID", 400)
		return
	}

	api := GetApiClient(req)
	lines, err := api.StationResult(stationId, datetime)
	if err != nil {
		http.Error(w, "Could not load time table", 500)
		return
	}

	json.NewEncoder(w).Encode(lines.Lines)
}

func SearchHandler(w http.ResponseWriter, req *http.Request) {

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

	w.Header().Set("Access-Control-Allow-Origin", "*")

	json.NewEncoder(w).Encode(result.AsFeatureCollection())

}

func SearchPointsHandler(w http.ResponseWriter, req *http.Request) {

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

	json.NewEncoder(w).Encode(result.AsFeatureCollection())
}

func NearestStationsHandler(w http.ResponseWriter, req *http.Request) {

	type Stop struct {
		Name     string
		Id       int
		Distance int
		Lat      float64
		Lon      float64
	}

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

	result, err := api.NearestStation(x, y, 1000)
	if err != nil {
		http.Error(w, "Could not find stations nearby", 500)
		return
	}

	json.NewEncoder(w).Encode(result.AsFeatureCollection())

}

func init() {
	http.HandleFunc("/points", SearchPointsHandler)
	http.HandleFunc("/stations", SearchHandler)
	http.HandleFunc("/stations/", StationHandler)
	http.HandleFunc("/nearby/", NearestStationsHandler)
	http.HandleFunc("/", SearchHandler)
}
