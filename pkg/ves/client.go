package ves

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ClientInterface interface {
	GetVehicleByRegistration(ctx context.Context, registration string) (*Vehicle, error)
}

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Custom time type to handle the API's date format
type CustomTime struct {
	time.Time
}

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	// Remove quotes from the JSON string
	s := string(b)
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}

	// Parse the date in YYYY-MM-DD format
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	ct.Time = t
	return nil
}

type Vehicle struct {
	RegistrationNumber  string     `json:"registrationNumber"`
	TaxStatus           string     `json:"taxStatus"`
	TaxDueDate          CustomTime `json:"taxDueDate"`
	Wheelplan           string     `json:"wheelplan"`
	DateOfLastV5CIssued CustomTime `json:"dateOfLastV5CIssued"`
	EuroStatus          string     `json:"euroStatus"`
}

type requestBody struct {
	RegistrationNumber string `json:"registrationNumber"`
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetVehicleByRegistration(ctx context.Context, registration string) (*Vehicle, error) {
	// Create request body
	reqBody := requestBody{
		RegistrationNumber: registration,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var vehicle Vehicle
	if err := json.NewDecoder(resp.Body).Decode(&vehicle); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &vehicle, nil
}
