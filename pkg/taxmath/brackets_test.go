package taxmath

import (
	"math"
	"testing"
)

const tolerance = 0.01 // Allow rounding to nearest cent

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < tolerance
}

// ---------------------------------------------------------------------------
// Federal bracket tests
// ---------------------------------------------------------------------------

func TestFederalSingle75k(t *testing.T) {
	// 10% on 11925 = 1192.50
	// 12% on (48475-11925)=36550 = 4386.00
	// 22% on (75000-48475)=26525 = 5835.50
	// Total = 11414.00
	tax := ComputeTax(75000, Single, 2025, Federal)
	expected := 11414.00
	if !approxEqual(tax, expected) {
		t.Errorf("Federal single $75k: got %.2f, want %.2f", tax, expected)
	}
}

func TestFederalMFJ150k(t *testing.T) {
	// 10% on 23850 = 2385.00
	// 12% on (96950-23850)=73100 = 8772.00
	// 22% on (150000-96950)=53050 = 11671.00
	// Total = 22828.00
	tax := ComputeTax(150000, MarriedFilingJoint, 2025, Federal)
	expected := 22828.00
	if !approxEqual(tax, expected) {
		t.Errorf("Federal MFJ $150k: got %.2f, want %.2f", tax, expected)
	}
}

// ---------------------------------------------------------------------------
// California bracket tests
// ---------------------------------------------------------------------------

func TestCASingle75k(t *testing.T) {
	// 1% on 10756 = 107.56
	// 2% on (25499-10756)=14743 = 294.86
	// 4% on (40245-25499)=14746 = 589.84
	// 6% on (55866-40245)=15621 = 937.26
	// 8% on (70612-55866)=14746 = 1179.68
	// 9.3% on (75000-70612)=4388 = 408.084
	// Total = 3517.264 (no mental health surcharge)
	tax := ComputeTax(75000, Single, 2025, StateCA)
	expected := 3517.284
	if !approxEqual(tax, expected) {
		t.Errorf("CA single $75k: got %.4f, want %.4f", tax, expected)
	}
}

func TestCAHighIncome1_2M(t *testing.T) {
	// Bracket tax:
	// 1% on 10756 = 107.56
	// 2% on 14743 = 294.86
	// 4% on 14746 = 589.84
	// 6% on 15621 = 937.26
	// 8% on 14746 = 1179.68
	// 9.3% on (360659-70612)=290047 = 26974.371
	// 10.3% on (432791-360659)=72132 = 7429.596
	// 11.3% on (721319-432791)=288528 = 32603.664
	// 12.3% on (1200000-721319)=478681 = 58877.763
	// Bracket total = 128994.594
	// Mental health: (1200000-1000000)*0.01 = 2000
	// Total = 130994.594
	tax := ComputeTax(1200000, Single, 2025, StateCA)
	expected := 130994.594
	if !approxEqual(tax, expected) {
		t.Errorf("CA single $1.2M: got %.4f, want %.4f", tax, expected)
	}
}

