package fortnox

import (
	"encoding/json"
	"fmt"
)

func (c *FortnoxClient) FetchInvoices(filters map[string]string) ([]Invoice, error) {
	var allInvoices []Invoice
	page := 1
	limit := 500

	for {
		query := fmt.Sprintf("limit=%d&page=%d", limit, page)
		for key, value := range filters {
			query += fmt.Sprintf("&%s=%s", key, value)
		}
		endpoint := fmt.Sprintf("/invoices?%s", query)
		respBody, err := c.makeAPIRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}

		var invoicesResponse InvoicesResponse
		if err := json.Unmarshal(respBody, &invoicesResponse); err != nil {
			return nil, err
		}

		allInvoices = append(allInvoices, invoicesResponse.Invoices...)

		if page >= invoicesResponse.MetaInformation.TotalPages {
			break
		}
		page++
	}

	return allInvoices, nil
}

func (c *FortnoxClient) FetchInvoicePDF(invoiceNumber string) ([]byte, error) {
	endpoint := fmt.Sprintf("/invoices/%s/preview", invoiceNumber)
	return c.makeAPIRequest("GET", endpoint, nil)
}
