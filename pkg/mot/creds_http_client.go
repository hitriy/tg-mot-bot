package mot

import (
	"context"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
	"time"
)

func CreateHTTPClient(clientID, clientSecret, tokenURL string) *http.Client {
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
		Scopes:       []string{scopeURL},
	}

	// Create an HTTP client that automatically handles token management
	httpClient := config.Client(context.Background())
	httpClient.Timeout = 30 * time.Second

	return httpClient
}
