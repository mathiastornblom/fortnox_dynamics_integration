package fortnox

type MetaInformation struct {
	TotalResources int `json:"@TotalResources"`
	TotalPages     int `json:"@TotalPages"`
	CurrentPage    int `json:"@CurrentPage"`
}

type InvoicesResponse struct {
	MetaInformation MetaInformation `json:"MetaInformation"`
	Invoices        []Invoice       `json:"Invoices"`
}

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