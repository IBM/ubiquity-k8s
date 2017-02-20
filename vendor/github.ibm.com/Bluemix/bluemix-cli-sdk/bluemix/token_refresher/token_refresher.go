package token_refresher

import (
	"encoding/base64"
	"fmt"

	"github.ibm.com/Bluemix/bluemix-cli-sdk/common/rest"
)

type InvalidTokenError struct {
	Description string
}

func NewInvalidTokenError(description string) *InvalidTokenError {
	return &InvalidTokenError{Description: description}
}

func (e *InvalidTokenError) Error() string {
	return "Invalid auth token: " + e.Description
}

type TokenRefresher interface {
	Refresh(oldToken string) (newToken string, refreshToken string, err error)
}

type uaaTokenRefresher struct {
	uaaEndpoint string
	client      *rest.Client
}

func NewTokenRefresher(uaaEndpoint string) *uaaTokenRefresher {
	return &uaaTokenRefresher{
		uaaEndpoint: uaaEndpoint,
		client:      rest.NewClient(),
	}
}

func (t *uaaTokenRefresher) Refresh(oldToken string) (string, string, error) {
	req := rest.PostRequest(fmt.Sprintf("%s/oauth/token", t.uaaEndpoint)).
		Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("cf:"))).
		Field("refresh_token", oldToken).
		Field("grant_type", "refresh_token").
		Field("scope", "")

	tokens := new(uaaTokenResponse)
	apiErr := new(uaaErrorResponse)
	resp, err := t.client.Do(req, tokens, apiErr)

	if err != nil {
		return "", "", err
	}

	if apiErr.Code != "" {
		if apiErr.Code == "invalid-token" {
			return "", "", NewInvalidTokenError(apiErr.Description)
		} else {
			return "", "", fmt.Errorf("Error response from server. StatusCode: %d; description: %s", resp.StatusCode, apiErr.Description)
		}
	}

	return fmt.Sprintf("%s %s", tokens.TokenType, tokens.AccessToken), tokens.RefreshToken, nil
}

type uaaErrorResponse struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
}

type uaaTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
}
