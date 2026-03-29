package taxmath

import (
	"math"
	"testing"
)

func assertNear(t *testing.T, got, want float64, label string) {
	t.Helper()
	if math.Abs(got-want) > 0.01 {
		t.Errorf("%s: got %.2f, want %.2f (diff %.2f)", label, got, want, got-want)
	}
}

func TestFEIELimit(t *testing.T) {
	tests := []struct {
		year int
		want float64
	}{
		{2024, 126500},
		{2025, 130000},
		{2026, 133600},
		{2023, 0}, // not defined
	}
	for _, tt := range tests {
		got := FEIELimit(tt.year)
		if got != tt.want {
			t.Errorf("FEIELimit(%d) = %.0f, want %.0f", tt.year, got, tt.want)
		}
	}
}

func TestHousingBaseAmount(t *testing.T) {
	// 16% of $130,000 = $20,800; 16% of $126,500 = $20,240
	assertNear(t, HousingBaseAmount(2025), 20800, "HousingBaseAmount(2025)")
	assertNear(t, HousingBaseAmount(2024), 20240, "HousingBaseAmount(2024)")
}

func TestHousingMaxAmount(t *testing.T) {
	// 30% of $130,000 = $39,000
	assertNear(t, HousingMaxAmount(2025), 39000, "HousingMaxAmount(2025)")
}

func TestProrateExclusion(t *testing.T) {
	tests := []struct {
		name           string
		limit          float64
		qualifyingDays int
		totalDays      int
		want           float64
	}{
		{"full year", 130000, 365, 365, 130000},
		{"330 days PPT", 130000, 330, 365, 130000 * 330.0 / 365.0},
		{"half year", 130000, 183, 365, 130000 * 183.0 / 365.0},
		{"zero days", 130000, 0, 365, 0},
		{"zero total", 130000, 100, 0, 0},
		{"more than total", 130000, 400, 365, 130000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProrateExclusion(tt.limit, tt.qualifyingDays, tt.totalDays)
			assertNear(t, got, tt.want, tt.name)
		})
	}
}

func TestComputeTaxWithStacking(t *testing.T) {
	// Single filer, 2025. $60k taxable + $100k excluded.
	// Without stacking: tax on $60k = $8,114
	// With stacking: tax($160k) - tax($100k) = much higher because
	// the $60k is taxed at higher brackets (22%-24% range instead of 10%-22%)
	taxNormal := ComputeTax(60000, Single, 2025, Federal)
	taxStacked := ComputeTaxWithStacking(60000, 100000, Single, 2025, Federal)

	if taxStacked <= taxNormal {
		t.Errorf("stacked tax (%.2f) should be > normal tax (%.2f)", taxStacked, taxNormal)
	}

	// Verify: tax($160k) - tax($100k)
	taxOn160k := ComputeTax(160000, Single, 2025, Federal)
	taxOn100k := ComputeTax(100000, Single, 2025, Federal)
	expected := taxOn160k - taxOn100k
	assertNear(t, taxStacked, expected, "stacking formula")

	// Zero excluded income should equal normal tax
	taxNoExclusion := ComputeTaxWithStacking(60000, 0, Single, 2025, Federal)
	assertNear(t, taxNoExclusion, taxNormal, "no exclusion = normal tax")

	// Zero taxable income
	taxZero := ComputeTaxWithStacking(0, 100000, Single, 2025, Federal)
	assertNear(t, taxZero, 0, "zero taxable income")
}

func TestComputeTaxWithStackingMFJ(t *testing.T) {
	// MFJ filer with $80k taxable and $130k excluded
	taxStacked := ComputeTaxWithStacking(80000, 130000, MarriedFilingJoint, 2025, Federal)
	taxOn210k := ComputeTax(210000, MarriedFilingJoint, 2025, Federal)
	taxOn130k := ComputeTax(130000, MarriedFilingJoint, 2025, Federal)
	expected := taxOn210k - taxOn130k
	assertNear(t, taxStacked, expected, "MFJ stacking formula")
}
