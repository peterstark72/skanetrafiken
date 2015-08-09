package openapi

import (
	"fmt"
	"testing"
)

func TestGridToGeodetic(t *testing.T) {

	lat, lon := GridToGeodetic(6158063, 1322703)

	if fmt.Sprintf("%.5f", lat) != "55.51992" || fmt.Sprintf("%.5f", lon) != "12.99795" {
		t.Error("Did not match")
	}

}

func TestGeodeticToGrid(t *testing.T) {

	x, y := GeodeticToGrid(55.519919, 12.997947)

	if fmt.Sprintf("%.0f", x) != "6158063" || fmt.Sprintf("%.0f", y) != "1322703" {
		t.Error("Did not match")
	}
}

func TestGridDistance(t *testing.T) {

	x, y := 6158063.0, 1322703.0

	d := GridDistance(x, y, x, y)

	if d != 0 {
		t.Error("Distance between same points should be zero!")
	}
}

func TestSphericalDistance(t *testing.T) {

	lat, lon := 55.519919, 12.997947

	d := GridDistance(lat, lon, lat, lon)

	if d != 0 {
		t.Error("Distance between same points should be zero!")
	}
}
