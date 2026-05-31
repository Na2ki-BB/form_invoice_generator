package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"form-invoice-generator/backend/internal/database"
	"form-invoice-generator/backend/internal/invoice"
	"form-invoice-generator/backend/internal/pricing"
	"form-invoice-generator/backend/internal/submission"
)

type submissionRequest struct {
	CustomerName string                  `json:"customerName"`
	CustomerKana string                  `json:"customerKana"`
	PostalCode   string                  `json:"postalCode"`
	Address      string                  `json:"address"`
	Phone        string                  `json:"phone"`
	Email        string                  `json:"email"`
	Note         string                  `json:"note"`
	Items        []submissionRequestItem `json:"items"`
}

type submissionRequestItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

func main() {
	const address = ":8080"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Open(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	submissionRepository := submission.NewRepository(db)

	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/invoice/download", handleInvoiceDownload)
	http.HandleFunc("/submissions", handleSubmission(submissionRepository))

	log.Printf("API server listening on %s", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatal(err)
	}
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func handleInvoiceDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	prayer, err := pricing.Calculate("prayer-a", 1)
	if err != nil {
		http.Error(w, "failed to calculate invoice", http.StatusInternalServerError)
		return
	}
	ofuda, err := pricing.Calculate("ofuda", 2)
	if err != nil {
		http.Error(w, "failed to calculate invoice", http.StatusInternalServerError)
		return
	}

	generated, err := invoice.Generate(invoice.Data{
		InvoiceNumber: "INV-TEST-0001",
		InvoiceDate:   time.Now(),
		CustomerName:  "山田 太郎",
		PostalCode:    "100-0001",
		Address:       "東京都千代田区千代田1-1",
		Note:          "固定データで生成した請求書です。",
		Items:         []pricing.Item{prayer, ofuda},
	})
	if err != nil {
		log.Printf("generate invoice: %v", err)
		http.Error(w, "failed to generate invoice", http.StatusInternalServerError)
		return
	}

	writeInvoice(w, generated, "invoice_test.xlsx")
}

func handleSubmission(repository *submission.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:5173")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request submissionRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		items := make([]pricing.Item, 0, len(request.Items))
		totalAmount := 0
		for _, requestedItem := range request.Items {
			item, err := pricing.Calculate(requestedItem.ProductID, requestedItem.Quantity)
			if err != nil {
				http.Error(w, "invalid item", http.StatusBadRequest)
				return
			}
			if item.Quantity == 0 {
				continue
			}
			items = append(items, item)
			totalAmount += item.Amount
		}
		if len(items) == 0 {
			http.Error(w, "at least one item is required", http.StatusBadRequest)
			return
		}

		submissionID, err := repository.Create(r.Context(), submission.Submission{
			CustomerName:  request.CustomerName,
			CustomerKana:  request.CustomerKana,
			CustomerEmail: request.Email,
			CustomerPhone: request.Phone,
			PostalCode:    request.PostalCode,
			Address:       request.Address,
			Note:          request.Note,
			TotalAmount:   totalAmount,
			Items:         items,
		})
		if err != nil {
			log.Printf("save submission: %v", err)
			http.Error(w, "failed to save submission", http.StatusInternalServerError)
			return
		}
		log.Printf("saved submission: id=%d customer=%q items=%d total=%d", submissionID, request.CustomerName, len(items), totalAmount)

		generated, err := invoice.Generate(invoice.Data{
			InvoiceNumber: "INV-TEMP-0001",
			InvoiceDate:   time.Now(),
			CustomerName:  request.CustomerName,
			PostalCode:    request.PostalCode,
			Address:       request.Address,
			Note:          request.Note,
			Items:         items,
		})
		if err != nil {
			log.Printf("generate submitted invoice: %v", err)
			http.Error(w, "failed to generate invoice", http.StatusInternalServerError)
			return
		}

		writeInvoice(w, generated, "invoice.xlsx")
	}
}

func writeInvoice(w http.ResponseWriter, generated []byte, filename string) {
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(generated)
}
