package main

import (
	"fmt"
	"github.com/peterstark72/golang-skanetrafiken/openapi"
	"os"
	"strconv"
	"strings"
	//"time"
)

func PrintPoints(points []openapi.Point) {
	for _, p := range points {
		lat, lon := openapi.GridToGeodetic(p.X, p.Y)
		fmt.Printf("%s,%d,%s,%f,%f\n",
			p.Name, p.Id, p.Type, lat, lon)
	}
}

func PrintNearestStopAreas(points []openapi.NearestStopArea) {
	for _, p := range points {
		lat, lon := openapi.GridToGeodetic(p.X, p.Y)
		fmt.Printf("%s,%d,%s,%f,%f,%d\n",
			p.Name, p.Id, "STOP_AREA", lat, lon, p.Distance)
	}
}

func SearchStation() {

	if len(os.Args) < 3 {
		fmt.Println("Try search <query>")
		return
	}

	q := os.Args[2]

	api := openapi.NewOpenAPI()

	result, err := api.QueryStation(q)
	if err != nil {
		fmt.Println(err)
		return
	}

	LetsPrint(Printables(result.StartPoints))

	//PrintPoints(result.StartPoints)
}

func SearchStartEndPoints() {

	if len(os.Args) < 4 {
		fmt.Println("Try points <start> <end>")
		return
	}

	start := os.Args[2]
	end := os.Args[3]

	api := openapi.NewOpenAPI()

	result, err := api.QueryPage(start, end)
	if err != nil {
		fmt.Println(err)
		return
	}

	PrintPoints(result.StartPoints)
	PrintPoints(result.EndPoints)

}

func SearchNearestStations() {

	if len(os.Args) < 3 {
		fmt.Println("Try nearest <lat>,<lon>")
		return
	}

	latlon := strings.Split(os.Args[2], ",")
	if len(latlon) != 2 {
		fmt.Println("Try nearest <lat>,<lon>")
		return
	}
	coords := [2]float64{}
	for i, c := range latlon {
		coords[i], _ = strconv.ParseFloat(c, 64)
	}

	x, y := openapi.GeodeticToGrid(coords[0], coords[1])

	api := openapi.NewOpenAPI()

	result, err := api.NearestStation(x, y, 1000)
	if err != nil {
		fmt.Println(err)
		return
	}

	PrintNearestStopAreas(result.NearestStopAreas)
}

func FindJourneys() {
	/*
		loc, _ := time.LoadLocation("Europe/Copenhagen")
		t := openapi.Point{"Hjärnarp Kyrkan", 92156, "STOP_AREA", openapi.Coord{}}
		f := openapi.Point{"Tygelsjö Laavägen", 80421, "STOP_AREA", openapi.Coord{6158063, 1322703}}

		result, _ := api.ResultsPage("next", f, t, time.Now().In(loc))

		path, _ := api.JourneyPath(result.JourneyResultKey, 0)
	*/
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Try skanetrafiken <command>")
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "search":
		SearchStation()
	case "points":
		SearchStartEndPoints()
	case "nearest":
		SearchNearestStations()
	}

}
