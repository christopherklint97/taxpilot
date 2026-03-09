package forms

import (
	"fmt"
	"testing"
)

// buildBenchRegistry creates a registry with n forms, each having fields that
// form a dependency chain. This simulates a realistic solver workload.
func buildBenchRegistry(numForms, fieldsPerForm int) *Registry {
	reg := NewRegistry()
	for i := 0; i < numForms; i++ {
		formID := FormID(fmt.Sprintf("bench_%d", i))
		fields := make([]FieldDef, 0, fieldsPerForm)

		// First field is always UserInput
		fields = append(fields, FieldDef{
			Line:  "1",
			Type:  UserInput,
			Label: fmt.Sprintf("Input %d", i),
		})

		// Remaining fields are Computed, each depending on the previous
		for j := 2; j <= fieldsPerForm; j++ {
			prevKey := FieldKey(formID, fmt.Sprintf("%d", j-1))
			line := fmt.Sprintf("%d", j)
			fields = append(fields, FieldDef{
				Line:      line,
				Type:      Computed,
				Label:     fmt.Sprintf("Computed %d-%d", i, j),
				DependsOn: []string{prevKey},
				Compute: func(d DepValues) float64 {
					return d.Get(prevKey) + 1
				},
			})
		}

		reg.Register(&FormDef{
			ID:           formID,
			Name:         fmt.Sprintf("Bench Form %d", i),
			Jurisdiction: Federal,
			TaxYears:     []int{2025},
			Fields:       fields,
		})
	}
	return reg
}

func BenchmarkDependencyGraphBuild(b *testing.B) {
	benchmarks := []struct {
		name           string
		numForms       int
		fieldsPerForm  int
	}{
		{"Small_5x5", 5, 5},
		{"Medium_10x10", 10, 10},
		{"Large_20x20", 20, 20},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			reg := buildBenchRegistry(bm.numForms, bm.fieldsPerForm)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				g := NewDependencyGraph(reg)
				if err := g.Build(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkTopologicalSort(b *testing.B) {
	benchmarks := []struct {
		name           string
		numForms       int
		fieldsPerForm  int
	}{
		{"Small_5x5", 5, 5},
		{"Medium_10x10", 10, 10},
		{"Large_20x20", 20, 20},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			reg := buildBenchRegistry(bm.numForms, bm.fieldsPerForm)
			g := NewDependencyGraph(reg)
			if err := g.Build(); err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := g.TopologicalSort()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkSolve(b *testing.B) {
	benchmarks := []struct {
		name           string
		numForms       int
		fieldsPerForm  int
	}{
		{"Small_5x5", 5, 5},
		{"Medium_10x10", 10, 10},
		{"Large_20x20", 20, 20},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			reg := buildBenchRegistry(bm.numForms, bm.fieldsPerForm)
			g := NewDependencyGraph(reg)
			if err := g.Build(); err != nil {
				b.Fatal(err)
			}

			// Build inputs map with all UserInput fields
			inputs := make(map[string]float64)
			for i := 0; i < bm.numForms; i++ {
				key := FieldKey(FormID(fmt.Sprintf("bench_%d", i)), "1")
				inputs[key] = float64(i * 1000)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := g.Solve(inputs, nil, 2025)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkMissingInputs(b *testing.B) {
	reg := buildBenchRegistry(10, 10)
	g := NewDependencyGraph(reg)
	if err := g.Build(); err != nil {
		b.Fatal(err)
	}

	inputs := make(map[string]float64)
	for i := 0; i < 10; i++ {
		key := FieldKey(FormID(fmt.Sprintf("bench_%d", i)), "1")
		inputs[key] = float64(i * 1000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.MissingInputs(inputs)
	}
}
