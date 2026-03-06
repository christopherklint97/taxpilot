package state

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// --- helpers ---

func populatedPriorReturn() *TaxReturn {
	ret := NewTaxReturn(2024, "CA")
	ret.FilingStatus = "single"
	ret.StrInputs["1040:filing_status"] = "single"
	ret.StrInputs["1040:first_name"] = "Jane"
	ret.StrInputs["1040:last_name"] = "Doe"
	ret.StrInputs["1040:ssn"] = "123-45-6789"
	ret.StrInputs["w2:1:employer_name"] = "Acme Corp"
	ret.StrInputs["w2:1:employer_ein"] = "12-3456789"

	ret.Inputs["1040:1a"] = 85000
	ret.Inputs["1040:25d"] = 12000

	ret.Computed["1040:11"] = 82000
	ret.Computed["1040:16"] = 11500
	ret.Computed["1040:34"] = 500
	ret.Computed["ca540:17"] = 80000
	ret.Computed["ca540:40"] = 4200
	ret.Computed["ca540:71"] = 5000
	ret.Computed["ca540:91"] = 800

	return ret
}

// --- TestExtractPriorYearContext ---

func TestExtractPriorYearContext(t *testing.T) {
	ret := populatedPriorReturn()
	ctx := ExtractPriorYearContext(ret)

	if ctx.PriorYear != 2024 {
		t.Errorf("PriorYear = %d, want 2024", ctx.PriorYear)
	}
	if ctx.FederalAGI != 82000 {
		t.Errorf("FederalAGI = %f, want 82000", ctx.FederalAGI)
	}
	if ctx.CAAdjustedAGI != 80000 {
		t.Errorf("CAAdjustedAGI = %f, want 80000", ctx.CAAdjustedAGI)
	}
	if ctx.FilingStatus != "single" {
		t.Errorf("FilingStatus = %q, want %q", ctx.FilingStatus, "single")
	}
	if ctx.FirstName != "Jane" {
		t.Errorf("FirstName = %q, want %q", ctx.FirstName, "Jane")
	}
	if ctx.LastName != "Doe" {
		t.Errorf("LastName = %q, want %q", ctx.LastName, "Doe")
	}
	if ctx.SSN != "123-45-6789" {
		t.Errorf("SSN = %q, want %q", ctx.SSN, "123-45-6789")
	}
	if ctx.TotalWages != 85000 {
		t.Errorf("TotalWages = %f, want 85000", ctx.TotalWages)
	}
	if ctx.FedWithholding != 12000 {
		t.Errorf("FedWithholding = %f, want 12000", ctx.FedWithholding)
	}
	if ctx.CAWithholding != 5000 {
		t.Errorf("CAWithholding = %f, want 5000", ctx.CAWithholding)
	}
	if ctx.FedTax != 11500 {
		t.Errorf("FedTax = %f, want 11500", ctx.FedTax)
	}
	if ctx.CATax != 4200 {
		t.Errorf("CATax = %f, want 4200", ctx.CATax)
	}
	if ctx.FedRefund != 500 {
		t.Errorf("FedRefund = %f, want 500", ctx.FedRefund)
	}
	if ctx.CARefund != 800 {
		t.Errorf("CARefund = %f, want 800", ctx.CARefund)
	}
	if ctx.PriorYearCAAGI != 80000 {
		t.Errorf("PriorYearCAAGI = %f, want 80000", ctx.PriorYearCAAGI)
	}

	// AllValues should include both Inputs and Computed
	if ctx.AllValues["1040:1a"] != 85000 {
		t.Errorf("AllValues[1040:1a] = %f, want 85000", ctx.AllValues["1040:1a"])
	}
	if ctx.AllValues["1040:11"] != 82000 {
		t.Errorf("AllValues[1040:11] = %f, want 82000", ctx.AllValues["1040:11"])
	}

	// AllStrValues
	if ctx.AllStrValues["1040:first_name"] != "Jane" {
		t.Errorf("AllStrValues[1040:first_name] = %q, want %q", ctx.AllStrValues["1040:first_name"], "Jane")
	}
}

func TestExtractPriorYearContext_NilReturn(t *testing.T) {
	ctx := ExtractPriorYearContext(nil)
	if ctx == nil {
		t.Fatal("expected non-nil context for nil return")
	}
	if ctx.PriorYear != 0 {
		t.Errorf("PriorYear = %d, want 0", ctx.PriorYear)
	}
	if ctx.AllValues == nil {
		t.Error("AllValues should be initialized")
	}
	if ctx.AllStrValues == nil {
		t.Error("AllStrValues should be initialized")
	}
}

