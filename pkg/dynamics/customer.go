package dynamics

import (
	"fmt"
	"net/url"
)

// SearchCustomer searches for a customer in Dynamics 365 based on customer number
func (d *D365) SearchCustomer(customerNumber string) ([]byte, error) {
	filter := url.QueryEscape(fmt.Sprintf("new_kundnummer eq '%s'", customerNumber))
	query := fmt.Sprintf("accounts?$filter=%s&$top=1", filter)
	return d.GetRequest(query)
}
