package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
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
	FormSlug     string                  `json:"formSlug"`
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

const (
	defaultAllowedOrigin = "http://127.0.0.1:5173"
	maxJSONBodyBytes     = 64 * 1024
)

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
	priceRuleRepository := pricing.NewRuleRepository(db)

	handler := newHandler(submissionRepository, formRepository, pricingRepository, priceRuleRepository, productRepository)

	log.Printf("API server listening on %s", address)
	if err := http.ListenAndServe(address, handler); err != nil {
		log.Fatal(err)
	}
}

func newHandler(submissionRepository *submission.Repository, formRepository *formrepository.Repository, pricingRepository *pricing.Repository, priceRuleRepository *pricing.RuleRepository, productRepository *product.Repository) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/submissions", handleSubmission(submissionRepository, formRepository, pricingRepository))
	mux.HandleFunc("/public/forms/", handlePublicForm(formRepository, pricingRepository))
	mux.Handle("/admin/submissions", requireLocalAdmin(handleAdminSubmissions(submissionRepository)))
	mux.Handle("/admin/submissions/", requireLocalAdmin(handleAdminSubmissionDetail(submissionRepository)))
	mux.Handle("/admin/invoices/bulk-download", requireLocalAdmin(handleBulkInvoiceDownload(submissionRepository)))
	mux.Handle("/admin/products", requireLocalAdmin(handleAdminProducts(productRepository)))
	mux.Handle("/admin/price-rules", requireLocalAdmin(handleAdminPriceRules(priceRuleRepository)))
	mux.Handle("/admin/forms", requireLocalAdmin(handleAdminForms(formRepository)))
	return mux
}

func configuredAllowedOrigin() string {
	allowedOrigin := strings.TrimSpace(os.Getenv("APP_CORS_ALLOWED_ORIGIN"))
	if allowedOrigin == "" {
		return defaultAllowedOrigin
	}
	return allowedOrigin
}

func setCORSHeaders(w http.ResponseWriter, methods string, headers string) {
	w.Header().Set("Access-Control-Allow-Origin", configuredAllowedOrigin())
	w.Header().Set("Access-Control-Allow-Methods", methods)
	if headers != "" {
		w.Header().Set("Access-Control-Allow-Headers", headers)
	}
}