func TestExtractPriorYearContext_EmptyReturn(t *testing.T) {
	ret := NewTaxReturn(2024, "CA")
	ctx := ExtractPriorYearContext(ret)

	if ctx.PriorYear != 2024 {
		t.Errorf("PriorYear = %d, want 2024", ctx.PriorYear)
	}
	if ctx.FederalAGI != 0 {
		t.Errorf("FederalAGI = %f, want 0", ctx.FederalAGI)
	}
	if ctx.FilingStatus != "" {
		t.Errorf("FilingStatus = %q, want empty", ctx.FilingStatus)
	}
}

// --- TestMigrateToCurrentYear ---

func TestMigrateToCurrentYear(t *testing.T) {
	prior := populatedPriorReturn()
	current := MigrateToCurrentYear(prior, 2025)

	if current.TaxYear != 2025 {
		t.Errorf("TaxYear = %d, want 2025", current.TaxYear)
	}
	if current.State != "CA" {
		t.Errorf("State = %q, want %q", current.State, "CA")
	}
	if current.FilingStatus != "single" {
		t.Errorf("FilingStatus = %q, want %q", current.FilingStatus, "single")
	}

	// Carryover string fields
	for key := range CarryoverFields {
		if priorVal, ok := prior.StrInputs[key]; ok {
			if current.StrInputs[key] != priorVal {
				t.Errorf("StrInputs[%s] = %q, want %q", key, current.StrInputs[key], priorVal)
			}
		}
	}

	// PriorYearCtx should be set
	if current.PriorYearCtx == nil {
		t.Fatal("PriorYearCtx should be set")
	}
	if current.PriorYearCtx.PriorYear != 2024 {
		t.Errorf("PriorYearCtx.PriorYear = %d, want 2024", current.PriorYearCtx.PriorYear)
	}
	if current.PriorYearCtx.FederalAGI != 82000 {
		t.Errorf("PriorYearCtx.FederalAGI = %f, want 82000", current.PriorYearCtx.FederalAGI)
	}

	// PriorYear map should have prior values
	if current.PriorYear["1040:11"] != 82000 {
		t.Errorf("PriorYear[1040:11] = %f, want 82000", current.PriorYear["1040:11"])
	}

	// Should not be marked complete
	if current.Complete {
		t.Error("new return should not be complete")
	}
}

func TestMigrateToCurrentYear_FilingStatusFromTopLevel(t *testing.T) {
	prior := NewTaxReturn(2024, "CA")
	prior.FilingStatus = "married_jointly"
	// No StrInputs for filing_status

	current := MigrateToCurrentYear(prior, 2025)
	if current.FilingStatus != "married_jointly" {
		t.Errorf("FilingStatus = %q, want %q", current.FilingStatus, "married_jointly")
	}
}

// --- TestCompareReturns ---

func TestCompareReturns(t *testing.T) {
	prior := &PriorYearContext{
		PriorYear: 2024,
		AllValues: map[string]float64{
			"1040:11":  82000,
			"1040:1a":  85000,
			"1040:16":  11500,
			"ca540:17": 80000,
		},
		AllStrValues: make(map[string]string),
	}

	current := &PriorYearContext{
		PriorYear: 2025,
		AllValues: map[string]float64{
			"1040:11":  100000, // ~22% increase => flagged
			"1040:1a":  86000,  // ~1.2% increase => not flagged
			"1040:16":  14000,  // ~21.7% increase => flagged
			"ca540:17": 98000,  // 22.5% increase => flagged
		},
		AllStrValues: make(map[string]string),
	}

	flags := CompareReturns(prior, current)

	// 1040:1a should NOT be flagged (< 20%)
	for _, f := range flags {
		if f.FieldKey == "1040:1a" {
			t.Errorf("1040:1a should not be flagged, change is only ~1.2%%")
		}
	}

	// Check that the three significant changes are flagged
	flagged := make(map[string]ChangeFlag)
	for _, f := range flags {
		flagged[f.FieldKey] = f
	}

	if _, ok := flagged["1040:11"]; !ok {
		t.Error("1040:11 (Federal AGI) should be flagged")
	}
	if _, ok := flagged["1040:16"]; !ok {
		t.Error("1040:16 (Federal Tax) should be flagged")
	}
	if _, ok := flagged["ca540:17"]; !ok {
		t.Error("ca540:17 (CA AGI) should be flagged")
	}

	// Check severity classification
	agiFlag := flagged["1040:11"]
	// ~22% change should be "info"
	if agiFlag.Severity != "info" {
		t.Errorf("1040:11 severity = %q, want %q", agiFlag.Severity, "info")
	}
	if agiFlag.Label != "Federal AGI" {
		t.Errorf("1040:11 label = %q, want %q", agiFlag.Label, "Federal AGI")
	}
}

