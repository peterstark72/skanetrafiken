/*
Package skanetrafiken wraps the Open API documented here:
http://labs.skanetrafiken.se/api.asp


Example usage:

api := NewOpenAPI()

res, err := api.QueryStation("Malmö")

for n, point := range res.StartPoints:
	fmt.Println(n, point.Name)

*/
package openapi

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const BaseURL = "http://www.labs.skanetrafiken.se/v2.2/"

const (
	QUERYPAGE      = "querypage.asp"
	RESULTSPAGE    = "resultspage.asp"
	QUERYSTATION   = "querystation.asp"
	JOURNEYPATH    = "journeypath.asp"
	NEARESTSTATION = "neareststation.asp"
	STATIONRESULT  = "stationresults.asp"
)

const (
	YYMMDD   = "060102"
	HHMM     = "1505"
	DATETIME = "2006-01-02 15:04"
)

type OpenApi struct {
	Client *http.Client
}

//NewOpenAPI creates a new instance of the OpenAPI
func NewOpenAPI() OpenApi {
	api := OpenApi{Client: new(http.Client)}
	return api
}

/*
SetHTTPClient sets an http.Client.

This is useful for the Google Appengine where we cannot use
the built in http.Client. Instead we use appengine's.
*/
func (api *OpenApi) SetHTTPClient(c *http.Client) {
	api.Client = c
}

/*
ParseDateTime parses the DateTime formatted strings
from the API and returns a proper Time object.
*/
func ParseDateTime(dt string) (time.Time, error) {
	loc, _ := time.LoadLocation("Europe/Copenhagen")
	return time.ParseInLocation(DATETIME, dt, loc)
}

type PointOnRouteLink struct {
	Id               int
	Name             string
	StopPoint        string
	ArrDateTime      string
	ArrIsTimingPoint bool
}

type RealTimeInfo struct {
	NewDepPoint        string
	NewArrPoint        string
	DepTimeDeviation   int
	DepDeviationAffect RealTimeAffect
	ArrTimeDeviation   int
	ArrDeviationAffect RealTimeAffect
	Canceled           bool
}

type Line struct {
	Name              string
	No                int
	RunNo             int
	LineTypeId        int
	LineTypeName      string
	TransportModeId   int
	JourneyDateTime   string
	TransportModeName string
	Towards           string
	TrainNo           int
	OperatorId        int
	OperatorName      string
	RealTime          RealTimeInfo
	PointsOnRouteLink []PointOnRouteLink `xml:"PointsOnRouteLink>PointOnRouteLink"`
}

const (
	STOP_AREA = iota
	ADDRESS
	POI
	UNKNOWN
)

//Coord is RT90 coordinates
type Coord struct {
	X float64
	Y float64
}

type Point struct {
	Name string
	Id   int
	Type string
	Coord
}

//AsURIParameter returns the point in the format used in URI query
func (p Point) AsURIParameter() string {
	var PointTypes = map[string]int{
		"STOP_AREA": STOP_AREA,
		"POI":       POI,
		"ADDRESS":   ADDRESS,
		"UNKNOWN":   UNKNOWN,
	}

	return fmt.Sprintf("%s|%d|%d", p.Name, p.Id, PointTypes[p.Type])
}

type RealTimeAffect string

type RouteLink struct {
	RouteLinkKey string
	DepDateTime  string
	ArrDateTime  string
	From         Point
	To           Point
	RealTime     RealTimeInfo
	Line         Line
}

type Journey struct {
	SequenceNo  int
	DepDateTime string
	ArrDateTime string
	DepWalkDist int
	ArrWalkDist int
	NoOfChanges int
	JourneyKey  string
	Guaranteed  bool
	CO2Factor   int
	RouteLinks  []RouteLink `xml:"RouteLinks>RouteLink"`
}

type Status struct {
	Code    int
	Message string
}

type GetStartEndPointResult struct {
	StartPoints []Point `xml:"StartPoints>Point"`
	EndPoints   []Point `xml:"EndPoints>Point"`
	ViaPoints   []Point `xml:"ViaPoints>Point"`
	Status
}

type Part struct {
	XMLName xml.Name `xml:"Part"`
	From    Point
	To      Point
	Line    Line
	Coords  []Coord `xml:"Coords>Coord"`
}

type GetStartEndPointResponse struct {
	GetStartEndPointResult GetStartEndPointResult
}

type GetJourneyResult struct {
	JourneyResultKey string
	Journeys         []Journey `xml:"Journeys>Journey"`
	Distance         int
	CO2value         float64
	Status
}

type GetJourneyResponse struct {
	GetJourneyResult GetJourneyResult
}

type GetJourneyPathResult struct {
	Status
	ResultXML []byte
}

type GetJourneyPathResponse struct {
	GetJourneyPathResult GetJourneyPathResult
}

type NearestStopArea struct {
	Point
	Distance int
}

type GetNearestStopAreaResult struct {
	Status
	NearestStopAreas []NearestStopArea `xml:"NearestStopAreas>NearestStopArea"`
}

