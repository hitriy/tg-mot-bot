package telegram

import (
	"context"
	"errors"
	"testing"

	"mot-bot/pkg/mot"
)

// MockMotClient is a mock implementation of the MOT client
type MockMotClient struct {
	GetVehicleByRegistrationFunc func(ctx context.Context, registration string) (*mot.VehicleResponse, error)
}

func (m *MockMotClient) GetVehicleByRegistration(ctx context.Context, registration string) (*mot.VehicleResponse, error) {
	return m.GetVehicleByRegistrationFunc(ctx, registration)
}

func TestHandleRegistration(t *testing.T) {
	// Create a mock MOT client
	mockClient := &MockMotClient{
		GetVehicleByRegistrationFunc: func(ctx context.Context, registration string) (*mot.VehicleResponse, error) {
			return &mot.VehicleResponse{
				Registration:     "AB12CDE",
				Make:             "FORD",
				Model:            "FOCUS",
				FirstUsedDate:    "2010-01-01",
				FuelType:         "PETROL",
				PrimaryColour:    "BLUE",
				RegistrationDate: "2010-01-01",
				ManufactureDate:  "2009-12-01",
				EngineSize:       "1596",
				MotTests: []mot.MotTest{
					{
						CompletedDate:      "2023-01-01T00:00:00Z",
						TestResult:         "PASSED",
						ExpiryDate:         "2024-01-01",
						OdometerValue:      "100000",
						OdometerUnit:       "MI",
						OdometerResultType: "READ",
						MotTestNumber:      "123456789012",
					},
				},
			}, nil
		},
	}

	// Create a test bot
	bot := &Bot{
		motClient: mockClient,
	}

	// Test the function
	err := bot.handleRegistration(context.Background(), 123456789, "AB12CDE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleRegistration_Error(t *testing.T) {
	// Create a mock MOT client that returns an error
	mockClient := &MockMotClient{
		GetVehicleByRegistrationFunc: func(ctx context.Context, registration string) (*mot.VehicleResponse, error) {
			return nil, errors.New("test error")
		},
	}

	// Create a test bot
	bot := &Bot{
		motClient: mockClient,
	}

	// Test the function
	err := bot.handleRegistration(context.Background(), 123456789, "AB12CDE")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestHandleRegistration_InvalidRegistration(t *testing.T) {
	// Create a mock MOT client
	mockClient := &MockMotClient{
		GetVehicleByRegistrationFunc: func(ctx context.Context, registration string) (*mot.VehicleResponse, error) {
			return nil, errors.New("invalid registration")
		},
	}

	// Create a test bot
	bot := &Bot{
		motClient: mockClient,
	}

	// Test the function
	err := bot.handleRegistration(context.Background(), 123456789, "INVALID")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
