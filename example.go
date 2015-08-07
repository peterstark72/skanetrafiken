package main

import (
	"fmt"
	"github.com/peterstark72/golang-skanetrafiken/openapi"
)

func main() {

	api := openapi.NewOpenAPI()

	result, err := api.QueryStation("Malmö")
	if err != nil {
		return
	}

}
