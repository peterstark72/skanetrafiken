# Golang Skanetrafiken

Go package to access the Open API from Skanetrafiken. 


## Open API

The Skanetrafiken Open API is documented at http://labs.skanetrafiken.se/api.asp.

It provides access to the following traffic-related resources: 

* Search stations by name, e.g. "Malmö", "Eslöv".
* Get line departures (i.e. timetable) for a given Station. 
* Get suggested journeys from A to B at any given time. 
* Get list of stations near a given geographical point.
* Get geographical journey path for a given journey. 


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



