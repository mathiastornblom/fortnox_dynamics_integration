package dynamics

import (
    "fmt"
    "github.com/go-resty/resty/v2"
    "time"
    "os"
)

// D365 represents the Dynamics 365 client
type D365 struct {
    Resty        *resty.Client
    URL          string
    TenantID     string
    ClientID     string
    ClientSecret string
    AccessToken  string
    ExpiresAt    time.Time
}

// NewD365Client initializes a new Dynamics 365 client
func NewD365Client() *D365 {
    client := resty.New()
    return &D365{
        Resty:        client,
        URL:          os.Getenv("DYNAMICS_API_BASE_URL"),
        TenantID:     os.Getenv("DYNAMICS_TENANT_ID"),
        ClientID:     os.Getenv("DYNAMICS_CLIENT_ID"),
        ClientSecret: os.Getenv("DYNAMICS_CLIENT_SECRET"),
    }
}

// CheckAndRefreshToken checks if the access token is expired and refreshes it if necessary
func (d *D365) CheckAndRefreshToken() error {
    if time.Now().After(d.ExpiresAt) {
        return d.AuthenticateApi()
    }
    return nil
}

// GetRequest makes an authenticated HTTP GET request to the specified endpoint
func (d *D365) GetRequest(endpoint string) ([]byte, error) {
    if err := d.CheckAndRefreshToken(); err != nil {
        return nil, err
    }

    resp, err := d.Resty.R().
        SetHeader("Authorization", fmt.Sprintf("Bearer %v", d.AccessToken)).
        Get(d.URL + "/api/data/v9.2/" + endpoint)

    if err != nil {
        return nil, err
    }

    if resp.StatusCode() != 200 {
        return nil, fmt.Errorf("error making GET request: %v", resp.String())
    }

    return resp.Body(), nil
}

// PostRequest makes an authenticated HTTP POST request to the specified endpoint with the given request body
func (d *D365) PostRequest(endpoint string, values interface{}) ([]byte, error) {
    if err := d.CheckAndRefreshToken(); err != nil {
        return nil, err
    }

    resp, err := d.Resty.R().
        SetHeader("Content-Type", "application/json; charset=utf-8").
        SetHeader("Authorization", fmt.Sprintf("Bearer %v", d.AccessToken)).
        SetHeader("Prefer", "return=representation").
        SetBody(values).
        Post(d.URL + "/api/data/v9.2/" + endpoint)

    if err != nil {
        return nil, err
    }

    if resp.StatusCode() != 200 && resp.StatusCode() != 201 {
        return nil, fmt.Errorf("error making POST request: %v", resp.String())
    }

    return resp.Body(), nil
}

// PatchRequest makes an authenticated HTTP PATCH request to the specified endpoint with the given request body
func (d *D365) PatchRequest(endpoint string, values interface{}) ([]byte, error) {
    if err := d.CheckAndRefreshToken(); err != nil {
        return nil, err
    }

    resp, err := d.Resty.R().
        SetHeader("Content-Type", "application/json; charset=utf-8").
        SetHeader("Authorization", fmt.Sprintf("Bearer %v", d.AccessToken)).
        SetHeader("Prefer", "return=representation").
        SetBody(values).
        Patch(d.URL + "/api/data/v9.2/" + endpoint)

    if err != nil {
        return nil, err
    }

    if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
        return nil, fmt.Errorf("error making PATCH request: %v", resp.String())
    }

    return resp.Body(), nil
}
