package dynamics

import (
	"encoding/json"
	"fmt"
)



// AuthenticateApi performs OAuth authentication to obtain an access token
func (d *D365) AuthenticateApi() error {
	resp, err := d.Resty.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"client_id":     d.ClientID,
			"resource":      d.URL,
			"client_secret": d.ClientSecret,
			"grant_type":    "client_credentials"}).
		Post("https://login.microsoftonline.com/" + d.TenantID + "/oauth2/token")

	if err != nil {
		return fmt.Errorf("error obtaining access token from Dynamics 365: %v", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to authenticate: %v", resp.String())
	}

	token := Token{}
	if err := json.Unmarshal(resp.Body(), &token); err != nil {
		return fmt.Errorf("error parsing access token JSON: %v", err)
	}

	d.AccessToken = token.AccessToken
	return nil
}


