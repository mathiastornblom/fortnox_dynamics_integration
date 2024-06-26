package dynamics

import (
	"encoding/json"
)

// Token represents the JSON structure of the OAuth token response from Dynamics 365
type Token struct {
	TokenType    string 		 `json:"token_type"`
	ExpiresIn    json.Number     `json:"expires_in"`
	ExtExpiresIn json.Number     `json:"ext_expires_in"`
	AccessToken  string 		 `json:"access_token"`
}

// DynamicsInvoice represents the structure of an invoice to be saved in Dynamics 365
type DynamicsInvoice struct {
	InvoiceNumber  string  `json:"new_fakturanummer"`
	Balance        float64 `json:"new_balance"`
	Booked         bool    `json:"new_booked"`
	Canceled       bool    `json:"new_cancelled"`
	DocumentNumber string  `json:"new_documentnumber"`
	DueDate        string  `json:"new_duedate"`
	InvoiceDate    string  `json:"new_invoicedate"`
	Total          float64 `json:"new_total"`
	Distributor    int     `json:"new_distributor"`
}

