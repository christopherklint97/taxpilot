package forms

import "strings"

// GetStr retrieves a string input, returning empty string if not found.
func GetStr(strInputs map[string]string, key string) string {
	if strInputs == nil {
		return ""
	}
	return strInputs[key]
}

// GetNum retrieves a numeric result, returning 0 if not found.
func GetNum(results map[string]float64, key string) float64 {
	if results == nil {
		return 0
	}
	return results[key]
}

// NumExists checks whether a key exists in the results map.
func NumExists(results map[string]float64, key string) bool {
	if results == nil {
		return false
	}
	_, ok := results[key]
	return ok
}

// HasKeyPrefix checks whether any key in the map starts with the given prefix.
func HasKeyPrefix(m map[string]float64, prefix string) bool {
	for k := range m {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}
