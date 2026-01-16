package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	IsActive    bool   `json:"isActive"`
}

type ProjectsResponse struct {
	Projects []Project `json:"projects"`
}

func (c *Client) ListProjects(accessToken string) (*ProjectsResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/developer/projects", nil)
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
		return nil, fmt.Errorf("failed to list projects: %s", string(body))
	}

	var projectsResponse ProjectsResponse
	if err := json.Unmarshal(body, &projectsResponse); err != nil {
		return nil, err
	}

	return &projectsResponse, nil
}

type Server struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	IsActive    bool   `json:"isActive"`
	HasApiKey   bool   `json:"hasApiKey"`
	DatabaseURL string `json:"databaseUrl,omitempty"`
}

type ServersResponse struct {
	Servers []Server `json:"servers"`
}

func (c *Client) ListServers(accessToken string, projectID string) (*ServersResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/developer/projects/"+projectID+"/servers", nil)
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
		return nil, fmt.Errorf("failed to list servers: %s", string(body))
	}

	var serversResponse ServersResponse
	if err := json.Unmarshal(body, &serversResponse); err != nil {
		return nil, err
	}

	return &serversResponse, nil
}

type Database struct {
	ID           string `json:"id"`
	DatabaseName string `json:"databaseName"`
	DatabaseURL  string `json:"databaseUrl,omitempty"`
	HasToken     bool   `json:"hasToken"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
	IsActive     bool   `json:"isActive"`
}

type DatabasesResponse struct {
	Databases []Database `json:"databases"`
}

func (c *Client) ListDatabases(accessToken string, projectID string, serverID string) (*DatabasesResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/developer/projects/"+projectID+"/servers/"+serverID+"/databases", nil)
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
		return nil, fmt.Errorf("failed to list databases: %s", string(body))
	}

	var databasesResponse DatabasesResponse
	if err := json.Unmarshal(body, &databasesResponse); err != nil {
		return nil, err
	}

	return &databasesResponse, nil
}

type QueryRequest struct {
	Query string        `json:"query"`
	Args  []interface{} `json:"args,omitempty"`
}

type QueryResponse struct {
	Columns      []string        `json:"columns"`
	Rows         [][]interface{} `json:"rows"`
	RowsAffected int             `json:"rowsAffected"`
	ExecutionTime int            `json:"executionTime"`
	Success      bool            `json:"success"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
}

func (c *Client) ExecuteQuery(accessToken string, projectID string, serverID string, query string, args []interface{}) (*QueryResponse, error) {
	url := fmt.Sprintf("%s/api/developer/projects/%s/servers/%s/database/query", c.BaseURL, projectID, serverID)
	
	reqBody := QueryRequest{
		Query: query,
		Args:  args,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

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
		return nil, fmt.Errorf("failed to execute query: %s", string(body))
	}

	// API may return response wrapped in "result" object or directly
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, err
	}

	// Check if response is wrapped in "result" object (for system queries)
	var responseData map[string]interface{}
	if result, ok := rawResponse["result"].(map[string]interface{}); ok {
		responseData = result
	} else {
		// Direct response format
		responseData = rawResponse
	}

	// Convert to QueryResponse
	queryResponse := QueryResponse{
		Success:      true, // Default to true if we got here
		RowsAffected: 0,
		ExecutionTime: 0,
	}

	// Extract columns
	if cols, ok := responseData["columns"].([]interface{}); ok {
		queryResponse.Columns = make([]string, len(cols))
		for i, col := range cols {
			if str, ok := col.(string); ok {
				queryResponse.Columns[i] = str
			}
		}
	}

	// Extract rows
	if rows, ok := responseData["rows"].([]interface{}); ok {
		queryResponse.Rows = make([][]interface{}, len(rows))
		for i, row := range rows {
			if rowArray, ok := row.([]interface{}); ok {
				queryResponse.Rows[i] = rowArray
			}
		}
	}

	// Extract execution time
	if et, ok := responseData["executionTime"].(float64); ok {
		queryResponse.ExecutionTime = int(et)
	}

	// Extract rows affected
	if ra, ok := responseData["rowsAffected"].(float64); ok {
		queryResponse.RowsAffected = int(ra)
	}

	// Extract success flag (may not be present, default to true)
	if success, ok := responseData["success"].(bool); ok {
		queryResponse.Success = success
	}

	// Extract error message
	if errMsg, ok := responseData["errorMessage"].(string); ok {
		queryResponse.ErrorMessage = errMsg
		queryResponse.Success = false
	}

	// If no columns but we have rows, it might be a different format
	// Try to handle empty result sets gracefully
	if len(queryResponse.Columns) == 0 && len(queryResponse.Rows) == 0 {
		// This is a valid empty result, not an error
		queryResponse.Success = true
	}

	return &queryResponse, nil
}

type CreateNameserverRequest struct {
	DatabaseName string `json:"databaseName"`
	DatabaseURL  string `json:"databaseUrl,omitempty"`
	DatabaseToken string `json:"databaseToken,omitempty"`
}

type CreateNameserverResponse struct {
	Database struct {
		ID           string `json:"id"`
		DatabaseName string `json:"databaseName"`
		DatabaseURL  string `json:"databaseUrl"`
		HasToken     bool   `json:"hasToken"`
		CreatedAt    string `json:"createdAt"`
		IsActive     bool   `json:"isActive"`
	} `json:"database"`
	Message string `json:"message"`
}

func (c *Client) CreateNameserver(accessToken string, projectID string, serverID string, nameserverName string) (*CreateNameserverResponse, error) {
	url := fmt.Sprintf("%s/api/developer/projects/%s/servers/%s/databases", c.BaseURL, projectID, serverID)
	
	reqBody := CreateNameserverRequest{
		DatabaseName: nameserverName,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil {
			return nil, &apiErr
		}
		return nil, fmt.Errorf("failed to create nameserver: %s", string(body))
	}

	var response CreateNameserverResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

type InitializeNameserverResponse struct {
	Message string   `json:"message"`
	Tables  []string `json:"tables,omitempty"`
}

func (c *Client) InitializeNameserver(accessToken string, projectID string, serverID string, nameserverID string) (*InitializeNameserverResponse, error) {
	url := fmt.Sprintf("%s/api/developer/projects/%s/servers/%s/databases/%s/initialize", c.BaseURL, projectID, serverID, nameserverID)
	
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err == nil {
			return nil, &apiErr
		}
		return nil, fmt.Errorf("failed to initialize nameserver: %s", string(body))
	}

	var response InitializeNameserverResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
