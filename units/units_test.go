package units

import (
	"math"
	"testing"
)

func TestAngle_Conversions(t *testing.T) {
	a := AngleFromDegrees(180.0)
	if math.Abs(a.Radians()-math.Pi) > 1e-15 {
		t.Errorf("180° in radians: got %f, want π", a.Radians())
	}
	if math.Abs(a.Degrees()-180.0) > 1e-12 {
		t.Errorf("180° in degrees: got %f", a.Degrees())
	}
	if math.Abs(a.Hours()-12.0) > 1e-12 {
		t.Errorf("180° in hours: got %f, want 12", a.Hours())
	}
	if math.Abs(a.Arcminutes()-10800.0) > 1e-8 {
		t.Errorf("180° in arcminutes: got %f", a.Arcminutes())
	}
	if math.Abs(a.Arcseconds()-648000.0) > 1e-6 {
		t.Errorf("180° in arcseconds: got %f", a.Arcseconds())
	}
}

func TestAngle_FromHours(t *testing.T) {
	a := AngleFromHours(6.0)
	if math.Abs(a.Degrees()-90.0) > 1e-12 {
		t.Errorf("6h in degrees: got %f, want 90", a.Degrees())
	}
}

func TestAngle_FromRadians(t *testing.T) {
	a := NewAngle(math.Pi / 2)
	if math.Abs(a.Degrees()-90.0) > 1e-12 {
		t.Errorf("π/2 in degrees: got %f, want 90", a.Degrees())
	}
}

func TestAngle_DMS(t *testing.T) {
	a := AngleFromDegrees(41.0 + 30.0/60.0 + 15.5/3600.0)
	sign, deg, min, sec := a.DMS()
	if sign != 1.0 || deg != 41 || min != 30 || math.Abs(sec-15.5) > 0.01 {
		t.Errorf("DMS: got sign=%f d=%d m=%d s=%f, want +41°30'15.5\"", sign, deg, min, sec)
	}
}

func TestAngle_DMS_Negative(t *testing.T) {
	a := AngleFromDegrees(-29.5)
	sign, deg, min, sec := a.DMS()
	if sign != -1.0 || deg != 29 || min != 30 || sec > 0.01 {
		t.Errorf("DMS negative: got sign=%f d=%d m=%d s=%f, want -29°30'0\"", sign, deg, min, sec)
	}
}

func TestAngle_HMS(t *testing.T) {
	a := AngleFromHours(17.0 + 45.0/60.0 + 40.0/3600.0)
	sign, h, m, s := a.HMS()
	if sign != 1.0 || h != 17 || m != 45 || math.Abs(s-40.0) > 0.01 {
		t.Errorf("HMS: got sign=%f h=%d m=%d s=%f, want 17h45m40s", sign, h, m, s)
	}
}

func TestAngle_Zero(t *testing.T) {
	a := NewAngle(0)
	if a.Degrees() != 0 || a.Hours() != 0 || a.Radians() != 0 {
		t.Error("zero angle should be zero in all units")
	}
}

func TestDistance_Conversions(t *testing.T) {
	d := NewDistance(149597870.7)
	if math.Abs(d.AU()-1.0) > 1e-12 {
		t.Errorf("1 AU in AU: got %f", d.AU())
	}
	if math.Abs(d.M()-149597870700.0) > 1.0 {
		t.Errorf("1 AU in meters: got %f", d.M())
	}
}

func TestDistance_FromAU(t *testing.T) {
	d := DistanceFromAU(1.0)
	if math.Abs(d.Km()-AUToKm) > 1e-6 {
		t.Errorf("1 AU in km: got %f, want %f", d.Km(), AUToKm)
	}
}

func TestDistance_FromMeters(t *testing.T) {
	d := DistanceFromMeters(1000.0)
	if math.Abs(d.Km()-1.0) > 1e-15 {
		t.Errorf("1000m in km: got %f", d.Km())
	}
}

func TestDistance_LightSeconds(t *testing.T) {
	d := NewDistance(299792.458)
	if math.Abs(d.LightSeconds()-1.0) > 1e-12 {
		t.Errorf("c km in light-seconds: got %f, want 1.0", d.LightSeconds())
	}
}

func TestDistance_AU_Roundtrip(t *testing.T) {
	d := DistanceFromAU(5.2)
	if math.Abs(d.AU()-5.2) > 1e-12 {
		t.Errorf("roundtrip AU: got %f, want 5.2", d.AU())
	}
}
