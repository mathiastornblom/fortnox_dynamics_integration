package fortnox

import (
	"encoding/json"
	"fmt"
)

// FetchInvoices fetches invoices from the Fortnox API based on the provided filters.
// It returns a slice of Invoice objects and an error if any.
// The filters parameter is a map of key-value pairs representing the filters to be applied to the API request.
// Each key represents a filter field, and the corresponding value represents the filter value.
// The function retrieves invoices in batches using pagination, with a default limit of 500 invoices per page.
// It continues fetching invoices until all pages have been retrieved or an error occurs.
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

// FetchInvoicePDF fetches the PDF preview of an invoice from the Fortnox API.
// It takes the invoiceNumber as a parameter and returns the PDF data as a byte slice and an error if any.
func (c *FortnoxClient) FetchInvoicePDF(invoiceNumber string) ([]byte, error) {
	endpoint := fmt.Sprintf("/invoices/%s/preview", invoiceNumber)
	return c.makeAPIRequest("GET", endpoint, nil)
}
