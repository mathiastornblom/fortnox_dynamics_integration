// +build test

package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"

    "github.com/joho/godotenv"
    "fortnox_dynamics_integration/pkg/dynamics"
)

func main() {
    // Load environment variables from .env file
    if err := godotenv.Load(); err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    // Set environment variables for testing (if not already set)
    dynamicsBaseURL := os.Getenv("DYNAMICS_API_BASE_URL")
    tenantID := os.Getenv("DYNAMICS_TENANT_ID")
    clientID := os.Getenv("DYNAMICS_CLIENT_ID")
    clientSecret := os.Getenv("DYNAMICS_CLIENT_SECRET")

    if dynamicsBaseURL == "" || tenantID == "" || clientID == "" || clientSecret == "" {
        log.Fatalf("Missing required environment variables")
    }

    // Skapa Dynamics 365 klient
    dynamicsClient := dynamics.NewD365Client()
    if err := dynamicsClient.AuthenticateApi(); err != nil {
        log.Fatalf("Failed to authenticate Dynamics client: %v", err)
    }

    // Mock data for testing
    customerNumber := "A44000" // Ale Folkets Hus

    // Sök efter kund i Dynamics 365
    customersData, err := dynamicsClient.SearchCustomer(customerNumber)
    if err != nil {
        log.Fatalf("Failed to search customer: %v", err)
    }

    var customers struct {
        Value []struct {
            CustomerID string `json:"@odata.id"`
            AccountID  string `json:"accountid"`
        } `json:"value"`
    }
    err = json.Unmarshal(customersData, &customers)
    if err != nil {
        log.Fatalf("Failed to unmarshal customers: %v", err)
    }

    if len(customers.Value) == 0 {
        log.Fatalf("No customer found for customer number: %s", customerNumber)
    }

    customerID := customers.Value[0].AccountID

    mockInvoice := dynamics.DynamicsInvoice{
        InvoiceNumber:  "2024-06-24-12345",
        Balance:        100.0,
        Booked:         true,
        Canceled:       false,
        DocumentNumber: "12345",
        DueDate:        "2024-07-24",
        InvoiceDate:    "2024-06-24",
        Total:          100.0,
        Distributor:    100000000,
    }

    // Spara mock faktura till Dynamics 365
    invoiceID, err := dynamicsClient.CreateInvoice(mockInvoice)
    if err != nil {
        log.Fatalf("Failed to save invoice to Dynamics 365: %v", err)
    }
    fmt.Printf("Mock invoice created with ID: %s\n", invoiceID)

    // Mock PDF data
    mockPDF := []byte("This is a mock PDF content for testing purposes.")

    // Ladda upp mock PDF-filen till Dynamics 365
    err = dynamicsClient.UploadFile(invoiceID, "new_invoicepdf", fmt.Sprintf("%s.pdf", mockInvoice.InvoiceNumber), mockPDF)
    if err != nil {
        log.Fatalf("Failed to upload invoice PDF to Dynamics 365: %v", err)
    }

    // Associera fakturan med kundkontot
    associateBody := map[string]string{
        "@odata.id": fmt.Sprintf("%s/api/data/v9.2/accounts(%s)", dynamicsBaseURL, customerID),
    }
    _, err = dynamicsClient.PostRequest(fmt.Sprintf("new_fakturas(%s)/new_customer_account/$ref", invoiceID), associateBody)
    if err != nil {
        log.Fatalf("Failed to associate invoice with customer: %v", err)
    }

    fmt.Println("Mock PDF uploaded and customer associated successfully.")
}