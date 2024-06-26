package dynamics

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// CreateInvoice creates a new invoice in Dynamics 365
func (d *D365) CreateInvoice(invoice DynamicsInvoice) (string, error) {
	response, err := d.PostRequest("new_fakturas", invoice)
	if err != nil {
		return "", fmt.Errorf("failed to create invoice: %v", err)
	}

	var createdInvoice struct {
		ID string `json:"new_fakturaid"`
	}
	if err := json.Unmarshal(response, &createdInvoice); err != nil {
		return "", fmt.Errorf("failed to unmarshal created invoice response: %v", err)
	}

	return createdInvoice.ID, nil
}

// SearchInvoice searches for an invoice in Dynamics 365 based on document number
func (d *D365) SearchInvoice(documentNumber string) (string, error) {
	filter := url.QueryEscape(fmt.Sprintf("new_documentnumber eq '%s'", documentNumber))
	query := fmt.Sprintf("new_fakturas?$filter=%s&$top=1", filter)
	response, err := d.GetRequest(query)
	if err != nil {
		return "", err
	}

	var invoices struct {
		Value []struct {
			InvoiceID string `json:"new_fakturaid"`
		} `json:"value"`
	}
	if err := json.Unmarshal(response, &invoices); err != nil {
		return "", fmt.Errorf("failed to unmarshal search invoice response: %v", err)
	}

	if len(invoices.Value) > 0 {
		return invoices.Value[0].InvoiceID, nil
	}

	return "", nil
}