package mef

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MeF SOAP endpoint URLs.
const (
	MeFProductionEndpoint = "https://la.www4.irs.gov/a2a/mef"
	MeFATSEndpoint        = "https://la.www4.irs.gov/a2a/mef/ats"
)

// SOAP action URIs for MeF operations.
const (
	soapActionSend = "SendSubmissions"
	soapActionAck  = "GetAcknowledgement"
)

// SubmissionStatus represents the state of an e-file submission.
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

// Rejection represents a MeF rejection with its business rule code.
type Rejection struct {
	Code    string // MeF business rule code (e.g., "R0001")
	Message string // human-readable message
	Field   string // field that caused the rejection, if known
}

// SubmissionResult holds the result of an e-file submission.
type SubmissionResult struct {
	SubmissionID string
	Status       SubmissionStatus
	Rejections   []Rejection
	Timestamp    time.Time
	Message      string
}

// Client is the interface for IRS MeF communication.
type Client interface {
	// SendSubmission transmits the return XML to IRS MeF.
	SendSubmission(xmlData []byte, pin string) (*SubmissionResult, error)
	// GetAcknowledgement checks the status of a previously submitted return.
	GetAcknowledgement(submissionID string) (*SubmissionResult, error)
}

// ProductionClient connects to the real IRS MeF A2A SOAP service.
type ProductionClient struct {
	EFIN       string      // Electronic Filing Identification Number
	CertConfig *CertConfig // Strong Authentication certificate configuration
	ATSMode    bool        // Application Testing System mode
	Endpoint   string      // MeF SOAP endpoint URL (auto-set from ATSMode if empty)
	httpClient *http.Client
}

// NewProductionClient creates a new production MeF client configured with
// the given EFIN and Strong Authentication certificate. It sets up an HTTP
// client with mutual TLS using the certificate.
func NewProductionClient(efin string, certConfig *CertConfig) (*ProductionClient, error) {
	if certConfig == nil {
		return nil, fmt.Errorf("mef: certificate configuration is required")
	}
	if efin == "" {
		return nil, fmt.Errorf("mef: EFIN is required")
	}

	if err := certConfig.ValidateCertificate(); err != nil {
		return nil, fmt.Errorf("mef: certificate validation failed: %w", err)
	}

	tlsConfig, err := certConfig.TLSConfig()
	if err != nil {
		return nil, fmt.Errorf("mef: failed to create TLS config: %w", err)
	}

	client := &ProductionClient{
		EFIN:       efin,
		CertConfig: certConfig,
		Endpoint:   MeFProductionEndpoint,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
			Timeout: 60 * time.Second,
		},
	}
	certConfig.EFIN = efin

	return client, nil
}

// SetATSMode enables or disables Application Testing System mode.
// When enabled, submissions go to the ATS endpoint instead of production.
func (c *ProductionClient) SetATSMode(enabled bool) {
	c.ATSMode = enabled
	if enabled {
		c.Endpoint = MeFATSEndpoint
	} else {
		c.Endpoint = MeFProductionEndpoint
	}
}

