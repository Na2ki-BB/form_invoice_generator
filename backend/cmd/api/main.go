package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"form-invoice-generator/backend/internal/database"
	formrepository "form-invoice-generator/backend/internal/form"
	"form-invoice-generator/backend/internal/invoice"
	"form-invoice-generator/backend/internal/pricing"
	"form-invoice-generator/backend/internal/product"
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
	formRepository := formrepository.NewRepository(db)
	pricingRepository := pricing.NewRepository(db)
	productRepository := product.NewRepository(db)

	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/invoice/download", handleInvoiceDownload)
	http.HandleFunc("/submissions", handleSubmission(submissionRepository, pricingRepository))
	http.HandleFunc("/public/forms/", handlePublicForm(formRepository))
	http.HandleFunc("/admin/submissions", handleAdminSubmissions(submissionRepository))
	http.HandleFunc("/admin/invoices/bulk-download", handleBulkInvoiceDownload(submissionRepository))
	http.HandleFunc("/admin/products", handleAdminProducts(productRepository))
	http.HandleFunc("/admin/forms", handleAdminForms(formRepository))

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

func handleAdminForms(repository *formrepository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:5173")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodGet {
			forms, err := repository.List(r.Context())
			if err != nil {
				log.Printf("list forms: %v", err)
				http.Error(w, "failed to list forms", http.StatusInternalServerError)
				return
			}
			if forms == nil {
				forms = []formrepository.AdminForm{}
			}
			writeJSON(w, http.StatusOK, forms)
			return
		}

		var requestedForm formrepository.AdminForm
		if err := json.NewDecoder(r.Body).Decode(&requestedForm); err != nil || requestedForm.Title == "" || requestedForm.PublicSlug == "" {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		var err error
		if r.Method == http.MethodPost {
			requestedForm.ID, err = repository.Create(r.Context(), requestedForm)
		} else if r.Method == http.MethodPut {
			err = repository.Update(r.Context(), requestedForm)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err != nil {
			log.Printf("save form: %v", err)
			http.Error(w, "failed to save form", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, requestedForm)
	}
}

func handleAdminProducts(repository *product.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:5173")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method == http.MethodGet {
			products, err := repository.List(r.Context())
			if err != nil {
				log.Printf("list products: %v", err)
				http.Error(w, "failed to list products", http.StatusInternalServerError)
				return
			}
			if products == nil {
				products = []product.Product{}
			}
			writeJSON(w, http.StatusOK, products)
			return
		}

		var requestedProduct product.Product
		if err := json.NewDecoder(r.Body).Decode(&requestedProduct); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if requestedProduct.ID == "" || requestedProduct.Name == "" || requestedProduct.Category == "" || requestedProduct.BaseUnitPrice < 0 {
			http.Error(w, "invalid product", http.StatusBadRequest)
			return
		}

		var err error
		switch r.Method {
		case http.MethodPost:
			err = repository.Create(r.Context(), requestedProduct)
		case http.MethodPut:
			err = repository.Update(r.Context(), requestedProduct)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err != nil {
			log.Printf("save product: %v", err)
			http.Error(w, "failed to save product", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, requestedProduct)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func handleBulkInvoiceDownload(repository *submission.Repository) http.HandlerFunc {
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

		var request struct {
			SubmissionIDs []int64 `json:"submissionIds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || len(request.SubmissionIDs) == 0 {
			http.Error(w, "submissionIds are required", http.StatusBadRequest)
			return
		}

		details, err := repository.FindDetailsByIDs(r.Context(), request.SubmissionIDs)
		if err != nil {
			log.Printf("find invoice submissions: %v", err)
			http.Error(w, "failed to find submissions", http.StatusInternalServerError)
			return
		}
		if len(details) != len(request.SubmissionIDs) {
			http.Error(w, "submission not found", http.StatusNotFound)
			return
		}

		generated, err := invoice.GenerateArchive(details)
		if err != nil {
			log.Printf("generate invoice archive: %v", err)
			http.Error(w, "failed to generate invoice archive", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", `attachment; filename="invoices.zip"`)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(generated)
	}
}

func handleAdminSubmissions(repository *submission.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:5173")
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		month, err := time.Parse("2006-01", r.URL.Query().Get("month"))
		if err != nil {
			http.Error(w, "month must use YYYY-MM format", http.StatusBadRequest)
			return
		}

		summaries, err := repository.ListByMonth(r.Context(), month)
		if err != nil {
			log.Printf("list submissions: %v", err)
			http.Error(w, "failed to list submissions", http.StatusInternalServerError)
			return
		}
		if summaries == nil {
			summaries = []submission.Summary{}
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(summaries)
	}
}

func handlePublicForm(repository *formrepository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:5173")
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		slug := strings.TrimPrefix(r.URL.Path, "/public/forms/")
		if slug == "" || strings.Contains(slug, "/") {
			http.Error(w, "form not found", http.StatusNotFound)
			return
		}

		foundForm, err := repository.FindBySlug(r.Context(), slug)
		if err == formrepository.ErrNotFound {
			http.Error(w, "form not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("find public form: %v", err)
			http.Error(w, "failed to find form", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(foundForm)
	}
}

func handleSubmission(repository *submission.Repository, pricingRepository *pricing.Repository) http.HandlerFunc {
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
			item, err := pricingRepository.Calculate(r.Context(), requestedItem.ProductID, requestedItem.Quantity)
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
