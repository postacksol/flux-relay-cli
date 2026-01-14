package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode       string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn      int    `json:"expires_in"`
	Interval       int    `json:"interval"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Developer    struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"developer"`
}

type APIError struct {
	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *APIError) Error() string {
	if e.ErrorDescription != "" {
		return fmt.Sprintf("%s: %s", e.ErrorCode, e.ErrorDescription)
	}
	return e.ErrorCode
}

func (e *APIError) Code() string {
	return e.ErrorCode
}

func (c *Client) InitiateDeviceCode() (*DeviceCodeResponse, error) {
	req, err := http.NewRequest("POST", c.BaseURL+"/api/cli/auth/initiate", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil {
			return nil, &apiErr
		}
		return nil, fmt.Errorf("failed to initiate device code: %s", string(body))
	}

	var deviceCode DeviceCodeResponse
	if err := json.Unmarshal(body, &deviceCode); err != nil {
		return nil, err
	}

	return &deviceCode, nil
}

func (c *Client) GetToken(deviceCode string) (*TokenResponse, error) {
	url := fmt.Sprintf("%s/api/cli/auth/token?device_code=%s", c.BaseURL, deviceCode)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusAccepted {
		// Still pending
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil {
			return nil, &apiErr
		}
		return nil, &APIError{
			ErrorCode:        "authorization_pending",
			ErrorDescription: "The user has not yet completed the authorization flow.",
		}
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil {
			return nil, &apiErr
		}
		// If we can't parse the error, create a generic one
		return nil, &APIError{
			ErrorCode:        "api_error",
			ErrorDescription: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		}
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

type UserInfo struct {
	Developer struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"developer"`
}

func (u *UserInfo) ID() string {
	return u.Developer.ID
}

func (u *UserInfo) Email() string {
	return u.Developer.Email
}

func (u *UserInfo) Username() string {
	return u.Developer.Name
}

func (c *Client) GetCurrentUser(accessToken string) (*UserInfo, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/developer/me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil {
			return nil, &apiErr
		}
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}

	var userInfo UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