func TestCAMentalHealthTax(t *testing.T) {
	tests := []struct {
		income   float64
		expected float64
	}{
		{500000, 0},
		{1000000, 0},
		{1000001, 0.01},
		{1200000, 2000},
		{2000000, 10000},
	}
	for _, tc := range tests {
		got := GetCAMentalHealthTax(tc.income)
		if !approxEqual(got, tc.expected) {
			t.Errorf("MentalHealthTax(%.0f): got %.2f, want %.2f", tc.income, got, tc.expected)
		}
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestZeroIncome(t *testing.T) {
	if tax := ComputeTax(0, Single, 2025, Federal); tax != 0 {
		t.Errorf("Federal $0: got %.2f, want 0", tax)
	}
	if tax := ComputeTax(0, Single, 2025, StateCA); tax != 0 {
		t.Errorf("CA $0: got %.2f, want 0", tax)
	}
}

func TestNegativeIncome(t *testing.T) {
	if tax := ComputeTax(-5000, Single, 2025, Federal); tax != 0 {
		t.Errorf("Federal negative: got %.2f, want 0", tax)
	}
	if tax := ComputeTax(-5000, Single, 2025, StateCA); tax != 0 {
		t.Errorf("CA negative: got %.2f, want 0", tax)
	}
}

func TestExactBracketBoundaryFederal(t *testing.T) {
	// Exactly at the top of the first bracket: $11,925
	// 10% on 11925 = 1192.50
	tax := ComputeTax(11925, Single, 2025, Federal)
	expected := 1192.50
	if !approxEqual(tax, expected) {
		t.Errorf("Federal single $11925 (boundary): got %.2f, want %.2f", tax, expected)
	}
}

func TestExactBracketBoundaryCA(t *testing.T) {
	// Exactly at first CA bracket boundary: $10,756
	// 1% on 10756 = 107.56
	tax := ComputeTax(10756, Single, 2025, StateCA)
	expected := 107.56
	if !approxEqual(tax, expected) {
		t.Errorf("CA single $10756 (boundary): got %.2f, want %.2f", tax, expected)
	}
}

// ---------------------------------------------------------------------------
// Standard deduction tests
// ---------------------------------------------------------------------------

func TestFederalStandardDeductions(t *testing.T) {
	tests := []struct {
		status   FilingStatus
		expected float64
	}{
		{Single, 15000},
		{MarriedFilingJoint, 30000},
		{MarriedFilingSep, 15000},
		{HeadOfHousehold, 22500},
		{QualifyingSurvivor, 30000},
	}
	for _, tc := range tests {
		got := GetStandardDeduction(2025, Federal, tc.status)
		if got != tc.expected {
			t.Errorf("Federal std deduction %s: got %.0f, want %.0f", tc.status, got, tc.expected)
		}
	}
}

func TestCAStandardDeductions(t *testing.T) {
	tests := []struct {
		status   FilingStatus
		expected float64
	}{
		{Single, 5706},
		{MarriedFilingJoint, 11412},
	}
	for _, tc := range tests {
		got := GetStandardDeduction(2025, StateCA, tc.status)
		if got != tc.expected {
			t.Errorf("CA std deduction %s: got %.0f, want %.0f", tc.status, got, tc.expected)
		}
	}
}

// ---------------------------------------------------------------------------
// CA Exemption Credit tests
// ---------------------------------------------------------------------------

func TestCAExemptionCredit(t *testing.T) {
	tests := []struct {
		status     FilingStatus
		dependents int
		expected   float64
	}{
		{Single, 0, 144},
		{MarriedFilingJoint, 0, 288},
		{Single, 2, 144 + 2*433},
		{MarriedFilingJoint, 3, 288 + 3*433},
	}
	for _, tc := range tests {
		got := GetCAExemptionCredit(2025, tc.status, tc.dependents)
		if !approxEqual(got, tc.expected) {
			t.Errorf("CA exemption credit %s deps=%d: got %.2f, want %.2f",
				tc.status, tc.dependents, got, tc.expected)
		}
	}
}

// ---------------------------------------------------------------------------
// Rounding tests
// ---------------------------------------------------------------------------

func TestRoundToNearest(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{100.49, 100},
		{100.50, 101}, // Go's math.Round rounds half away from zero
		{100.51, 101},
		{101.50, 102},
		{99.99, 100},
		{0, 0},
		{-50.6, -51},
	}
	for _, tc := range tests {
		got := RoundToNearest(tc.input)
		if got != tc.expected {
			t.Errorf("RoundToNearest(%.2f): got %.0f, want %.0f", tc.input, got, tc.expected)
		}
	}
}

func TestRoundDown(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{100.99, 100},
		{100.01, 100},
		{100.00, 100},
		{0, 0},
		{-50.1, -51},
	}
	for _, tc := range tests {
		got := RoundDown(tc.input)
		if got != tc.expected {
			t.Errorf("RoundDown(%.2f): got %.0f, want %.0f", tc.input, got, tc.expected)
		}
	}
}

// ---------------------------------------------------------------------------
// Unsupported year
// ---------------------------------------------------------------------------

func TestUnsupportedYear(t *testing.T) {
	brackets := GetBrackets(2023, Federal, Single)
	if brackets != nil {
		t.Error("Expected nil brackets for unsupported year 2023")
	}
	tax := ComputeTax(75000, Single, 2023, Federal)
	if tax != 0 {
		t.Errorf("Expected 0 tax for unsupported year, got %.2f", tax)
	}
	deduction := GetStandardDeduction(2023, Federal, Single)
	if deduction != 0 {
		t.Errorf("Expected 0 deduction for unsupported year, got %.0f", deduction)
	}
	credit := GetCAExemptionCredit(2023, Single, 0)
	if credit != 0 {
		t.Errorf("Expected 0 credit for unsupported year, got %.2f", credit)
	}
}

func TestYear2024Brackets(t *testing.T) {
	// Federal 2024 single: $14,600 standard deduction (IRS Rev. Proc. 2023-34)
	deduction := GetStandardDeduction(2024, Federal, Single)
	if deduction != 14600 {
		t.Errorf("2024 federal single deduction = %.0f, want 14600", deduction)
	}

	// CA 2024 single: $5,540 standard deduction
	caDeduction := GetStandardDeduction(2024, StateCA, Single)
	if caDeduction != 5540 {
		t.Errorf("2024 CA single deduction = %.0f, want 5540", caDeduction)
	}

	// Federal tax on $75K single should be reasonable (~$11,553)
	tax := ComputeTax(75000, Single, 2024, Federal)
	if tax < 11000 || tax > 12000 {
		t.Errorf("2024 federal tax on $75K single = %.2f, expected ~$11,553", tax)
	}

	// CA exemption credit
	credit := GetCAExemptionCredit(2024, Single, 0)
	if credit != 140 {
		t.Errorf("2024 CA single exemption credit = %.0f, want 140", credit)
	}
}
