// +build test

package main

import (
    "fmt"
    "log"
    "time"

    "fortnox_dynamics_integration/pkg/fortnox"
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
    invoices, err := fortnoxClient.FetchInvoices(filters)
    if err != nil {
        log.Fatalf("Failed to fetch invoices: %v", err)
    }

    fmt.Printf("Fetched %d invoices\n", len(invoices))
    for _, invoice := range invoices {
        fmt.Printf("Invoice: %+v\n", invoice)
    }
}
