# Golang Skanetrafiken

Go package to access traffic data from Skanetrafiken. 

1. Open API -- Wrapper for the public Open API
2. Geo -- Functions to convert between RT90 and WGS84
3. Appengine -- A simple API Fasade, running on Google Appengine, that serves mostly GeoJSON instead of the original structures.  


## Open API

The Skanetrafiken Open API is documented at http://labs.skanetrafiken.se/api.asp.

It provides access to the following traffic-related resources: 

* Search stations by name, e.g. "Malmö", "Eslöv".
* Get line departures (i.e. timetable) for a given Station. 
* Get suggested journeys from A to B at any given time. 
* Get list of stations near a given geographical point.
* Get geographical journey path for a given journey. 


Method names follow the names used in the documentation, "/querypage.asp" is QueryPage() etc. 


First, create an API instance:

```Go
	api := openapi.NewOpenAPI()
```

Then, make a query: 

```Go
	stations, err := api.QueryStation("Malmö")
	if err != nil {
		return
	}
```

And then, print the results:

```Go
	for _, station := range stations.StartPoints {
		fmt.Printf("%s, %d\n", station.Name, station.Id)
	}
```



