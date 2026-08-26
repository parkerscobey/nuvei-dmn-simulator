package server

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"

	appconfig "github.com/parkerscobey/nuvei-dmn-simulator/internal/config"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/credentials"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/dmn/payment"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/dmn/payment/apm"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/sender"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/siminput"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/targetsafe"
)

//go:embed templates/*.gohtml
var templateFS embed.FS

type VerifyFunc func(ctx context.Context, profile credentials.Profile) (credentials.Verification, error)
type SendFunc func(ctx context.Context, targetURL, encodedPayload string) (sender.Result, error)

type Handler struct {
	configPath string
	verifyFn   VerifyFunc
	sendFn     SendFunc
	tmpl       *template.Template
}

type PageData struct {
	MerchantProfiles []string
	Targets          []string
	APMs             []string
	DefaultStatus    string
}

type ResultData struct {
	Title             string
	Error             string
	TargetHost        string
	TargetClass       string
	TargetReason      string
	VerificationNote  string
	HTTPStatus        int
	HTTPBody          string
	PayloadRows       []PayloadRow
	RawPayload        string
	ShowTargetSummary bool
	classification    targetsafe.Classification
}

type PayloadRow struct {
	Key   string
	Value string
}

func NewHandler(configPath string, verifyFn VerifyFunc, sendFn SendFunc) (*Handler, error) {
	if strings.TrimSpace(configPath) == "" {
		return nil, fmt.Errorf("config path is required")
	}
	tmpl, err := template.ParseFS(templateFS, "templates/page.gohtml", "templates/result.gohtml")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	return &Handler{configPath: configPath, verifyFn: verifyFn, sendFn: sendFn, tmpl: tmpl}, nil
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.handleHome)
	mux.HandleFunc("POST /htmx/preview", h.handlePreview)
	mux.HandleFunc("POST /htmx/send", h.handleSend)
	mux.HandleFunc("POST /htmx/verify", h.handleVerify)
	return mux
}

