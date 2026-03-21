package calc

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

// exchangeRateAPI is a free API that requires no key.
const exchangeRateAPI = "https://open.er-api.com/v6/latest/USD"

// rateCache caches exchange rates for the session.
var (
	rateCache     map[string]float64
	rateCacheMu   sync.Mutex
	rateCacheTime time.Time
	rateCacheTTL  = 1 * time.Hour
)

// rateResponse is the JSON shape from the exchange rate API.
type rateResponse struct {
	Result string             `json:"result"`
	Rates  map[string]float64 `json:"rates"`
}

// FetchRates fetches USD-based exchange rates, using a 1-hour cache.
func FetchRates() (map[string]float64, error) {
	rateCacheMu.Lock()
	defer rateCacheMu.Unlock()

	if rateCache != nil && time.Since(rateCacheTime) < rateCacheTTL {
		return rateCache, nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(exchangeRateAPI)
	if err != nil {
		return nil, fmt.Errorf("fetch exchange rates: %w", err)
	}
	defer resp.Body.Close()

	var data rateResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode exchange rates: %w", err)
	}
	if data.Result != "success" {
		return nil, fmt.Errorf("exchange rate API returned: %s", data.Result)
	}

	rateCache = data.Rates
	rateCacheTime = time.Now()
	return rateCache, nil
}

// ConvertToUSD converts an amount in the given currency to USD.
// Returns the USD amount and the rate used.
func ConvertToUSD(amount float64, currency string, rates map[string]float64) (float64, float64, error) {
	currency = strings.ToUpper(currency)
	if currency == "USD" || currency == "$" {
		return amount, 1.0, nil
	}
	rate, ok := rates[currency]
	if !ok {
		return 0, 0, fmt.Errorf("unknown currency: %s", currency)
	}
	if rate == 0 {
		return 0, 0, fmt.Errorf("zero rate for %s", currency)
	}
	usd := amount / rate
	return usd, rate, nil
}

// token types for the expression parser.
type tokenKind int

const (
	tokNumber tokenKind = iota
	tokPlus
	tokMinus
	tokMul
	tokDiv
)

type token struct {
	kind     tokenKind
	value    float64  // for tokNumber
	currency string   // optional currency code attached to number
}

// tokenize breaks an expression string into tokens.
// Supports: numbers (with optional decimal), currency codes after numbers,
// and operators +, -, *, /, x.
func tokenize(expr string) ([]token, error) {
	expr = strings.TrimSpace(expr)
	var tokens []token
	i := 0

	for i < len(expr) {
		ch := rune(expr[i])

		// Skip whitespace
		if unicode.IsSpace(ch) {
			i++
			continue
		}

		// Operators
		switch ch {
		case '+':
			tokens = append(tokens, token{kind: tokPlus})
			i++
			continue
		case '-':
			// Could be negative number or subtraction
			if len(tokens) == 0 || tokens[len(tokens)-1].kind != tokNumber {
				// Negative number — fall through to number parsing
			} else {
				tokens = append(tokens, token{kind: tokMinus})
				i++
				continue
			}
		case '*', 'x', 'X':
			tokens = append(tokens, token{kind: tokMul})
			i++
			continue
		case '/':
			tokens = append(tokens, token{kind: tokDiv})
			i++
			continue
		}

		// Number (possibly with leading currency symbol or trailing currency code)
		if ch == '-' || ch == '.' || unicode.IsDigit(ch) || ch == '$' {
			numStart := i
			leadingCurrency := ""

			// Skip leading $ sign
			if ch == '$' {
				leadingCurrency = "USD"
				i++
			}

			// Handle negative sign
			if i < len(expr) && expr[i] == '-' {
				i++
			}

			// Parse digits and decimal
			hasDigit := false
			for i < len(expr) && (unicode.IsDigit(rune(expr[i])) || expr[i] == '.' || expr[i] == ',') {
				if expr[i] == ',' {
					// Skip comma separators (e.g., 1,000.50)
					i++
					continue
				}
				hasDigit = true
				i++
			}

			if !hasDigit {
				return nil, fmt.Errorf("expected number at position %d", numStart)
			}

			// Extract the numeric part (strip commas)
			numStr := strings.ReplaceAll(expr[numStart:i], ",", "")
			if leadingCurrency != "" {
				numStr = strings.TrimPrefix(numStr, "$")
			}
			val, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number %q: %w", numStr, err)
			}

			// Check for trailing currency code (e.g., "1000 EUR" or "1000EUR")
			// Skip whitespace between number and currency
			j := i
			for j < len(expr) && expr[j] == ' ' {
				j++
			}

			currency := leadingCurrency
			if j < len(expr) && unicode.IsLetter(rune(expr[j])) {
				cStart := j
				for j < len(expr) && unicode.IsLetter(rune(expr[j])) {
					j++
				}
				code := strings.ToUpper(expr[cStart:j])
				// Only treat as currency if it's 3 letters (ISO code)
				if len(code) == 3 {
					currency = code
					i = j
				}
			}

			tokens = append(tokens, token{kind: tokNumber, value: val, currency: currency})
			continue
		}

		// Unknown character — might be a currency code at start
		if unicode.IsLetter(ch) {
			// Could be a currency prefix like "EUR 1000"
			cStart := i
			for i < len(expr) && unicode.IsLetter(rune(expr[i])) {
				i++
			}
			code := strings.ToUpper(expr[cStart:i])
			if len(code) != 3 {
				return nil, fmt.Errorf("unexpected text %q at position %d", code, cStart)
			}
			// Expect a number to follow
			for i < len(expr) && expr[i] == ' ' {
				i++
			}
			if i >= len(expr) || (!unicode.IsDigit(rune(expr[i])) && expr[i] != '.' && expr[i] != '-') {
				return nil, fmt.Errorf("expected number after currency code %s", code)
			}
			// Parse the number
			numStart := i
			if expr[i] == '-' {
				i++
			}
			for i < len(expr) && (unicode.IsDigit(rune(expr[i])) || expr[i] == '.' || expr[i] == ',') {
				if expr[i] == ',' {
					i++
					continue
				}
				i++
			}
			numStr := strings.ReplaceAll(expr[numStart:i], ",", "")
			val, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number %q: %w", numStr, err)
			}
			tokens = append(tokens, token{kind: tokNumber, value: val, currency: code})
			continue
		}

		return nil, fmt.Errorf("unexpected character %q at position %d", string(ch), i)
	}

	return tokens, nil
}