func TestCompareReturns_LargeChange(t *testing.T) {
	prior := &PriorYearContext{
		PriorYear:    2024,
		AllValues:    map[string]float64{"1040:11": 50000},
		AllStrValues: make(map[string]string),
	}
	current := &PriorYearContext{
		PriorYear:    2025,
		AllValues:    map[string]float64{"1040:11": 100000}, // 100% increase
		AllStrValues: make(map[string]string),
	}

	flags := CompareReturns(prior, current)
	if len(flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(flags))
	}
	if flags[0].Severity != "attention" {
		t.Errorf("severity = %q, want %q", flags[0].Severity, "attention")
	}
	if math.Abs(flags[0].PercentChange-1.0) > 0.001 {
		t.Errorf("PercentChange = %f, want 1.0", flags[0].PercentChange)
	}
}

func TestCompareReturns_Decrease(t *testing.T) {
	prior := &PriorYearContext{
		PriorYear:    2024,
		AllValues:    map[string]float64{"1040:11": 100000},
		AllStrValues: make(map[string]string),
	}
	current := &PriorYearContext{
		PriorYear:    2025,
		AllValues:    map[string]float64{"1040:11": 60000}, // 40% decrease
		AllStrValues: make(map[string]string),
	}

	flags := CompareReturns(prior, current)
	if len(flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(flags))
	}
	if flags[0].PercentChange >= 0 {
		t.Errorf("PercentChange = %f, want negative", flags[0].PercentChange)
	}
	if flags[0].Severity != "warning" {
		t.Errorf("severity = %q, want %q", flags[0].Severity, "warning")
	}
}

func TestCompareReturns_NilInputs(t *testing.T) {
	flags := CompareReturns(nil, nil)
	if flags != nil {
		t.Errorf("expected nil flags for nil inputs, got %v", flags)
	}

	ctx := &PriorYearContext{AllValues: map[string]float64{}, AllStrValues: map[string]string{}}
	flags = CompareReturns(nil, ctx)
	if flags != nil {
		t.Errorf("expected nil flags when prior is nil, got %v", flags)
	}
	flags = CompareReturns(ctx, nil)
	if flags != nil {
		t.Errorf("expected nil flags when current is nil, got %v", flags)
	}
}

func TestCompareReturns_ZeroPriorValue(t *testing.T) {
	prior := &PriorYearContext{
		AllValues:    map[string]float64{"1040:11": 0},
		AllStrValues: make(map[string]string),
	}
	current := &PriorYearContext{
		AllValues:    map[string]float64{"1040:11": 50000},
		AllStrValues: make(map[string]string),
	}

	flags := CompareReturns(prior, current)
	if len(flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(flags))
	}
	// 100% change from zero
	if flags[0].PercentChange != 1.0 {
		t.Errorf("PercentChange = %f, want 1.0", flags[0].PercentChange)
	}
}

func TestCompareReturns_BothZero(t *testing.T) {
	prior := &PriorYearContext{
		AllValues:    map[string]float64{"1040:11": 0},
		AllStrValues: make(map[string]string),
	}
	current := &PriorYearContext{
		AllValues:    map[string]float64{"1040:11": 0},
		AllStrValues: make(map[string]string),
	}

	flags := CompareReturns(prior, current)
	if len(flags) != 0 {
		t.Errorf("expected 0 flags for zero-to-zero, got %d", len(flags))
	}
}

// --- TestPriorYearStore ---