func handleCORSPreflight(w http.ResponseWriter, r *http.Request, methods string, headers string) bool {
	setCORSHeaders(w, methods, headers)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func decodeJSONRequest(w http.ResponseWriter, r *http.Request, destination any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	return true
}

func requireLocalAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handleCORSPreflight(w, r, "GET, POST, PUT, OPTIONS", "Content-Type, X-Local-Admin") {
			return
		}

		if !isLoopbackRequest(r) || r.Header.Get("X-Local-Admin") != "true" {
			http.Error(w, "admin authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isLoopbackRequest(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func handleAdminForms(repository *formrepository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handleCORSPreflight(w, r, "GET, POST, PUT, OPTIONS", "Content-Type") {
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
		if !decodeJSONRequest(w, r, &requestedForm) {
			return
		}
		if requestedForm.Title == "" || requestedForm.PublicSlug == "" {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		if err := formrepository.Validate(requestedForm, invoice.MaxItems); err != nil {
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

func handleAdminPriceRules(repository *pricing.RuleRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handleCORSPreflight(w, r, "GET, POST, PUT, OPTIONS", "Content-Type") {
			return
		}

		if r.Method == http.MethodGet {
			productID := r.URL.Query().Get("productId")
			if productID == "" {
				http.Error(w, "productId is required", http.StatusBadRequest)
				return
			}
			rules, err := repository.ListByProduct(r.Context(), productID)
			if err != nil {
				log.Printf("list price rules: %v", err)
				http.Error(w, "failed to list price rules", http.StatusInternalServerError)
				return
			}
			if rules == nil {
				rules = []pricing.AdminRule{}
			}
			writeJSON(w, http.StatusOK, rules)
			return
		}

		var requestedRule pricing.AdminRule
		if !decodeJSONRequest(w, r, &requestedRule) {
			return
		}
		if err := pricing.ValidateAdminRule(requestedRule); err != nil {
			http.Error(w, "invalid price rule", http.StatusBadRequest)
			return
		}

		var err error
		if r.Method == http.MethodPost {
			requestedRule.ID, err = repository.Create(r.Context(), requestedRule)
		} else if r.Method == http.MethodPut {
			err = repository.Update(r.Context(), requestedRule)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err != nil {
			log.Printf("save price rule: %v", err)
			http.Error(w, "failed to save price rule", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, requestedRule)
	}
}

func handleAdminProducts(repository *product.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handleCORSPreflight(w, r, "GET, POST, PUT, OPTIONS", "Content-Type") {
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
		if !decodeJSONRequest(w, r, &requestedProduct) {
			return
		}
		if err := product.Validate(requestedProduct); err != nil {
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
		if handleCORSPreflight(w, r, "POST, OPTIONS", "Content-Type") {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			SubmissionIDs []int64 `json:"submissionIds"`
		}
		if !decodeJSONRequest(w, r, &request) {
			return
		}
		if len(request.SubmissionIDs) == 0 {
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

func handleAdminSubmissionDetail(repository *submission.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w, "GET, OPTIONS", "")
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idText := strings.TrimPrefix(r.URL.Path, "/admin/submissions/")
		id, err := strconv.ParseInt(idText, 10, 64)
		if err != nil || id <= 0 {
			http.Error(w, "invalid submission id", http.StatusBadRequest)
			return
		}
		detail, err := repository.FindDetailByID(r.Context(), id)
		if err != nil {
			http.Error(w, "submission not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, detail)
	}
}

func handleAdminSubmissions(repository *submission.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setCORSHeaders(w, "GET, OPTIONS", "")
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

func handlePublicForm(repository *formrepository.Repository, pricingRepository *pricing.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handleCORSPreflight(w, r, "GET, POST, OPTIONS", "Content-Type") {
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/public/forms/")
		if strings.HasSuffix(path, "/quote") {
			slug := strings.TrimSuffix(path, "/quote")
			if slug == "" || strings.Contains(slug, "/") {
				http.Error(w, "form not found", http.StatusNotFound)
				return
			}
			handlePublicFormQuote(w, r, pricingRepository, slug)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if path == "" || strings.Contains(path, "/") {
			http.Error(w, "form not found", http.StatusNotFound)
			return
		}

		foundForm, err := repository.FindBySlug(r.Context(), path)
		if err == formrepository.ErrNotFound {
			http.Error(w, "form not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("find public form: %v", err)
			http.Error(w, "failed to find form", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, foundForm)
	}
}

func handlePublicFormQuote(w http.ResponseWriter, r *http.Request, repository *pricing.Repository, slug string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Items []submissionRequestItem `json:"items"`
	}
	if !decodeJSONRequest(w, r, &request) {
		return
	}
	items, totalAmount, err := calculateItemsForForm(r.Context(), repository, slug, request.Items)
	if err != nil {
		http.Error(w, "invalid item", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Items       []pricing.Item `json:"items"`
		TotalAmount int            `json:"totalAmount"`
	}{Items: items, TotalAmount: totalAmount})
}

func calculateItemsForForm(ctx context.Context, repository *pricing.Repository, slug string, requestedItems []submissionRequestItem) ([]pricing.Item, int, error) {
	items := make([]pricing.Item, 0, len(requestedItems))
	totalAmount := 0
	seenProductIDs := map[string]struct{}{}
	for _, requestedItem := range requestedItems {
		productID := strings.TrimSpace(requestedItem.ProductID)
		if productID == "" {
			return nil, 0, fmt.Errorf("productId is required")
		}
		if _, exists := seenProductIDs[productID]; exists {
			return nil, 0, fmt.Errorf("duplicate productId: %s", productID)
		}
		seenProductIDs[productID] = struct{}{}
		requestedItem.ProductID = productID
		item, err := repository.CalculateForForm(ctx, slug, requestedItem.ProductID, requestedItem.Quantity)
		if err != nil {
			return nil, 0, err
		}
		if item.Quantity == 0 {
			continue
		}
		items = append(items, item)
		totalAmount += item.Amount
	}
	if len(items) == 0 {
		return nil, 0, fmt.Errorf("at least one item is required")
	}
	return items, totalAmount, nil
}

func handleSubmission(repository *submission.Repository, formRepository *formrepository.Repository, pricingRepository *pricing.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if handleCORSPreflight(w, r, "POST, OPTIONS", "Content-Type") {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request submissionRequest
		if !decodeJSONRequest(w, r, &request) {
			return
		}
		if request.FormSlug == "" {
			http.Error(w, "formSlug is required", http.StatusBadRequest)
			return
		}
		formID, err := formRepository.FindIDBySlug(r.Context(), request.FormSlug)
		if err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		items, totalAmount, err := calculateItemsForForm(r.Context(), pricingRepository, request.FormSlug, request.Items)
		if err != nil {
			http.Error(w, "invalid item", http.StatusBadRequest)
			return
		}

		requestedSubmission := submission.Submission{
			FormID:        formID,
			CustomerName:  request.CustomerName,
			CustomerKana:  request.CustomerKana,
			CustomerEmail: request.Email,
			CustomerPhone: request.Phone,
			PostalCode:    request.PostalCode,
			Address:       request.Address,
			Note:          request.Note,
			TotalAmount:   totalAmount,
			Items:         items,
		}
		if err := submission.Validate(requestedSubmission, invoice.MaxItems); err != nil {
			http.Error(w, "invalid submission", http.StatusBadRequest)
			return
		}

		created, err := repository.Create(r.Context(), requestedSubmission)
		if err != nil {
			log.Printf("save submission: %v", err)
			http.Error(w, "failed to save submission", http.StatusInternalServerError)
			return
		}
		log.Printf("saved submission: id=%d invoice=%q customer=%q items=%d total=%d", created.ID, created.InvoiceNumber, request.CustomerName, len(items), totalAmount)

		generated, err := invoice.Generate(invoice.Data{
			InvoiceNumber: created.InvoiceNumber,
			InvoiceDate:   created.SubmittedAt,
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