// Eval evaluates a simple arithmetic expression with optional currency conversions.
// Returns the result in USD (if any currencies were used) or plain number.
// Also returns a human-readable breakdown string.
func Eval(expr string, rates map[string]float64) (float64, string, error) {
	tokens, err := tokenize(expr)
	if err != nil {
		return 0, "", err
	}
	if len(tokens) == 0 {
		return 0, "", fmt.Errorf("empty expression")
	}

	// Convert all currency amounts to USD
	hasCurrency := false
	var breakdownParts []string
	for i := range tokens {
		if tokens[i].kind == tokNumber && tokens[i].currency != "" {
			hasCurrency = true
			if tokens[i].currency != "USD" {
				if rates == nil {
					return 0, "", fmt.Errorf("exchange rates not available — check your internet connection")
				}
				usd, rate, err := ConvertToUSD(tokens[i].value, tokens[i].currency, rates)
				if err != nil {
					return 0, "", err
				}
				breakdownParts = append(breakdownParts,
					fmt.Sprintf("%.2f %s = $%.2f (1 USD = %.4f %s)",
						tokens[i].value, tokens[i].currency, usd, rate, tokens[i].currency))
				tokens[i].value = usd
			}
		}
	}

	// Simple precedence: evaluate * and / first, then + and -
	// First pass: resolve multiplication and division
	var addTokens []token
	i := 0
	for i < len(tokens) {
		if tokens[i].kind == tokNumber {
			val := tokens[i].value
			// Look ahead for * or /
			for i+2 < len(tokens) && (tokens[i+1].kind == tokMul || tokens[i+1].kind == tokDiv) {
				op := tokens[i+1]
				if tokens[i+2].kind != tokNumber {
					return 0, "", fmt.Errorf("expected number after operator")
				}
				right := tokens[i+2].value
				if op.kind == tokMul {
					val *= right
				} else {
					if right == 0 {
						return 0, "", fmt.Errorf("division by zero")
					}
					val /= right
				}
				i += 2
			}
			addTokens = append(addTokens, token{kind: tokNumber, value: val})
			i++
		} else {
			addTokens = append(addTokens, tokens[i])
			i++
		}
	}

	// Second pass: resolve addition and subtraction
	if len(addTokens) == 0 || addTokens[0].kind != tokNumber {
		return 0, "", fmt.Errorf("expression must start with a number")
	}
	result := addTokens[0].value
	j := 1
	for j < len(addTokens) {
		if j+1 >= len(addTokens) {
			return 0, "", fmt.Errorf("unexpected end of expression")
		}
		op := addTokens[j]
		if addTokens[j+1].kind != tokNumber {
			return 0, "", fmt.Errorf("expected number after operator")
		}
		right := addTokens[j+1].value
		switch op.kind {
		case tokPlus:
			result += right
		case tokMinus:
			result -= right
		default:
			return 0, "", fmt.Errorf("unexpected operator")
		}
		j += 2
	}

	// Round to 2 decimal places
	result = math.Round(result*100) / 100

	breakdown := ""
	if len(breakdownParts) > 0 {
		breakdown = strings.Join(breakdownParts, "\n")
	}
	if hasCurrency {
		if breakdown != "" {
			breakdown += "\n"
		}
		breakdown += fmt.Sprintf("= $%.2f USD", result)
	}

	return result, breakdown, nil
}
