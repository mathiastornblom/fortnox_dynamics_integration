package fortnox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Invoice struct {
	Balance        float64 `json:"Balance"`
	Booked         bool    `json:"Booked"`
	Cancelled      bool    `json:"Cancelled"`
	CustomerName   string  `json:"CustomerName"`
	CustomerNumber string  `json:"CustomerNumber"`
	DocumentNumber string  `json:"DocumentNumber"`
	DueDate        string  `json:"DueDate"`
	InvoiceDate    string  `json:"InvoiceDate"`
	Total          float64 `json:"Total"`
}

type MetaInformation struct {
	TotalResources int `json:"@TotalResources"`
	TotalPages     int `json:"@TotalPages"`
	CurrentPage    int `json:"@CurrentPage"`
}

type InvoicesResponse struct {
	MetaInformation MetaInformation `json:"MetaInformation"`
	Invoices        []Invoice       `json:"Invoices"`
}

var rateLimitMutex sync.Mutex
var lastRequestTime time.Time

func (c *FortnoxClient) makeAPIRequest(method, endpoint string, body []byte) ([]byte, error) {
	rateLimitMutex.Lock()
	defer rateLimitMutex.Unlock()

	// Fortnox rate limit handling: 25 requests per 5 seconds
	if time.Since(lastRequestTime) < 200*time.Millisecond {
		time.Sleep(200*time.Millisecond - time.Since(lastRequestTime))
	}
	lastRequestTime = time.Now()

	if time.Now().After(c.ExpiresAt) {
		if err := c.RefreshAccessToken(); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, c.APIBaseURL+endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.AccessToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	var resp *http.Response
	var respErr error
	backoff := time.Millisecond * 100
	maxRetries := 5

	for retries := 0; retries < maxRetries; retries++ {
		resp, respErr = client.Do(req)
		if respErr != nil {
			return nil, respErr
		}
		if resp.StatusCode != http.StatusTooManyRequests {
			break
		}
		time.Sleep(backoff)
		backoff *= 2
	}
	if resp == nil {
		return nil, fmt.Errorf("failed to get a response after %d retries", maxRetries)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

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