type GetNearestStopAreaResponse struct {
	GetNearestStopAreaResult GetNearestStopAreaResult
}

type StopAreaData struct {
	Name string
	X    float64
	Y    float64
}

type GetDepartureArrivalResult struct {
	Status
	Lines        []Line `xml:"Lines>Line"`
	StopAreaData StopAreaData
}

type GetDepartureArrivalResponse struct {
	GetDepartureArrivalResult GetDepartureArrivalResult
}

type SOAPBody struct {
	GetStartEndPointResponse    GetStartEndPointResponse
	GetJourneyResponse          GetJourneyResponse
	GetJourneyPathResponse      GetJourneyPathResponse
	GetNearestStopAreaResponse  GetNearestStopAreaResponse
	GetDepartureArrivalResponse GetDepartureArrivalResponse
}

type SOAPEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    SOAPBody `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

//get loads SOAP Envelope from the endpoint and stores it in the body parameter.
func (api OpenApi) get(endpoint string, params url.Values, body interface{}) error {

	var err error

	url := BaseURL + endpoint + "?" + params.Encode()

	res, err := api.Client.Get(url)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}

	return xml.Unmarshal([]byte(data), &body)

}

//QueryStation returns stations with matching names
func (api OpenApi) QueryStation(inpPointFr string) (res GetStartEndPointResult, err error) {

	params := url.Values{}
	params.Set("inpPointFr", inpPointFr)

	soap := SOAPEnvelope{}
	if err = api.get(QUERYSTATION, params, &soap); err != nil {
		return res, err
	}

	return soap.Body.GetStartEndPointResponse.GetStartEndPointResult, nil
}

//QueryPage returns matching start/end stations
func (api OpenApi) QueryPage(inpPointFr, inpPointTo string) (res GetStartEndPointResult, err error) {

	params := url.Values{}
	params.Set("inpPointFr", inpPointFr)
	params.Set("inpPointTo", inpPointTo)

	soap := SOAPEnvelope{}
	if err = api.get(QUERYPAGE, params, &soap); err != nil {
		return res, err
	}

	return soap.Body.GetStartEndPointResponse.GetStartEndPointResult, nil
}

//ResultsPage returns list of journeys between two points
func (api OpenApi) ResultsPage(cmdaction string, from, to Point, LastStart time.Time) (res GetJourneyResult, err error) {

	params := url.Values{}
	params.Set("cmdaction", cmdaction)
	params.Set("selPointFr", from.AsURIParameter())
	params.Set("selPointTo", to.AsURIParameter())
	params.Set("LastStart", LastStart.Format(DATETIME))
	params.Set("DetailedResult", "True")

	soap := SOAPEnvelope{}
	if err = api.get(RESULTSPAGE, params, &soap); err != nil {
		return res, err
	}
	return soap.Body.GetJourneyResponse.GetJourneyResult, nil
}

//NearestStation returns stations nearby X,Y point, within radius R
func (api OpenApi) NearestStation(x, y float64, R int) (res GetNearestStopAreaResult, err error) {

	params := url.Values{}
	params.Set("x", fmt.Sprintf("%.0f", x))
	params.Set("y", fmt.Sprintf("%.0f", y))
	params.Set("R", fmt.Sprintf("%d", R))

	soap := SOAPEnvelope{}
	if err = api.get(NEARESTSTATION, params, &soap); err != nil {
		return res, err
	}
	return soap.Body.GetNearestStopAreaResponse.GetNearestStopAreaResult, nil
}

//StationResult returns timetable for a given station
func (api OpenApi) StationResult(selPointFrKey int, t time.Time) (res GetDepartureArrivalResult, err error) {

	params := url.Values{}
	params.Set("selPointFrKey", fmt.Sprintf("%d", selPointFrKey))
	params.Set("inpDate", t.Format(YYMMDD))
	params.Set("inpTime", t.Format(HHMM))

	soap := SOAPEnvelope{}
	if err = api.get(STATIONRESULT, params, &soap); err != nil {
		return res, err
	}
	return soap.Body.GetDepartureArrivalResponse.GetDepartureArrivalResult, nil
}

//JourneyPath returns geo path for a given JourneyResultKey and sequence number
func (api OpenApi) JourneyPath(cf string, sequenceNo int) (res GetJourneyPathResult, err error) {

	params := url.Values{}
	params.Set("cf", cf)
	params.Set("id", fmt.Sprintf("%d", sequenceNo))

	soap := SOAPEnvelope{}
	if err = api.get(JOURNEYPATH, params, &soap); err != nil {
		return res, err
	}

	return soap.Body.GetJourneyPathResponse.GetJourneyPathResult, nil
}

//Part returns the geo coordinates
func (res GetJourneyPathResult) Part() (part *Part, err error) {

	err = xml.Unmarshal(res.ResultXML, &part)
	if err != nil {
		return part, err
	}

	return part, nil
}
