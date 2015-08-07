package geo

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