func TestPriorYearStore_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewPriorYearStoreWithDir(tmpDir)

	ctx := &PriorYearContext{
		PriorYear:      2024,
		FederalAGI:     82000,
		CAAdjustedAGI:  80000,
		FilingStatus:   "single",
		FirstName:      "Jane",
		LastName:       "Doe",
		SSN:            "123-45-6789",
		TotalWages:     85000,
		FedWithholding: 12000,
		CAWithholding:  5000,
		FedTax:         11500,
		CATax:          4200,
		FedRefund:      500,
		CARefund:       800,
		PriorYearCAAGI: 80000,
		AllValues: map[string]float64{
			"1040:11": 82000,
			"1040:1a": 85000,
		},
		AllStrValues: map[string]string{
			"1040:first_name": "Jane",
		},
	}

	if err := store.SaveContext(ctx); err != nil {
		t.Fatalf("SaveContext: %v", err)
	}

	if !store.HasPriorYear(2024) {
		t.Error("HasPriorYear(2024) = false, want true")
	}
	if store.HasPriorYear(2023) {
		t.Error("HasPriorYear(2023) = true, want false")
	}

	loaded, err := store.LoadContext(2024)
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}

	if loaded.PriorYear != ctx.PriorYear {
		t.Errorf("PriorYear = %d, want %d", loaded.PriorYear, ctx.PriorYear)
	}
	if loaded.FederalAGI != ctx.FederalAGI {
		t.Errorf("FederalAGI = %f, want %f", loaded.FederalAGI, ctx.FederalAGI)
	}
	if loaded.FilingStatus != ctx.FilingStatus {
		t.Errorf("FilingStatus = %q, want %q", loaded.FilingStatus, ctx.FilingStatus)
	}
	if loaded.FirstName != ctx.FirstName {
		t.Errorf("FirstName = %q, want %q", loaded.FirstName, ctx.FirstName)
	}
	if loaded.SSN != ctx.SSN {
		t.Errorf("SSN = %q, want %q", loaded.SSN, ctx.SSN)
	}
	if loaded.AllValues["1040:11"] != 82000 {
		t.Errorf("AllValues[1040:11] = %f, want 82000", loaded.AllValues["1040:11"])
	}
	if loaded.AllStrValues["1040:first_name"] != "Jane" {
		t.Errorf("AllStrValues[1040:first_name] = %q, want %q", loaded.AllStrValues["1040:first_name"], "Jane")
	}
}

func TestPriorYearStore_LoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewPriorYearStoreWithDir(tmpDir)

	_, err := store.LoadContext(1999)
	if err == nil {
		t.Error("expected error loading non-existent year")
	}
}

func TestPriorYearStore_SaveNilContext(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewPriorYearStoreWithDir(tmpDir)

	err := store.SaveContext(nil)
	if err == nil {
		t.Error("expected error saving nil context")
	}
}

func TestPriorYearStore_DefaultPriorYearDir(t *testing.T) {
	store := NewPriorYearStoreWithDir("/tmp/test_taxpilot")
	dir := store.DefaultPriorYearDir(2024)
	expected := filepath.Join("/tmp/test_taxpilot", "2024")
	if dir != expected {
		t.Errorf("DefaultPriorYearDir = %q, want %q", dir, expected)
	}
}

func TestPriorYearStore_OverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewPriorYearStoreWithDir(tmpDir)

	ctx1 := &PriorYearContext{
		PriorYear:    2024,
		FederalAGI:   50000,
		AllValues:    map[string]float64{},
		AllStrValues: map[string]string{},
	}
	if err := store.SaveContext(ctx1); err != nil {
		t.Fatalf("SaveContext first: %v", err)
	}

	ctx2 := &PriorYearContext{
		PriorYear:    2024,
		FederalAGI:   75000,
		AllValues:    map[string]float64{},
		AllStrValues: map[string]string{},
	}
	if err := store.SaveContext(ctx2); err != nil {
		t.Fatalf("SaveContext second: %v", err)
	}

	loaded, err := store.LoadContext(2024)
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if loaded.FederalAGI != 75000 {
		t.Errorf("FederalAGI = %f, want 75000 (overwritten value)", loaded.FederalAGI)
	}
}

func TestNewPriorYearStore(t *testing.T) {
	store := NewPriorYearStore()
	home, _ := os.UserHomeDir()
	expectedBase := filepath.Join(home, ".taxpilot", "prior_years")
	dir := store.DefaultPriorYearDir(2024)
	expected := filepath.Join(expectedBase, "2024")
	if dir != expected {
		t.Errorf("DefaultPriorYearDir = %q, want %q", dir, expected)
	}
}