func (h *Handler) handleHome(w http.ResponseWriter, r *http.Request) {
	cfg, err := appconfig.Load(h.configPath)
	if err != nil {
		h.renderResult(w, ResultData{Title: "Load config", Error: err.Error()})
		return
	}
	data := PageData{
		MerchantProfiles: sortedMerchantProfiles(cfg),
		Targets:          sortedTargets(cfg),
		APMs:             []string{"pix", "boleto", "card", "local-payments-africa"},
		DefaultStatus:    payment.StatusApproved,
	}
	if err := h.tmpl.ExecuteTemplate(w, "page", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) handlePreview(w http.ResponseWriter, r *http.Request) {
	input, result, err := h.resolveInput(r)
	if err != nil {
		h.renderResult(w, ResultData{Title: "Preview", Error: err.Error()})
		return
	}

	result.Title = "Preview"
	result.PayloadRows = payloadRows(input.payload)
	result.RawPayload = input.payload.Encode()
	result.ShowTargetSummary = true
	h.renderResult(w, result)
}

func (h *Handler) handleVerify(w http.ResponseWriter, r *http.Request) {
	if h.verifyFn == nil {
		h.renderResult(w, ResultData{Title: "Verify", Error: "credential verifier is not configured"})
		return
	}

	rawProfile := strings.TrimSpace(r.FormValue("profile"))
	if rawProfile == "" {
		h.renderResult(w, ResultData{Title: "Verify", Error: "profile is required"})
		return
	}

	cfg, err := appconfig.Load(h.configPath)
	if err != nil {
		h.renderResult(w, ResultData{Title: "Verify", Error: err.Error()})
		return
	}

	profile, ok := cfg.Merchants[rawProfile]
	if !ok {
		h.renderResult(w, ResultData{Title: "Verify", Error: fmt.Sprintf("merchant profile %q not found", rawProfile)})
		return
	}
	if err := appconfig.ValidateMerchantProfile(profile); err != nil {
		h.renderResult(w, ResultData{Title: "Verify", Error: err.Error()})
		return
	}

	verification, err := h.verifyFn(r.Context(), credentials.Profile{
		Environment:       profile.Environment,
		MerchantID:        profile.MerchantID,
		MerchantSiteID:    profile.MerchantSiteID,
		MerchantSecretKey: profile.MerchantSecretKey,
	})
	if err != nil {
		h.renderResult(w, ResultData{Title: "Verify", Error: err.Error()})
		return
	}

	note := fmt.Sprintf("Verified profile %q with Nuvei %s environment via /getSessionToken.", rawProfile, verification.Environment)
	if verification.Cached {
		note += " (cached)"
	}
	h.renderResult(w, ResultData{Title: "Verify", VerificationNote: note})
}

func (h *Handler) handleSend(w http.ResponseWriter, r *http.Request) {
	if h.sendFn == nil {
		h.renderResult(w, ResultData{Title: "Send", Error: "sender is not configured"})
		return
	}
	if h.verifyFn == nil {
		h.renderResult(w, ResultData{Title: "Send", Error: "credential verifier is not configured"})
		return
	}

	input, result, err := h.resolveInput(r)
	if err != nil {
		h.renderResult(w, ResultData{Title: "Send", Error: err.Error()})
		return
	}

	allowUntrusted := r.FormValue("allow_untrusted") == "on"
	confirmer := checkboxConfirmer{confirmed: r.FormValue("confirm_send") == "on"}
	checkResult := resultFromData(result)
	err = targetsafe.RequireAllowed(checkResult, allowUntrusted, confirmer)
	if err != nil {
		h.renderResult(w, ResultData{Title: "Send", Error: safetyError(checkResult, err).Error(), ShowTargetSummary: true, TargetHost: result.TargetHost, TargetClass: result.TargetClass, TargetReason: result.TargetReason})
		return
	}

	verification, err := h.verifyFn(r.Context(), credentials.Profile{
		Environment:       input.merchantProfile.Environment,
		MerchantID:        input.merchantProfile.MerchantID,
		MerchantSiteID:    input.merchantProfile.MerchantSiteID,
		MerchantSecretKey: input.merchantProfile.MerchantSecretKey,
	})
	if err != nil {
		h.renderResult(w, ResultData{Title: "Send", Error: err.Error(), ShowTargetSummary: true, TargetHost: result.TargetHost, TargetClass: result.TargetClass, TargetReason: result.TargetReason})
		return
	}

	sendResult, err := h.sendFn(r.Context(), input.targetURL, input.payload.Encode())
	if err != nil {
		h.renderResult(w, ResultData{Title: "Send", Error: err.Error(), ShowTargetSummary: true, TargetHost: result.TargetHost, TargetClass: result.TargetClass, TargetReason: result.TargetReason})
		return
	}

	note := fmt.Sprintf("Verified with Nuvei %s environment via /getSessionToken.", verification.Environment)
	if verification.Cached {
		note += " (cached)"
	}

	result.Title = "Send"
	result.VerificationNote = note
	result.PayloadRows = payloadRows(input.payload)
	result.RawPayload = input.payload.Encode()
	result.HTTPStatus = sendResult.StatusCode
	result.HTTPBody = sendResult.Body
	result.ShowTargetSummary = true
	h.renderResult(w, result)
}

type resolvedInput struct {
	merchantProfile appconfig.MerchantProfile
	targetURL       string
	payload         payment.Payload
}

func (h *Handler) resolveInput(r *http.Request) (resolvedInput, ResultData, error) {
	if err := r.ParseForm(); err != nil {
		return resolvedInput{}, ResultData{}, fmt.Errorf("parse form: %w", err)
	}

	cfg, err := appconfig.Load(h.configPath)
	if err != nil {
		return resolvedInput{}, ResultData{}, err
	}

	rawProfile := strings.TrimSpace(r.FormValue("profile"))
	if rawProfile == "" {
		return resolvedInput{}, ResultData{}, fmt.Errorf("profile is required")
	}
	profile, ok := cfg.Merchants[rawProfile]
	if !ok {
		return resolvedInput{}, ResultData{}, fmt.Errorf("merchant profile %q not found", rawProfile)
	}
	if err := appconfig.ValidateMerchantProfile(profile); err != nil {
		return resolvedInput{}, ResultData{}, err
	}

	rawTarget := strings.TrimSpace(r.FormValue("target"))
	if rawTarget == "" {
		return resolvedInput{}, ResultData{}, fmt.Errorf("target is required")
	}
	targetURL, err := siminput.ResolveTargetURL(cfg, rawTarget)
	if err != nil {
		return resolvedInput{}, ResultData{}, err
	}

	status := strings.ToUpper(strings.TrimSpace(r.FormValue("status")))
	if status == "" {
		status = payment.StatusApproved
	}

	apmValue := strings.ToLower(strings.TrimSpace(r.FormValue("apm")))
	if apmValue == "" {
		apmValue = "pix"
	}

	buildAPM := apm.Pix
	switch apmValue {
	case "pix":
		buildAPM = apm.Pix
	case "boleto":
		buildAPM = apm.Boleto
	case "card":
		buildAPM = apm.Card
	case "local-payments-africa":
		buildAPM = apm.LocalPaymentsAfrica
	default:
		return resolvedInput{}, ResultData{}, fmt.Errorf("unsupported APM %q", apmValue)
	}

	p, err := buildAPM(payment.Options{
		MerchantID:          profile.MerchantID,
		MerchantSiteID:      profile.MerchantSiteID,
		MerchantSecretKey:   profile.MerchantSecretKey,
		Status:              status,
		TotalAmount:         strings.TrimSpace(r.FormValue("total_amount")),
		Currency:            strings.TrimSpace(r.FormValue("currency")),
		UserPaymentOptionID: strings.TrimSpace(r.FormValue("user_payment_option_id")),
		PPPTransactionID:    strings.TrimSpace(r.FormValue("ppp_transaction_id")),
		TransactionID:       strings.TrimSpace(r.FormValue("transaction_id")),
		ClientUniqueID:      strings.TrimSpace(r.FormValue("client_unique_id")),
		ClientRequestID:     strings.TrimSpace(r.FormValue("client_request_id")),
		Reason:              strings.TrimSpace(r.FormValue("reason")),
		ReasonCode:          strings.TrimSpace(r.FormValue("reason_code")),
	})
	if err != nil {
		return resolvedInput{}, ResultData{}, err
	}

	classification := targetsafe.Classify(targetURL, siminput.TrustedTargetProfiles(cfg), nil)
	result := ResultData{
		TargetHost:     classification.Host,
		TargetClass:    classification.Classification.String(),
		TargetReason:   classification.Reason,
		classification: classification.Classification,
	}

	return resolvedInput{merchantProfile: profile, targetURL: targetURL, payload: p}, result, nil
}

func (h *Handler) renderResult(w http.ResponseWriter, data ResultData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "result", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func sortedMerchantProfiles(cfg appconfig.Config) []string {
	profiles := make([]string, 0, len(cfg.Merchants))
	for name := range cfg.Merchants {
		profiles = append(profiles, name)
	}
	sort.Strings(profiles)
	return profiles
}

func sortedTargets(cfg appconfig.Config) []string {
	targets := make([]string, 0, len(cfg.Targets))
	for name := range cfg.Targets {
		targets = append(targets, name)
	}
	sort.Strings(targets)
	return targets
}

func payloadRows(payload payment.Payload) []PayloadRow {
	keys := make([]string, 0, len(payload.Fields))
	for key := range payload.Fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	rows := make([]PayloadRow, 0, len(keys))
	for _, key := range keys {
		rows = append(rows, PayloadRow{Key: key, Value: payload.Fields[key]})
	}
	return rows
}

func resultFromData(data ResultData) targetsafe.Result {
	return targetsafe.Result{Classification: data.classification, Host: data.TargetHost, Reason: data.TargetReason}
}

type checkboxConfirmer struct {
	confirmed bool
}

func (c checkboxConfirmer) Confirm(string) (bool, error) {
	return c.confirmed, nil
}

func safetyError(result targetsafe.Result, original error) error {
	if result.Host == "" {
		return original
	}
	return fmt.Errorf("target host %q is not allowed: %s", result.Host, result.Reason)
}
