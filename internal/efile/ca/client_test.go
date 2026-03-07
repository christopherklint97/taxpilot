package ca

import (
	"strings"
	"testing"
)

func TestTestClient_AcceptAll(t *testing.T) {
	client := NewTestClient(true)

	result, err := client.SendSubmission([]byte("<CAReturn/>"), "12345", 75000)
	if err != nil {
		t.Fatalf("SendSubmission error: %v", err)
	}
	if result.Status != StatusAccepted {
		t.Errorf("Status: got %v, want Accepted", result.Status)
	}
	if result.SubmissionID == "" {
		t.Error("SubmissionID should not be empty")
	}
	if len(client.Submissions) != 1 {
		t.Errorf("Submissions count: got %d, want 1", len(client.Submissions))
	}
	if client.Submissions[0].PriorYearCAagi != 75000 {
		t.Errorf("PriorYearCAagi: got %.2f, want 75000", client.Submissions[0].PriorYearCAagi)
	}
}

func TestTestClient_Pending(t *testing.T) {
	client := NewTestClient(false)

	result, err := client.SendSubmission([]byte("<CAReturn/>"), "12345", 0)
	if err != nil {
		t.Fatalf("SendSubmission error: %v", err)
	}
	if result.Status != StatusPending {
		t.Errorf("Status: got %v, want Pending", result.Status)
	}

	ack, err := client.GetAcknowledgement(result.SubmissionID)
	if err != nil {
		t.Fatalf("GetAcknowledgement error: %v", err)
	}
	if ack.Status != StatusAccepted {
		t.Errorf("Ack Status: got %v, want Accepted", ack.Status)
	}
}

func TestTestClient_EmptyXML(t *testing.T) {
	client := NewTestClient(true)
	_, err := client.SendSubmission([]byte{}, "12345", 0)
	if err == nil {
		t.Error("expected error for empty XML")
	}
}

func TestTestClient_InvalidPIN(t *testing.T) {
	client := NewTestClient(true)
	_, err := client.SendSubmission([]byte("<CAReturn/>"), "1", 0)
	if err == nil {
		t.Error("expected error for short PIN")
	}
}

func TestTestClient_NotFound(t *testing.T) {
	client := NewTestClient(true)
	_, err := client.GetAcknowledgement("nonexistent")
	if err == nil {
		t.Error("expected error for unknown submission ID")
	}
}

func TestSubmissionStatus_String(t *testing.T) {
	tests := []struct {
		status SubmissionStatus
		want   string
	}{
		{StatusPending, "Pending"},
		{StatusAccepted, "Accepted"},
		{StatusRejected, "Rejected"},
		{StatusError, "Error"},
		{SubmissionStatus(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("%d.String(): got %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestProductionClient_NilProvider(t *testing.T) {
	_, err := NewProductionClient("")
	if err == nil {
		t.Error("expected error when creating production client without provider ID")
	}
	if !strings.Contains(err.Error(), "provider ID is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestProductionClient_Valid(t *testing.T) {
	client, err := NewProductionClient("PROV123")
	if err != nil {
		t.Fatalf("NewProductionClient: %v", err)
	}
	if client.ProviderID != "PROV123" {
		t.Errorf("ProviderID: got %q, want %q", client.ProviderID, "PROV123")
	}
	if client.Endpoint != FTBProductionEndpoint {
		t.Errorf("default endpoint: got %q, want %q", client.Endpoint, FTBProductionEndpoint)
	}
}

func TestProductionClient_PATSMode(t *testing.T) {
	client, err := NewProductionClient("PROV123")
	if err != nil {
		t.Fatalf("NewProductionClient: %v", err)
	}

	// Default should be production endpoint
	if client.Endpoint != FTBProductionEndpoint {
		t.Errorf("default endpoint: got %q, want %q", client.Endpoint, FTBProductionEndpoint)
	}

	// Switch to PATS mode
	client.SetPATSMode(true)
	if client.Endpoint != FTBPATSEndpoint {
		t.Errorf("PATS endpoint: got %q, want %q", client.Endpoint, FTBPATSEndpoint)
	}
	if !client.PATSMode {
		t.Error("PATSMode should be true")
	}

	// Switch back to production
	client.SetPATSMode(false)
	if client.Endpoint != FTBProductionEndpoint {
		t.Errorf("production endpoint: got %q, want %q", client.Endpoint, FTBProductionEndpoint)
	}
	if client.PATSMode {
		t.Error("PATSMode should be false")
	}
}

func TestConvertFTBResponse_Accepted(t *testing.T) {
	resp := &ftbSubmitResponse{
		SubmissionID: "CA-123",
		Status:       "Accepted",
		Message:      "Return accepted",
	}

	result := convertFTBResponse(resp)
	if result.Status != StatusAccepted {
		t.Errorf("Status: got %v, want Accepted", result.Status)
	}
	if result.SubmissionID != "CA-123" {
		t.Errorf("SubmissionID: got %q, want %q", result.SubmissionID, "CA-123")
	}
}

func TestConvertFTBResponse_Rejected(t *testing.T) {
	resp := &ftbSubmitResponse{
		SubmissionID: "CA-456",
		Status:       "Rejected",
		Message:      "Return rejected",
		Rejections: []ftbRejection{
			{Code: "CA001", Message: "Invalid AGI", Field: "AGI"},
		},
	}

	result := convertFTBResponse(resp)
	if result.Status != StatusRejected {
		t.Errorf("Status: got %v, want Rejected", result.Status)
	}
	if len(result.Rejections) != 1 {
		t.Fatalf("Rejections count: got %d, want 1", len(result.Rejections))
	}
	if result.Rejections[0].Code != "CA001" {
		t.Errorf("Rejection code: got %q, want %q", result.Rejections[0].Code, "CA001")
	}
}

func TestConvertFTBResponse_Pending(t *testing.T) {
	resp := &ftbSubmitResponse{
		SubmissionID: "CA-789",
		Status:       "Pending",
		Message:      "Processing",
	}

	result := convertFTBResponse(resp)
	if result.Status != StatusPending {
		t.Errorf("Status: got %v, want Pending", result.Status)
	}
}
