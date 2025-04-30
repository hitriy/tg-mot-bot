package mot

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetVehicleByRegistration(t *testing.T) {
	// Create a test server for OAuth2 token
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
		require.NoError(t, err)
	}))
	defer tokenServer.Close()

	// Create a test server for the MOT API
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		assert.Equal(t, "/registration/AB12CDE", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"))

		// Return a test response
		response := VehicleResponse{
			Registration:     "AB12CDE",
			Make:             "FORD",
			Model:            "FOCUS",
			FirstUsedDate:    "2010-01-01",
			FuelType:         "PETROL",
			PrimaryColour:    "BLUE",
			RegistrationDate: "2010-01-01",
			ManufactureDate:  "2009-12-01",
			EngineSize:       "1596",
			MotTests: []MotTest{
				{
					CompletedDate:      "2023-01-01T00:00:00Z",
					TestResult:         "PASSED",
					ExpiryDate:         "2024-01-01",
					OdometerValue:      "100000",
					OdometerUnit:       "MI",
					OdometerResultType: "READ",
					MotTestNumber:      "123456789012",
					Defects: []Defect{
						{
							Text:      "Test defect",
							Type:      "ADVISORY",
							Dangerous: false,
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}))
	defer apiServer.Close()

	// Override the URLs for testing
	tokenURL = tokenServer.URL

	httpClient := CreateHTTPClient("test-client-id", "test-client-secret", tokenServer.URL)

	// Create a client
	client := NewClient(httpClient, "test-api-key", apiServer.URL)

	// Test the function
	vehicle, err := client.GetVehicleByRegistration(context.Background(), "AB12CDE")
	require.NoError(t, err)

	// Verify the response
	assert.Equal(t, "AB12CDE", vehicle.Registration)
	assert.Equal(t, "FOCUS", vehicle.Model)
	require.Len(t, vehicle.MotTests, 1)
	assert.Equal(t, "PASSED", vehicle.MotTests[0].TestResult)
}

func TestGetVehicleByRegistration_Error(t *testing.T) {
	// Create a test server for the MOT API that returns an error
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer apiServer.Close()

	// Create a client
	client := NewClient(&http.Client{}, "test-api-key", apiServer.URL)

	// Test the function
	_, err := client.GetVehicleByRegistration(context.Background(), "AB12CDE")
	assert.Error(t, err)
}
