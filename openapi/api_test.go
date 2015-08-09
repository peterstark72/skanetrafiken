package openapi

import (
	"testing"
	"time"
)

var api = NewOpenAPI()

func TestQueryPage(t *testing.T) {

	_, err := api.QueryPage("Lund", "Ystad")
	if err != nil {
		t.Error(err)
	}

}

func TestResultsPage(t *testing.T) {

	_, err := api.ResultsPage("next",
		Point{"Malmö C", 80000, "STOP_AREA", Coord{0, 0}},
		Point{"Landskrona", 82000, "STOP_AREA", Coord{0, 0}},
		time.Now())
	if err != nil {
		t.Error(err)
	}
}

func TestQueryStation(t *testing.T) {

	_, err := api.QueryStation("Malmö")
	if err != nil {
		t.Error(err)
	}
}

func TestNearestStopAreas(t *testing.T) {

	_, err := api.NearestStation(6167930, 1323215, 1000)
	if err != nil {
		t.Error(err)
	}

}

func TestStationResult(t *testing.T) {

	_, err := api.StationResult(80000, time.Now())
	if err != nil {
		t.Error(err)
	}
}

func TestJourneyPath(t *testing.T) {

	var err error

	respage, err := api.ResultsPage("next", Point{"Malmö C", 80000, "STOP_AREA", Coord{0, 0}}, Point{"Landskrona", 82000, "STOP_AREA", Coord{0, 0}}, time.Now())

	path, err := api.JourneyPath(respage.JourneyResultKey, 0)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = path.Parts()
	if err != nil {
		t.Error(err)
		return
	}

}
