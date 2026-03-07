package mef

import (
	"strings"
	"testing"
	"time"
)

func TestTestClient_AcceptAll(t *testing.T) {
	client := NewTestClient(true)

	result, err := client.SendSubmission([]byte("<Return/>"), "12345")
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
}

func TestTestClient_Pending(t *testing.T) {
	client := NewTestClient(false)

	result, err := client.SendSubmission([]byte("<Return/>"), "12345")
	if err != nil {
		t.Fatalf("SendSubmission error: %v", err)
	}
	if result.Status != StatusPending {
		t.Errorf("Status: got %v, want Pending", result.Status)
	}

	// GetAcknowledgement should transition pending to accepted
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
	_, err := client.SendSubmission([]byte{}, "12345")
	if err == nil {
		t.Error("expected error for empty XML")
	}
}

func TestTestClient_InvalidPIN(t *testing.T) {
	client := NewTestClient(true)
	_, err := client.SendSubmission([]byte("<Return/>"), "123")
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

func TestProductionClient_NoCert(t *testing.T) {
	_, err := NewProductionClient("12345", nil)
	if err == nil {
		t.Error("expected error when creating production client without certificate")
	}
	if !strings.Contains(err.Error(), "certificate configuration is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestProductionClient_NoEFIN(t *testing.T) {
	// Create a valid cert for this test
	notBefore := time.Now().Add(-24 * time.Hour)
	notAfter := time.Now().Add(365 * 24 * time.Hour)
	p12Data, password := generateTestCert(t, notBefore, notAfter)
	certPath := writeTestCert(t, p12Data)

	cc, err := LoadCertificate(certPath, password)
	if err != nil {
		t.Fatalf("LoadCertificate: %v", err)
	}

	_, err = NewProductionClient("", cc)
	if err == nil {
		t.Error("expected error when creating production client without EFIN")
	}
	if !strings.Contains(err.Error(), "EFIN is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestProductionClient_ATSMode(t *testing.T) {
	notBefore := time.Now().Add(-24 * time.Hour)
	notAfter := time.Now().Add(365 * 24 * time.Hour)
	p12Data, password := generateTestCert(t, notBefore, notAfter)
	certPath := writeTestCert(t, p12Data)

	cc, err := LoadCertificate(certPath, password)
	if err != nil {
		t.Fatalf("LoadCertificate: %v", err)
	}

	client, err := NewProductionClient("12345", cc)
	if err != nil {
		t.Fatalf("NewProductionClient: %v", err)
	}

	// Default should be production endpoint
	if client.Endpoint != MeFProductionEndpoint {
		t.Errorf("default endpoint: got %q, want %q", client.Endpoint, MeFProductionEndpoint)
	}

	// Switch to ATS mode
	client.SetATSMode(true)
	if client.Endpoint != MeFATSEndpoint {
		t.Errorf("ATS endpoint: got %q, want %q", client.Endpoint, MeFATSEndpoint)
	}
	if !client.ATSMode {
		t.Error("ATSMode should be true")
	}

	// Switch back to production
	client.SetATSMode(false)
	if client.Endpoint != MeFProductionEndpoint {
		t.Errorf("production endpoint: got %q, want %q", client.Endpoint, MeFProductionEndpoint)
	}
}

func TestProductionClient_ExpiredCert(t *testing.T) {
	notBefore := time.Now().Add(-365 * 24 * time.Hour)
	notAfter := time.Now().Add(-24 * time.Hour) // expired
	p12Data, password := generateTestCert(t, notBefore, notAfter)
	certPath := writeTestCert(t, p12Data)

	cc, err := LoadCertificate(certPath, password)
	if err != nil {
		t.Fatalf("LoadCertificate: %v", err)
	}

	_, err = NewProductionClient("12345", cc)
	if err == nil {
		t.Error("expected error when creating production client with expired cert")
	}
	if !strings.Contains(err.Error(), "certificate validation failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestBuildSOAPEnvelope(t *testing.T) {
	xmlData := []byte("<Return><Data>test</Data></Return>")
	efin := "12345"

	envelope := buildSOAPEnvelope("SendSubmissions", xmlData, efin)

	envStr := string(envelope)

	// Check XML declaration
	if !strings.Contains(envStr, "<?xml version=") {
		t.Error("envelope should contain XML declaration")
	}

	// Check SOAP namespace
	if !strings.Contains(envStr, "http://schemas.xmlsoap.org/soap/envelope/") {
		t.Error("envelope should contain SOAP namespace")
	}

	// Check MeF namespace
	if !strings.Contains(envStr, "http://www.irs.gov/a2a/mef") {
		t.Error("envelope should contain MeF namespace")
	}

	// Check EFIN is included
	if !strings.Contains(envStr, "<mef:EFIN>12345</mef:EFIN>") {
		t.Error("envelope should contain EFIN")
	}

	// Check the action is included
	if !strings.Contains(envStr, "<mef:SendSubmissions>") {
		t.Error("envelope should contain the action element")
	}

	// Check the data is included
	if !strings.Contains(envStr, "<Return><Data>test</Data></Return>") {
		t.Error("envelope should contain the XML data")
	}

	// Verify it contains the closing envelope tag (well-formed)
	if !strings.Contains(envStr, "</soap:Envelope>") {
		t.Error("envelope should contain closing SOAP envelope tag")
	}
}

func TestParseSOAPResponse_Accepted(t *testing.T) {
	soapResp := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <SubmissionId>SUB-123456</SubmissionId>
    <StatusTxt>Accepted</StatusTxt>
  </soap:Body>
</soap:Envelope>`

	result, err := parseSOAPResponse([]byte(soapResp))
	if err != nil {
		t.Fatalf("parseSOAPResponse error: %v", err)
	}
	if result.SubmissionID != "SUB-123456" {
		t.Errorf("SubmissionID: got %q, want %q", result.SubmissionID, "SUB-123456")
	}
	if result.Status != StatusAccepted {
		t.Errorf("Status: got %v, want Accepted", result.Status)
	}
}

func TestParseSOAPResponse_Rejected(t *testing.T) {
	soapResp := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <SubmissionId>SUB-789</SubmissionId>
    <StatusTxt>Rejected</StatusTxt>
    <Rejection>
      <RuleCode>R0001</RuleCode>
      <RuleDesc>Invalid SSN</RuleDesc>
      <FieldName>PrimarySSN</FieldName>
    </Rejection>
  </soap:Body>
</soap:Envelope>`

	result, err := parseSOAPResponse([]byte(soapResp))
	if err != nil {
		t.Fatalf("parseSOAPResponse error: %v", err)
	}
	if result.Status != StatusRejected {
		t.Errorf("Status: got %v, want Rejected", result.Status)
	}
	if len(result.Rejections) != 1 {
		t.Fatalf("Rejections count: got %d, want 1", len(result.Rejections))
	}
	if result.Rejections[0].Code != "R0001" {
		t.Errorf("Rejection code: got %q, want %q", result.Rejections[0].Code, "R0001")
	}
	if result.Rejections[0].Message != "Invalid SSN" {
		t.Errorf("Rejection message: got %q", result.Rejections[0].Message)
	}
}

func TestParseSOAPResponse_Pending(t *testing.T) {
	soapResp := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <SubmissionId>SUB-PENDING</SubmissionId>
    <StatusTxt>Pending</StatusTxt>
  </soap:Body>
</soap:Envelope>`

	result, err := parseSOAPResponse([]byte(soapResp))
	if err != nil {
		t.Fatalf("parseSOAPResponse error: %v", err)
	}
	if result.Status != StatusPending {
		t.Errorf("Status: got %v, want Pending", result.Status)
	}
}

func TestParseSOAPResponse_InvalidXML(t *testing.T) {
	_, err := parseSOAPResponse([]byte("not xml"))
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}
