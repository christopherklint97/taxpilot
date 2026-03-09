package forms

// RefField creates a Computed field that references a single dependency.
func RefField(line, label, dep string) FieldDef {
	return FieldDef{
		Line: line, Type: Computed, Label: label,
		DependsOn: []string{dep},
		Compute:   func(dv DepValues) float64 { return dv.Get(dep) },
	}
}

// SumField creates a Computed field that sums multiple dependencies.
func SumField(line, label string, deps ...string) FieldDef {
	return FieldDef{
		Line: line, Type: Computed, Label: label,
		DependsOn: deps,
		Compute: func(dv DepValues) float64 {
			var s float64
			for _, d := range deps {
				s += dv.Get(d)
			}
			return s
		},
	}
}

// DiffField creates a Computed field that returns a - b.
func DiffField(line, label, a, b string) FieldDef {
	return FieldDef{
		Line: line, Type: Computed, Label: label,
		DependsOn: []string{a, b},
		Compute:   func(dv DepValues) float64 { return dv.Get(a) - dv.Get(b) },
	}
}

// MaxZeroField creates a Computed field that returns max(a - b, 0).
func MaxZeroField(line, label, a, b string) FieldDef {
	return FieldDef{
		Line: line, Type: Computed, Label: label,
		DependsOn: []string{a, b},
		Compute: func(dv DepValues) float64 {
			v := dv.Get(a) - dv.Get(b)
			if v < 0 {
				return 0
			}
			return v
		},
	}
}

// NegField creates a Computed field that negates the value of dep.
func NegField(line, label, dep string) FieldDef {
	return FieldDef{
		Line: line, Type: Computed, Label: label,
		DependsOn: []string{dep},
		Compute:   func(dv DepValues) float64 { return -dv.Get(dep) },
	}
}

// WildcardSumField creates a Computed field that sums all values matching a wildcard pattern.
func WildcardSumField(line, label, pattern string) FieldDef {
	return FieldDef{
		Line: line, Type: Computed, Label: label,
		DependsOn: []string{pattern},
		Compute:   func(dv DepValues) float64 { return dv.SumAll(pattern) },
	}
}

// ZeroField creates a Computed field that always returns 0 (placeholder).
func ZeroField(line, label string) FieldDef {
	return FieldDef{
		Line: line, Type: Computed, Label: label,
		Compute: func(dv DepValues) float64 { return 0 },
	}
}

// StrRefField creates a Computed field that copies a string dependency.
func StrRefField(line, label, dep string) FieldDef {
	return FieldDef{
		Line: line, Type: Computed, Label: label,
		DependsOn:  []string{dep},
		ComputeStr: func(dv DepValues) string { return dv.GetString(dep) },
	}
}

// InputField creates a UserInput numeric field.
func InputField(line, label, prompt string) FieldDef {
	return FieldDef{
		Line:   line,
		Type:   UserInput,
		Label:  label,
		Prompt: prompt,
	}
}

// StringInputField creates a UserInput string field.
func StringInputField(line, label, prompt string) FieldDef {
	return FieldDef{
		Line:      line,
		Type:      UserInput,
		ValueType: StringValue,
		Label:     label,
		Prompt:    prompt,
	}
}

// EnumField creates a UserInput field with options.
func EnumField(line, label, prompt string, options ...string) FieldDef {
	return FieldDef{
		Line:    line,
		Type:    UserInput,
		Label:   label,
		Prompt:  prompt,
		Options: options,
	}
}
