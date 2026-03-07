package ca

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CA FTB e-file endpoint URLs.
const (
	FTBProductionEndpoint = "https://efile.ftb.ca.gov/submit"
	FTBPATSEndpoint       = "https://efile-test.ftb.ca.gov/submit"
)

// SubmissionStatus represents the state of a CA e-file submission.
type SubmissionStatus int

const (
	StatusPending  SubmissionStatus = iota
	StatusAccepted
	StatusRejected
	StatusError
)

func (s SubmissionStatus) String() string {
	switch s {
	case StatusPending:
		return "Pending"
	case StatusAccepted:
		return "Accepted"
	case StatusRejected:
		return "Rejected"
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// Rejection represents a CA FTB rejection.
type Rejection struct {
	Code    string
	Message string
	Field   string
}

// SubmissionResult holds the result of a CA e-file submission.
type SubmissionResult struct {
	SubmissionID string
	Status       SubmissionStatus
	Rejections   []Rejection
	Timestamp    time.Time
	Message      string
}

// Client is the interface for CA FTB e-file communication.
type Client interface {
	SendSubmission(xmlData []byte, pin string, priorYearCAagi float64) (*SubmissionResult, error)
	GetAcknowledgement(submissionID string) (*SubmissionResult, error)
}

// ProductionClient connects to the real CA FTB e-file service.
// CA FTB uses HTTPS REST (not SOAP), simpler than MeF.
type ProductionClient struct {
	ProviderID string       // FTB provider registration ID
	PATSMode   bool         // Provider Acceptance Testing System mode
	Endpoint   string       // FTB endpoint URL (auto-set from PATSMode if empty)
	httpClient *http.Client // HTTP client for FTB communication
}

// NewProductionClient creates a new production CA FTB client.
func NewProductionClient(providerID string) (*ProductionClient, error) {
	if providerID == "" {
		return nil, fmt.Errorf("ca: provider ID is required")
	}

	client := &ProductionClient{
		ProviderID: providerID,
		Endpoint:   FTBProductionEndpoint,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	return client, nil
}

// SetPATSMode enables or disables Provider Acceptance Testing System mode.
// When enabled, submissions go to the PATS endpoint instead of production.
func (c *ProductionClient) SetPATSMode(enabled bool) {
	c.PATSMode = enabled
	if enabled {
		c.Endpoint = FTBPATSEndpoint
	} else {
		c.Endpoint = FTBProductionEndpoint
	}
}

// ftbSubmitRequest is the JSON request body for CA FTB submission.
type ftbSubmitRequest struct {
	XMLData        string  `json:"xml_data"`
	PIN            string  `json:"pin"`
	PriorYearCAagi float64 `json:"prior_year_ca_agi"`
}

// ftbSubmitResponse is the JSON response from CA FTB submission.
type ftbSubmitResponse struct {
	SubmissionID string         `json:"submission_id"`
	Status       string         `json:"status"`
	Message      string         `json:"message"`
	Rejections   []ftbRejection `json:"rejections,omitempty"`
}

// ftbRejection represents a single rejection from CA FTB.
type ftbRejection struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field"`
}

// SendSubmission sends the CA return XML to the FTB e-file service.
// The prior year CA AGI is used as a shared secret for authentication.
func (c *ProductionClient) SendSubmission(xmlData []byte, pin string, priorYearCAagi float64) (*SubmissionResult, error) {
	if len(xmlData) == 0 {
		return nil, fmt.Errorf("ca: empty XML data")
	}
	if len(pin) != 5 {
		return nil, fmt.Errorf("ca: invalid PIN length — must be 5 digits")
	}

	reqBody := ftbSubmitRequest{
		XMLData:        string(xmlData),
		PIN:            pin,
		PriorYearCAagi: priorYearCAagi,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ca: marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.Endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("ca: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-FTB-Provider-ID", c.ProviderID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ca: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ca: read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ca: HTTP %d — %s", resp.StatusCode, string(body))
	}

	var ftbResp ftbSubmitResponse
	if err := json.Unmarshal(body, &ftbResp); err != nil {
		return nil, fmt.Errorf("ca: parse response: %w", err)
	}

	return convertFTBResponse(&ftbResp), nil
}

// GetAcknowledgement checks the status of a previously submitted CA return.
func (c *ProductionClient) GetAcknowledgement(submissionID string) (*SubmissionResult, error) {
	if submissionID == "" {
		return nil, fmt.Errorf("ca: submission ID is required")
	}

	url := fmt.Sprintf("%s/%s", c.Endpoint, submissionID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ca: create request: %w", err)
	}

	req.Header.Set("X-FTB-Provider-ID", c.ProviderID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ca: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ca: read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ca: HTTP %d — %s", resp.StatusCode, string(body))
	}

	var ftbResp ftbSubmitResponse
	if err := json.Unmarshal(body, &ftbResp); err != nil {
		return nil, fmt.Errorf("ca: parse response: %w", err)
	}

	return convertFTBResponse(&ftbResp), nil
}

// convertFTBResponse converts an FTB response to a SubmissionResult.
func convertFTBResponse(resp *ftbSubmitResponse) *SubmissionResult {
	result := &SubmissionResult{
		SubmissionID: resp.SubmissionID,
		Timestamp:    time.Now(),
		Message:      resp.Message,
	}

	switch resp.Status {
	case "Accepted":
		result.Status = StatusAccepted
	case "Rejected":
		result.Status = StatusRejected
		for _, r := range resp.Rejections {
			result.Rejections = append(result.Rejections, Rejection{
				Code:    r.Code,
				Message: r.Message,
				Field:   r.Field,
			})
		}
	case "Pending", "":
		result.Status = StatusPending
	default:
		result.Status = StatusError
	}

	return result
}

// TestClient simulates CA FTB submissions.
type TestClient struct {
	AcceptAll   bool
	Submissions []TestSubmission
}

// TestSubmission records a test submission.
type TestSubmission struct {
	XMLData        []byte
	PIN            string
	PriorYearCAagi float64
	Result         *SubmissionResult
}

// NewTestClient creates a test CA FTB client.
func NewTestClient(acceptAll bool) *TestClient {
	return &TestClient{AcceptAll: acceptAll}
}

// SendSubmission simulates a CA FTB submission.
func (c *TestClient) SendSubmission(xmlData []byte, pin string, priorYearCAagi float64) (*SubmissionResult, error) {
	if len(xmlData) == 0 {
		return nil, fmt.Errorf("empty XML data")
	}
	if len(pin) != 5 {
		return nil, fmt.Errorf("invalid PIN length")
	}

	result := &SubmissionResult{
		SubmissionID: fmt.Sprintf("CA-TEST-%d", time.Now().UnixNano()),
		Timestamp:    time.Now(),
	}

	if c.AcceptAll {
		result.Status = StatusAccepted
		result.Message = "CA return accepted (test mode)"
	} else {
		result.Status = StatusPending
		result.Message = "CA submission received (test mode)"
	}

	c.Submissions = append(c.Submissions, TestSubmission{
		XMLData:        xmlData,
		PIN:            pin,
		PriorYearCAagi: priorYearCAagi,
		Result:         result,
	})

	return result, nil
}

// GetAcknowledgement simulates checking CA FTB submission status.
func (c *TestClient) GetAcknowledgement(submissionID string) (*SubmissionResult, error) {
	for _, sub := range c.Submissions {
		if sub.Result.SubmissionID == submissionID {
			if sub.Result.Status == StatusPending {
				sub.Result.Status = StatusAccepted
				sub.Result.Message = "CA return accepted (test mode)"
			}
			return sub.Result, nil
		}
	}
	return nil, fmt.Errorf("submission %s not found", submissionID)
}
