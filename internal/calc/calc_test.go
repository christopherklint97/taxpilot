package calc

import (
	"math"
	"testing"
)

// mockRates for testing currency conversion.
var mockRates = map[string]float64{
	"USD": 1.0,
	"EUR": 0.92,
	"GBP": 0.79,
	"SEK": 10.5,
	"JPY": 149.5,
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

func TestEvalArithmetic(t *testing.T) {
	tests := []struct {
		expr string
		want float64
	}{
		{"100", 100},
		{"100 + 200", 300},
		{"1000 - 250", 750},
		{"50 * 3", 150},
		{"100 / 4", 25},
		{"10 + 20 * 3", 70},   // precedence: 20*3=60, 10+60=70
		{"1,000 + 500", 1500}, // comma separator
		{"$1000", 1000},       // dollar sign
		{"-50 + 100", 50},     // negative number
	}

	for _, tt := range tests {
		result, _, err := Eval(tt.expr, nil)
		if err != nil {
			t.Errorf("Eval(%q) error: %v", tt.expr, err)
			continue
		}
		if !almostEqual(result, tt.want) {
			t.Errorf("Eval(%q) = %.2f, want %.2f", tt.expr, result, tt.want)
		}
	}
}

func TestEvalCurrency(t *testing.T) {
	tests := []struct {
		expr string
		want float64
	}{
		{"1000 EUR", 1000 / 0.92},
		{"100 GBP", 100 / 0.79},
		{"EUR 500", 500 / 0.92},
		{"1000 SEK + 500 USD", 1000/10.5 + 500},
	}

	for _, tt := range tests {
		result, _, err := Eval(tt.expr, mockRates)
		if err != nil {
			t.Errorf("Eval(%q) error: %v", tt.expr, err)
			continue
		}
		if !almostEqual(result, tt.want) {
			t.Errorf("Eval(%q) = %.2f, want %.2f", tt.expr, result, tt.want)
		}
	}
}

func TestEvalErrors(t *testing.T) {
	tests := []struct {
		expr string
	}{
		{""},
		{"+ 5"},
		{"100 /"},
		{"100 / 0"},
	}

	for _, tt := range tests {
		_, _, err := Eval(tt.expr, mockRates)
		if err == nil {
			t.Errorf("Eval(%q) expected error, got nil", tt.expr)
		}
	}
}

func TestTokenize(t *testing.T) {
	tokens, err := tokenize("1000 EUR + 500")
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[0].currency != "EUR" {
		t.Errorf("token[0] currency = %q, want EUR", tokens[0].currency)
	}
	if tokens[0].value != 1000 {
		t.Errorf("token[0] value = %.2f, want 1000", tokens[0].value)
	}
}
