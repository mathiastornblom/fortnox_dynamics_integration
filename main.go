package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"fortnox_dynamics_integration/pkg/fortnox"
	"fortnox_dynamics_integration/pkg/dynamics"
)

const (
	numWorkers     = 5                     // Antalet goroutines som ska köras parallellt
	rateLimit      = 25                    // Max antal förfrågningar per period
	rateLimitPeriod = 5 * time.Second      // Perioden för rate limiting
)

func main() {
	fortnoxClient, err := fortnox.NewFortnoxClient()
	if err != nil {
		log.Fatalf("Failed to create Fortnox client: %v", err)
	}

	// Om vi inte har en giltig access token, behöver vi starta auktoriseringsflödet
	if fortnoxClient.AccessToken == "" || time.Now().After(fortnoxClient.ExpiresAt) {
		log.Println("Starting authorization flow")
		err := fortnoxClient.StartAuthorizationFlow()
		if err != nil {
			log.Fatalf("Failed to start authorization flow: %v", err)
		}
	}

	// Filtrering och sortering om det behövs
	filters := map[string]string{
		// "lastmodified": "2023-01-01", // exempel på filter
	}

	// Nu kan vi använda klienten för att göra API-anrop
	startTime := time.Now()
	invoices, err := fortnoxClient.FetchInvoices(filters)
	if err != nil {
		log.Fatalf("Failed to fetch invoices: %v", err)
	}
	elapsedTime := time.Since(startTime)
	fmt.Printf("Fetched %d invoices in %s\n", len(invoices), elapsedTime)

	// Skapa Dynamics 365 klient
	dynamicsClient := dynamics.NewD365Client()
	if err := dynamicsClient.AuthenticateApi(); err != nil {
		log.Fatalf("Failed to authenticate Dynamics client: %v", err)
	}

	invoiceChan := make(chan fortnox.Invoice, len(invoices))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(fortnoxClient, dynamicsClient, invoiceChan, &wg)
	}

	for _, invoice := range invoices {
		invoiceChan <- invoice
	}
	close(invoiceChan)

	wg.Wait()
}

func worker(fortnoxClient *fortnox.FortnoxClient, dynamicsClient *dynamics.D365, invoices <-chan fortnox.Invoice, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(rateLimitPeriod / rateLimit)
	defer ticker.Stop()

	for invoice := range invoices {
		<-ticker.C
		startTime := time.Now()
		processInvoice(fortnoxClient, dynamicsClient, invoice)
		elapsedTime := time.Since(startTime)
		fmt.Printf("Processed invoice %s in %s\n", invoice.DocumentNumber, elapsedTime)
	}
}

func processInvoice(fortnoxClient *fortnox.FortnoxClient, dynamicsClient *dynamics.D365, invoice fortnox.Invoice) {
	// Kontrollera om fakturan redan finns i Dynamics 365
	existingInvoiceID, err := dynamicsClient.SearchInvoice(invoice.DocumentNumber)
	if err != nil {
		log.Printf("Failed to search invoice for document number %s: %v", invoice.DocumentNumber, err)
		return
	}

	if existingInvoiceID != "" {
		log.Printf("Invoice %s already exists in Dynamics 365", invoice.DocumentNumber)
		return
	}

	// Sök efter kund i Dynamics 365
	customersData, err := dynamicsClient.SearchCustomer(invoice.CustomerNumber)
	if err != nil {
		log.Printf("Failed to search customer for customer number %s, document number %s: %v", invoice.CustomerNumber, invoice.DocumentNumber, err)
		return
	}

	var customers struct {
		Value []struct {
			CustomerID string `json:"@odata.id"`
			AccountID  string `json:"accountid"`
		} `json:"value"`
	}
	err = json.Unmarshal(customersData, &customers)
	if err != nil {
		log.Printf("Failed to unmarshal customers for customer number %s, document number %s: %v", invoice.CustomerNumber, invoice.DocumentNumber, err)
		return
	}

	if len(customers.Value) == 0 {
		log.Printf("No customer found for customer number %s, document number %s", invoice.CustomerNumber, invoice.DocumentNumber)
		return
	}

	customerID := customers.Value[0].AccountID

	// Hämta PDF för fakturan
	invoicePDF, err := fortnoxClient.FetchInvoicePDF(invoice.DocumentNumber)
	if err != nil {
		log.Printf("Failed to fetch invoice PDF for document number %s: %v", invoice.DocumentNumber, err)
		return
	}

	// Förbered data för Dynamics 365
	invoiceNumber := fmt.Sprintf("%s-%s", invoice.InvoiceDate, invoice.DocumentNumber)
	dynamicsInvoice := dynamics.DynamicsInvoice{
		InvoiceNumber:  invoiceNumber,
		Balance:        invoice.Balance,
		Booked:         invoice.Booked,
		Canceled:       invoice.Cancelled,
		DocumentNumber: invoice.DocumentNumber,
		DueDate:        invoice.DueDate,
		InvoiceDate:    invoice.InvoiceDate,
		Total:          invoice.Total,
		Distributor:    100000001,
	}

	// Spara faktura till Dynamics 365
	invoiceID, err := dynamicsClient.CreateInvoice(dynamicsInvoice)
	if err != nil {
		log.Printf("Failed to save invoice for customer number %s, document number %s to Dynamics 365: %v", invoice.CustomerNumber, invoice.DocumentNumber, err)
		return
	}

	// Ladda upp PDF-filen till Dynamics 365
	err = dynamicsClient.UploadFile(invoiceID, "new_invoicepdf", fmt.Sprintf("%s.pdf", invoiceNumber), invoicePDF)
	if err != nil {
		log.Printf("Failed to upload invoice PDF for invoice ID %s, document number %s to Dynamics 365: %v", invoiceID, invoice.DocumentNumber, err)
		return
	}

	// Associera fakturan med kundkontot
	associateBody := map[string]string{
		"@odata.id": fmt.Sprintf("%s/api/data/v9.2/accounts(%s)", dynamicsClient.URL, customerID),
	}
	_, err = dynamicsClient.PostRequest(fmt.Sprintf("new_fakturas(%s)/new_customer_account/$ref", invoiceID), associateBody)
	if err != nil {
		log.Printf("Failed to associate invoice ID %s with customer ID %s for document number %s: %v", invoiceID, customerID, invoice.DocumentNumber, err)
		return
	}

	fmt.Printf("Processed invoice %s for customer %s\n", invoice.DocumentNumber, invoice.CustomerNumber)
}
