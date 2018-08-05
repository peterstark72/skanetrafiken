/*
Golang wrapper of Skanetrafiken Open API, as documented here:

http://labs.skanetrafiken.se/api.asp

Method names keep the names from the API as much as possible, so for example "/querypage.asp" is QueryPage() etc.


Example usage:

api := NewOpenAPI()

res, err := api.QueryStation("Malmö")

for n, point := range res.StartPoints:
	fmt.Println(n, point.Name)

*/
package openapi

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

type OpenApi struct {
	client *http.Client
}

var DefaultClient = &OpenApi{}

func (api OpenApi) transport() *http.Client {
	if api.client != nil {
		return api.client
	}
	return new(http.Client)
}

//NewOpenAPI creates a new instance of the OpenAPI
func NewOpenAPI() OpenApi {
	api := OpenApi{new(http.Client)}
	return api
}

/*
SetHTTPClient sets an http.Client.

This is useful for the Google Appengine where we cannot use
the built in http.Client. Instead we use appengine's.
*/
func (api *OpenApi) SetHTTPClient(c *http.Client) {
	api.client = c
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
	DepDeviationAffect string
	ArrTimeDeviation   int
	ArrDeviationAffect string
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

/*
NewPointFromURIParameter creates a new Point from the URI parameter format

<name>|<id>|<type>

Where <type> is one of "STOP_AREA", "POI", etc.
*/
func NewPointFromURIParameter(s string) (*Point, error) {
	parts := strings.Split(s, "|")
	if len(parts) != 3 {
		return nil, errors.New("Incorrect Point parameters")
	}
	id, _ := strconv.ParseInt(parts[1], 0, 0)
	return &Point{parts[0], int(id), parts[2], Coord{}}, nil
}

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

type PartLine struct {
	Name     string
	No       int
	LinTName string
	Distance int //This is not in the Open API
}

type PartPoint struct {
	Id       int
	Poi      string
	PoiAlias string
	Name     string
	Coord
}

type Part struct {
	Line   PartLine  `xml:"Line"`
	To     PartPoint `xml:"To"`
	From   PartPoint `xml:"From"`
	Coords []Coord   `xml:"Coords>Coord"`
}

type ResultXML struct {
	XMLName xml.Name `xml:"Parts"`
	Parts   []Part   `xml:"Part"`
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
	Coord
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

	res, err := api.transport().Get(url)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
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

//QueryPage returns matching start/end points
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
	params.Set("LastStart", LastStart.Format("2006-01-02 15:04"))
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
	params.Set("inpDate", t.Format("060102"))
	params.Set("inpTime", t.Format("1504"))

	soap := SOAPEnvelope{}
	if err = api.get(STATIONRESULT, params, &soap); err != nil {
		return res, err
	}
	return soap.Body.GetDepartureArrivalResponse.GetDepartureArrivalResult, nil
}

//GetStationResult returns timetable for a given station
func GetStationResult(stationID int, t time.Time) (res GetDepartureArrivalResult, err error) {
	return DefaultClient.StationResult(stationID, t)
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

//Parts unmarshals the raw XML included in GetJourneyPathResult
func (res GetJourneyPathResult) Parts() (parts []Part, err error) {

	//We need to wrap the raw XML with <Parts> to make it well formed [sigh]
	data := []byte("<Parts>")
	data = append(data, []byte(res.ResultXML)...)
	data = append(data, []byte("</Parts>")...)

	r := ResultXML{}
	err = xml.Unmarshal(data, &r)
	if err != nil {
		return nil, err
	}

	return r.Parts, nil
}
