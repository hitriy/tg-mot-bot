package mot

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetVehicleByRegistration(t *testing.T) {
	// Create a test server for OAuth2 token
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer tokenServer.Close()

	// Create a test server for the MOT API
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/trade/vehicles/registration/AB12CDE" {
			t.Errorf("expected path /v1/trade/vehicles/registration/AB12CDE, got %s", r.URL.Path)
		}
		if r.Header.Get("X-API-Key") != "test-api-key" {
			t.Errorf("expected X-API-Key header to be test-api-key, got %s", r.Header.Get("X-API-Key"))
		}

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
		json.NewEncoder(w).Encode(response)
	}))
	defer apiServer.Close()

	// Override the URLs for testing
	tokenURL = tokenServer.URL
	baseURL = apiServer.URL

	// Create a client
	client := NewClient("test-client-id", "test-client-secret", "test-api-key")

	// Test the function
	vehicle, err := client.GetVehicleByRegistration(context.Background(), "AB12CDE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the response
	if vehicle.Registration != "AB12CDE" {
		t.Errorf("expected registration AB12CDE, got %s", vehicle.Registration)
	}
	if vehicle.Make != "FORD" {
		t.Errorf("expected make FORD, got %s", vehicle.Make)
	}
	if len(vehicle.MotTests) != 1 {
		t.Errorf("expected 1 MOT test, got %d", len(vehicle.MotTests))
	}
	if vehicle.MotTests[0].TestResult != "PASSED" {
		t.Errorf("expected test result PASSED, got %s", vehicle.MotTests[0].TestResult)
	}
}

func TestGetVehicleByRegistration_Error(t *testing.T) {
	// Create a test server for OAuth2 token
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer tokenServer.Close()

	// Create a test server for the MOT API that returns an error
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer apiServer.Close()

	// Override the URLs for testing
	tokenURL = tokenServer.URL
	baseURL = apiServer.URL

	// Create a client
	client := NewClient("test-client-id", "test-client-secret", "test-api-key")

	// Test the function
	_, err := client.GetVehicleByRegistration(context.Background(), "AB12CDE")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
