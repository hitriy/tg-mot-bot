package mot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	scopeURL = "https://tapi.dvsa.gov.uk/.default"
)

var (
	srcBaseURL = "https://history.mot.api.gov.uk/v1/trade/vehicles"
	tokenURL   = getTokenURL()
)

// getTokenURL returns the token URL from environment variables or a default value
func getTokenURL() string {
	if url := os.Getenv("MOT_TOKEN_URL"); url != "" {
		return url
	}
	return "https://login.microsoftonline.com/a455b827-244f-4c97-b5b4-ce5d13b4d00c/oauth2/v2.0/token"
}

// ClientInterface defines the interface for the MOT client
type ClientInterface interface {
	GetVehicleByRegistration(ctx context.Context, registration string) (*VehicleResponse, error)
}

// Client implements the MOT API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewClient(httpClient *http.Client, apiKey, baseURL string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

type VehicleResponse struct {
	Registration     string    `json:"registration"`
	Make             string    `json:"make"`
	Model            string    `json:"model"`
	FirstUsedDate    string    `json:"firstUsedDate"`
	FuelType         string    `json:"fuelType"`
	PrimaryColour    string    `json:"primaryColour"`
	RegistrationDate string    `json:"registrationDate"`
	ManufactureDate  string    `json:"manufactureDate"`
	EngineSize       string    `json:"engineSize"`
	MotTests         []MotTest `json:"motTests"`
}

type MotTest struct {
	CompletedDate      string   `json:"completedDate"`
	TestResult         string   `json:"testResult"`
	ExpiryDate         string   `json:"expiryDate"`
	OdometerValue      string   `json:"odometerValue"`
	OdometerUnit       string   `json:"odometerUnit"`
	OdometerResultType string   `json:"odometerResultType"`
	MotTestNumber      string   `json:"motTestNumber"`
	Defects            []Defect `json:"defects"`
}

type Defect struct {
	Text      string `json:"text"`
	Type      string `json:"type"`
	Dangerous bool   `json:"dangerous"`
}

func (c *Client) GetVehicleByRegistration(ctx context.Context, registration string) (*VehicleResponse, error) {
	url := fmt.Sprintf("%s/registration/%s", c.baseURL, registration)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var vehicle VehicleResponse
	if err := json.NewDecoder(resp.Body).Decode(&vehicle); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &vehicle, nil
}

// FormatVehicleInfo returns a formatted string representation of the vehicle information
func (v *VehicleResponse) FormatVehicleInfo() string {
	var result strings.Builder

	// Vehicle Details Section
	result.WriteString("🚗 Vehicle Details\n")
	result.WriteString("=================\n")
	result.WriteString(fmt.Sprintf("📝 Registration: %s\n", v.Registration))
	result.WriteString(fmt.Sprintf("🏭 Make: %s\n", v.Make))
	result.WriteString(fmt.Sprintf("🚘 Model: %s\n", v.Model))
	result.WriteString(fmt.Sprintf("⛽ Fuel Type: %s\n", v.FuelType))
	result.WriteString(fmt.Sprintf("🎨 Color: %s\n", v.PrimaryColour))
	result.WriteString(fmt.Sprintf("🔧 Engine Size: %s\n", v.EngineSize))
	result.WriteString(fmt.Sprintf("📅 First Used: %s\n", v.FirstUsedDate))
	result.WriteString(fmt.Sprintf("🏭 Manufacture Date: %s\n", v.ManufactureDate))
	result.WriteString(fmt.Sprintf("📝 Registration Date: %s\n", v.RegistrationDate))

	// MOT History Section
	if len(v.MotTests) > 0 {
		result.WriteString("\n📋 MOT History\n")
		result.WriteString("=============\n")

		for i, test := range v.MotTests {
			result.WriteString(fmt.Sprintf("\nTest #%d (Completed: %s)\n", i+1, test.CompletedDate))
			result.WriteString(fmt.Sprintf("Result: %s\n", getResultEmoji(test.TestResult)))
			result.WriteString(fmt.Sprintf("Expiry: %s\n", test.ExpiryDate))
			result.WriteString(fmt.Sprintf("Mileage: %s %s\n", test.OdometerValue, test.OdometerUnit))

			if len(test.Defects) > 0 {
				result.WriteString("\n⚠️ Defects:\n")
				for _, defect := range test.Defects {
					severity := "⚠️"
					if defect.Dangerous {
						severity = "🚨"
					}
					result.WriteString(fmt.Sprintf("%s %s (%s)\n", severity, defect.Text, defect.Type))
				}
			}
		}
	}

	return result.String()
}

// getResultEmoji returns an appropriate emoji based on the test result
func getResultEmoji(result string) string {
	switch strings.ToLower(result) {
	case "pass":
		return "✅ PASS"
	case "fail":
		return "❌ FAIL"
	default:
		return result
	}
}