// SendSubmission sends the return XML to the IRS MeF service via SOAP.
func (c *ProductionClient) SendSubmission(xmlData []byte, pin string) (*SubmissionResult, error) {
	if len(xmlData) == 0 {
		return nil, fmt.Errorf("mef: empty XML data")
	}
	if len(pin) != 5 {
		return nil, fmt.Errorf("mef: invalid PIN length — must be 5 digits")
	}

	envelope := buildSOAPEnvelope(soapActionSend, xmlData, c.EFIN)

	resp, err := c.doSOAPRequest(soapActionSend, envelope)
	if err != nil {
		return nil, fmt.Errorf("mef: SOAP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("mef: read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mef: HTTP %d — %s", resp.StatusCode, string(body))
	}

	result, err := parseSOAPResponse(body)
	if err != nil {
		return nil, fmt.Errorf("mef: parse SOAP response: %w", err)
	}

	return result, nil
}

// GetAcknowledgement checks the status of a previously submitted return.
func (c *ProductionClient) GetAcknowledgement(submissionID string) (*SubmissionResult, error) {
	if submissionID == "" {
		return nil, fmt.Errorf("mef: submission ID is required")
	}

	ackXML := []byte(fmt.Sprintf(`<SubmissionId>%s</SubmissionId>`, submissionID))
	envelope := buildSOAPEnvelope(soapActionAck, ackXML, c.EFIN)

	resp, err := c.doSOAPRequest(soapActionAck, envelope)
	if err != nil {
		return nil, fmt.Errorf("mef: SOAP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("mef: read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mef: HTTP %d — %s", resp.StatusCode, string(body))
	}

	result, err := parseSOAPResponse(body)
	if err != nil {
		return nil, fmt.Errorf("mef: parse SOAP response: %w", err)
	}

	return result, nil
}

// doSOAPRequest sends a SOAP request to the MeF endpoint.
func (c *ProductionClient) doSOAPRequest(action string, envelope []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.Endpoint, bytes.NewReader(envelope))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", fmt.Sprintf(`"%s"`, action))

	return c.httpClient.Do(req)
}

// --- SOAP envelope types ---

// soapEnvelope is the SOAP envelope for MeF requests.
type soapEnvelope struct {
	XMLName xml.Name `xml:"soap:Envelope"`
	SoapNS  string   `xml:"xmlns:soap,attr"`
	MefNS   string   `xml:"xmlns:mef,attr"`
	Body    soapBody `xml:"soap:Body"`
}

// soapBody wraps the SOAP body content.
type soapBody struct {
	Content string `xml:",innerxml"`
}

// soapResponseEnvelope represents a parsed SOAP response.
type soapResponseEnvelope struct {
	XMLName xml.Name         `xml:"Envelope"`
	Body    soapResponseBody `xml:"Body"`
}

// soapResponseBody contains the response body.
type soapResponseBody struct {
	Content string `xml:",innerxml"`
}

// mefSubmissionResponse represents the MeF submission result in the SOAP body.
type mefSubmissionResponse struct {
	SubmissionID string         `xml:"SubmissionId"`
	StatusTxt    string         `xml:"StatusTxt"`
	Rejections   []mefRejection `xml:"Rejection"`
}

// mefRejection represents a single rejection in the MeF response.
type mefRejection struct {
	Code    string `xml:"RuleCode"`
	Message string `xml:"RuleDesc"`
	Field   string `xml:"FieldName"`
}

// buildSOAPEnvelope wraps MeF XML in the required SOAP envelope.
func buildSOAPEnvelope(action string, xmlData []byte, efin string) []byte {
	inner := fmt.Sprintf(
		`<mef:%s><mef:EFIN>%s</mef:EFIN><mef:Data>%s</mef:Data></mef:%s>`,
		action, efin, string(xmlData), action,
	)

	env := soapEnvelope{
		SoapNS: "http://schemas.xmlsoap.org/soap/envelope/",
		MefNS:  "http://www.irs.gov/a2a/mef",
		Body:   soapBody{Content: inner},
	}

	out, _ := xml.MarshalIndent(env, "", "  ")
	header := []byte(xml.Header)
	return append(header, out...)
}

// parseSOAPResponse extracts the submission result from a SOAP response.
func parseSOAPResponse(body []byte) (*SubmissionResult, error) {
	var env soapResponseEnvelope
	if err := xml.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("unmarshal SOAP envelope: %w", err)
	}

	var mefResp mefSubmissionResponse
	if err := xml.Unmarshal([]byte("<Root>"+env.Body.Content+"</Root>"), &mefResp); err != nil {
		// Try parsing the body content directly
		if err2 := xml.Unmarshal([]byte(env.Body.Content), &mefResp); err2 != nil {
			return nil, fmt.Errorf("unmarshal MeF response: %w (also tried: %w)", err, err2)
		}
	}

	result := &SubmissionResult{
		SubmissionID: mefResp.SubmissionID,
		Timestamp:    time.Now(),
	}

	switch mefResp.StatusTxt {
	case "Accepted":
		result.Status = StatusAccepted
		result.Message = "Return accepted by IRS"
	case "Rejected":
		result.Status = StatusRejected
		result.Message = "Return rejected by IRS"
		for _, r := range mefResp.Rejections {
			result.Rejections = append(result.Rejections, Rejection{
				Code:    r.Code,
				Message: r.Message,
				Field:   r.Field,
			})
		}
	case "Pending", "":
		result.Status = StatusPending
		result.Message = "Submission received, awaiting processing"
	default:
		result.Status = StatusError
		result.Message = fmt.Sprintf("Unexpected status: %s", mefResp.StatusTxt)
	}

	return result, nil
}

// TestClient simulates MeF submissions for development and testing.
type TestClient struct {
	// AcceptAll makes all submissions succeed.
	AcceptAll bool
	// Submissions stores submitted data for inspection.
	Submissions []TestSubmission
}

// TestSubmission records a test submission.
type TestSubmission struct {
	XMLData []byte
	PIN     string
	Result  *SubmissionResult
}

// NewTestClient creates a test MeF client.
func NewTestClient(acceptAll bool) *TestClient {
	return &TestClient{AcceptAll: acceptAll}
}

// SendSubmission simulates a submission.
func (c *TestClient) SendSubmission(xmlData []byte, pin string) (*SubmissionResult, error) {
	if len(xmlData) == 0 {
		return nil, fmt.Errorf("empty XML data")
	}
	if len(pin) != 5 {
		return nil, fmt.Errorf("invalid PIN length")
	}

	result := &SubmissionResult{
		SubmissionID: fmt.Sprintf("TEST-%d", time.Now().UnixNano()),
		Timestamp:    time.Now(),
	}

	if c.AcceptAll {
		result.Status = StatusAccepted
		result.Message = "Return accepted (test mode)"
	} else {
		result.Status = StatusPending
		result.Message = "Submission received, awaiting acknowledgement (test mode)"
	}

	c.Submissions = append(c.Submissions, TestSubmission{
		XMLData: xmlData,
		PIN:     pin,
		Result:  result,
	})

	return result, nil
}

// GetAcknowledgement simulates checking submission status.
func (c *TestClient) GetAcknowledgement(submissionID string) (*SubmissionResult, error) {
	for _, sub := range c.Submissions {
		if sub.Result.SubmissionID == submissionID {
			// In test mode, pending becomes accepted
			if sub.Result.Status == StatusPending {
				sub.Result.Status = StatusAccepted
				sub.Result.Message = "Return accepted (test mode)"
			}
			return sub.Result, nil
		}
	}
	return nil, fmt.Errorf("submission %s not found", submissionID)
}
