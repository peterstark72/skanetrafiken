package main

import (
	"fmt"
	"github.com/peterstark72/golang-skanetrafiken/openapi"
)

func main() {

	api := openapi.NewOpenAPI()

	stations, err := api.QueryStation("Malmö")
	if err != nil {
		return
	}

	for _, station := range stations.StartPoints {
		fmt.Printf("%s, %d\n", station.Name, station.Id)
	}

}
