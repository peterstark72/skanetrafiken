/*
Converts coordindates between WGS84 and RT90 (rt90_2.5_gon_v).

From RT90 to WGS84:
	GridToGeodetic()

From WGS84 to RT90:
	GeodeticToGrid

*/
package geo

import (
	"math"
)

// Parameters for rt90_2.5_gon_v
const (
	CentralMeridian = 15.0 + 48.0/60.0 + 22.624306/3600.0
	Scale           = 1.00000561024
	FalseNorthing   = -667.711
	FalseEasting    = 1500064.274
)

//GRS80 Defaults
const (
	Axis       = 6378137.0           // GRS 80.
	Flattening = 1.0 / 298.257222101 // GRS 80.
	//CentralMeridian = 31337.0
)

//GridToGeodetic converts RT90 coordinates to WGS84
func GridToGeodetic(x, y float64) [2]float64 {

	if CentralMeridian == 31337.0 {
		return [2]float64{0.0, 0.0}
	}

	e2 := Flattening * (2.0 - Flattening)
	n := Flattening / (2.0 - Flattening)
	a_roof := Axis / (1.0 + n) * (1.0 + n*n/4.0 + n*n*n*n/64.0)
	delta1 := n/2.0 - 2.0*n*n/3.0 + 37.0*n*n*n/96.0 - n*n*n*n/360.0
	delta2 := n*n/48.0 + n*n*n/15.0 - 437.0*n*n*n*n/1440.0
	delta3 := 17.0*n*n*n/480.0 - 37*n*n*n*n/840.0
	delta4 := 4397.0 * n * n * n * n / 161280.0

	Astar := e2 + e2*e2 + e2*e2*e2 + e2*e2*e2*e2
	Bstar := -(7.0*e2*e2 + 17.0*e2*e2*e2 + 30.0*e2*e2*e2*e2) / 6.0
	Cstar := (224.0*e2*e2*e2 + 889.0*e2*e2*e2*e2) / 120.0
	Dstar := -(4279.0 * e2 * e2 * e2 * e2) / 1260.0

	DegToRad := math.Pi / 180
	LambdaZero := CentralMeridian * DegToRad
	xi := (x - FalseNorthing) / (Scale * a_roof)
	eta := (y - FalseEasting) / (Scale * a_roof)
	xi_prim := xi - delta1*math.Sin(2.0*xi)*math.Cosh(2.0*eta) - delta2*math.Sin(4.0*xi)*math.Cosh(4.0*eta) - delta3*math.Sin(6.0*xi)*math.Cosh(6.0*eta) - delta4*math.Sin(8.0*xi)*math.Cosh(8.0*eta)
	eta_prim := eta - delta1*math.Cos(2.0*xi)*math.Sinh(2.0*eta) - delta2*math.Cos(4.0*xi)*math.Sinh(4.0*eta) - delta3*math.Cos(6.0*xi)*math.Sinh(6.0*eta) - delta4*math.Cos(8.0*xi)*math.Sinh(8.0*eta)
	phi_star := math.Asin(math.Sin(xi_prim) / math.Cosh(eta_prim))
	delta_lambda := math.Atan(math.Sinh(eta_prim) / math.Cos(xi_prim))
	lon_radian := LambdaZero + delta_lambda
	lat_radian := phi_star + math.Sin(phi_star)*math.Cos(phi_star)*(Astar+Bstar*math.Pow(math.Sin(phi_star), 2)+Cstar*math.Pow(math.Sin(phi_star), 4)+Dstar*math.Pow(math.Sin(phi_star), 6))

	return [2]float64{lat_radian * 180.0 / math.Pi, lon_radian * 180.0 / math.Pi}
}

//GeodeticToGrid converts WGS84 coordinates to RT90
func GeodeticToGrid(lat, lon float64) (x, y float64) {

	// Prepare ellipsoid-based stuff.
	e2 := Flattening * (2.0 - Flattening)
	n := Flattening / (2.0 - Flattening)
	a_roof := Axis / (1.0 + n) * (1.0 + n*n/4.0 + n*n*n*n/64.0)
	A := e2
	B := (5.0*e2*e2 - e2*e2*e2) / 6.0
	C := (104.0*e2*e2*e2 - 45.0*e2*e2*e2*e2) / 120.0
	D := (1237.0 * e2 * e2 * e2 * e2) / 1260.0
	beta1 := n/2.0 - 2.0*n*n/3.0 + 5.0*n*n*n/16.0 + 41.0*n*n*n*n/180.0
	beta2 := 13.0*n*n/48.0 - 3.0*n*n*n/5.0 + 557.0*n*n*n*n/1440.0
	beta3 := 61.0*n*n*n/240.0 - 103.0*n*n*n*n/140.0
	beta4 := 49561.0 * n * n * n * n / 161280.0

	// Convert.
	DegToRad := math.Pi / 180.0
	phi := lat * DegToRad
	lambd := lon * DegToRad
	lambda_zero := CentralMeridian * DegToRad

	phi_star := phi - math.Sin(phi)*math.Cos(phi)*(A+
		B*math.Pow(math.Sin(phi), 2)+
		C*math.Pow(math.Sin(phi), 4)+
		D*math.Pow(math.Sin(phi), 6))
	delta_lambda := lambd - lambda_zero
	xi_prim := math.Atan(math.Tan(phi_star) / math.Cos(delta_lambda))
	eta_prim := math.Atanh(math.Cos(phi_star) * math.Sin(delta_lambda))
	x = Scale*a_roof*(xi_prim+
		beta1*math.Sin(2.0*xi_prim)*math.Cosh(2.0*eta_prim)+
		beta2*math.Sin(4.0*xi_prim)*math.Cosh(4.0*eta_prim)+
		beta3*math.Sin(6.0*xi_prim)*math.Cosh(6.0*eta_prim)+
		beta4*math.Sin(8.0*xi_prim)*math.Cosh(8.0*eta_prim)) +
		FalseNorthing
	y = Scale*a_roof*(eta_prim+
		beta1*math.Cos(2.0*xi_prim)*math.Sinh(2.0*eta_prim)+
		beta2*math.Cos(4.0*xi_prim)*math.Sinh(4.0*eta_prim)+
		beta3*math.Cos(6.0*xi_prim)*math.Sinh(6.0*eta_prim)+
		beta4*math.Cos(8.0*xi_prim)*math.Sinh(8.0*eta_prim)) +
		FalseEasting
	return x, y
}
