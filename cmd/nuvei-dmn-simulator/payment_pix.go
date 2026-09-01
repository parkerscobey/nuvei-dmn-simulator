package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	appconfig "github.com/parkerscobey/nuvei-dmn-simulator/internal/config"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/credentials"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/dmn/payment"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/dmn/payment/apm"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/sender"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/siminput"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/targetsafe"
	"github.com/spf13/cobra"
)

type paymentPixFlags struct {
	configPath               string
	profile                  string
	rawFile                  string
	keepRawMerchantFields    bool
	status                   string
	target                   string
	totalAmount              string
	currency                 string
	userPaymentOptionID      string
	pppTransactionID         string
	transactionID            string
	clientUniqueID           string
	clientRequestID          string
	reason                   string
	reasonCode               string
	allowUntrustedTarget     bool
	requireCorrelationFields bool
	customField1             string
	customField2             string
	customField3             string
	customField4             string
	customField5             string
	customField6             string
	customField7             string
	customField8             string
	customField9             string
	customField10            string
	customField11            string
	customField12            string
	customField13            string
	customField14            string
	customField15            string
}

var verifyMerchantProfile = func(ctx context.Context, profile credentials.Profile) (credentials.Verification, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return credentials.NewClient(nil).Verify(ctx, profile)
}

var sendDMNPayload = func(ctx context.Context, targetURL, encodedPayload string) (sender.Result, error) {
	return sender.NewClient(nil).Send(ctx, sender.Request{TargetURL: targetURL, EncodedPayload: encodedPayload})
}

func newPreviewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Build and preview a signed DMN payload without sending",
	}

	cmd.AddCommand(newPreviewPaymentCommand())
	return cmd
}

func newPreviewPaymentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payment",
		Short: "Preview payment DMN payloads",
	}
	cmd.AddCommand(newPreviewPaymentPixCommand())
	cmd.AddCommand(newPreviewPaymentBoletoCommand())
	cmd.AddCommand(newPreviewPaymentCardCommand())
	cmd.AddCommand(newPreviewPaymentLocalPaymentsAfricaCommand())
	cmd.AddCommand(newPreviewPaymentFromRawCommand())
	return cmd
}

func newPreviewPaymentPixCommand() *cobra.Command {
	return newPreviewPaymentAPMCommand("pix", "Preview a signed Pix payment DMN payload", resolvePaymentPixInputs)
}

func newPreviewPaymentBoletoCommand() *cobra.Command {
	return newPreviewPaymentAPMCommand("boleto", "Preview a signed Boleto payment DMN payload", resolvePaymentBoletoInputs)
}

func newPreviewPaymentCardCommand() *cobra.Command {
	return newPreviewPaymentAPMCommand("card", "Preview a signed card payment DMN payload", resolvePaymentCardInputs)
}

func newPreviewPaymentLocalPaymentsAfricaCommand() *cobra.Command {
	return newPreviewPaymentAPMCommand("local-payments-africa", "Preview a signed Local Payments Africa payment DMN payload", resolvePaymentLocalPaymentsAfricaInputs)
}

func newPreviewPaymentAPMCommand(use, short string, resolve func(paymentPixFlags) (resolvedPaymentPixInputs, error)) *cobra.Command {
	flags := paymentPixFlags{}

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateStrictMode(cmd, flags); err != nil {
				return err
			}

			resolved, err := resolve(flags)
			if err != nil {
				return err
			}

			classification := targetsafe.Classify(resolved.targetURL, resolved.trustedProfiles, nil)
			printTargetSummary(cmd, classification)
			if classification.Classification == targetsafe.ClassificationUntrusted {
				fmt.Fprintln(cmd.OutOrStdout(), "Note: this target is untrusted and send is blocked by default unless you pass --allow-untrusted-target.")
			}

			printPayloadPreview(cmd, resolved.payload)
			return nil
		},
	}

	bindPaymentPixFlags(cmd, &flags, false)
	return cmd
}

func newSendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a signed DMN payload",
	}

	cmd.AddCommand(newSendPaymentCommand())
	return cmd
}

func newSendPaymentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payment",
		Short: "Send payment DMN payloads",
	}
	cmd.AddCommand(newSendPaymentPixCommand())
	cmd.AddCommand(newSendPaymentBoletoCommand())
	cmd.AddCommand(newSendPaymentCardCommand())
	cmd.AddCommand(newSendPaymentLocalPaymentsAfricaCommand())
	cmd.AddCommand(newSendPaymentFromRawCommand())
	return cmd
}

func newPreviewPaymentFromRawCommand() *cobra.Command {
	flags := paymentPixFlags{}

	cmd := &cobra.Command{
		Use:   "from-raw",
		Short: "Preview a signed payment DMN payload from a raw URL-encoded payload",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := resolvePaymentFromRawInputs(flags)
			if err != nil {
				return err
			}

			classification := targetsafe.Classify(resolved.targetURL, resolved.trustedProfiles, nil)
			printTargetSummary(cmd, classification)
			if classification.Classification == targetsafe.ClassificationUntrusted {
				fmt.Fprintln(cmd.OutOrStdout(), "Note: this target is untrusted and send is blocked by default unless you pass --allow-untrusted-target.")
			}

			printPayloadPreview(cmd, resolved.payload)
			return nil
		},
	}

	bindPaymentFromRawFlags(cmd, &flags, false)
	return cmd
}

func newSendPaymentFromRawCommand() *cobra.Command {
	flags := paymentPixFlags{}

	cmd := &cobra.Command{
		Use:   "from-raw",
		Short: "Send a signed payment DMN payload from a raw URL-encoded payload",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := resolvePaymentFromRawInputs(flags)
			if err != nil {
				return err
			}

			checkOpts := targetsafe.CheckOptions{
				TrustedProfiles: resolved.trustedProfiles,
				AllowUntrusted:  flags.allowUntrustedTarget,
				Confirmer:       targetsafe.NewConsoleConfirmer(),
			}
			classification, err := targetsafe.Check(resolved.targetURL, checkOpts)
			if err != nil {
				return safetyError(classification, err)
			}
			printTargetSummary(cmd, classification)

			_, err = verifyMerchantProfile(cmd.Context(), toCredentialProfile(resolved.merchantProfile))
			if err != nil {
				return err
			}

			result, err := sendDMNPayload(cmd.Context(), resolved.targetURL, resolved.payload.Encode())
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "HTTP status: %d\n", result.StatusCode)
			fmt.Fprintln(cmd.OutOrStdout(), "Response body:")
			fmt.Fprintln(cmd.OutOrStdout(), result.Body)
			return nil
		},
	}

	bindPaymentFromRawFlags(cmd, &flags, true)
	return cmd
}

func newSendPaymentPixCommand() *cobra.Command {
	return newSendPaymentAPMCommand("pix", "Send a signed Pix payment DMN payload", resolvePaymentPixInputs)
}

func newSendPaymentBoletoCommand() *cobra.Command {
	return newSendPaymentAPMCommand("boleto", "Send a signed Boleto payment DMN payload", resolvePaymentBoletoInputs)
}

func newSendPaymentCardCommand() *cobra.Command {
	return newSendPaymentAPMCommand("card", "Send a signed card payment DMN payload", resolvePaymentCardInputs)
}

func newSendPaymentLocalPaymentsAfricaCommand() *cobra.Command {
	return newSendPaymentAPMCommand("local-payments-africa", "Send a signed Local Payments Africa payment DMN payload", resolvePaymentLocalPaymentsAfricaInputs)
}

func newSendPaymentAPMCommand(use, short string, resolve func(paymentPixFlags) (resolvedPaymentPixInputs, error)) *cobra.Command {
	flags := paymentPixFlags{}

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateStrictMode(cmd, flags); err != nil {
				return err
			}

			resolved, err := resolve(flags)
			if err != nil {
				return err
			}

			checkOpts := targetsafe.CheckOptions{
				TrustedProfiles: resolved.trustedProfiles,
				AllowUntrusted:  flags.allowUntrustedTarget,
				Confirmer:       targetsafe.NewConsoleConfirmer(),
			}
			classification, err := targetsafe.Check(resolved.targetURL, checkOpts)
			if err != nil {
				return safetyError(classification, err)
			}
			printTargetSummary(cmd, classification)

			_, err = verifyMerchantProfile(cmd.Context(), toCredentialProfile(resolved.merchantProfile))
			if err != nil {
				return err
			}

			result, err := sendDMNPayload(cmd.Context(), resolved.targetURL, resolved.payload.Encode())
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "HTTP status: %d\n", result.StatusCode)
			fmt.Fprintln(cmd.OutOrStdout(), "Response body:")
			fmt.Fprintln(cmd.OutOrStdout(), result.Body)
			return nil
		},
	}

	bindPaymentPixFlags(cmd, &flags, true)
	return cmd
}

func bindPaymentPixFlags(cmd *cobra.Command, flags *paymentPixFlags, includeAllowUntrusted bool) {
	cmd.Flags().StringVar(&flags.configPath, "config", "", "config file path (defaults to user config directory)")
	cmd.Flags().StringVar(&flags.profile, "profile", "", "merchant profile name")
	cmd.Flags().StringVar(&flags.status, "status", payment.StatusApproved, "payment status: PENDING, APPROVED, or DECLINED")
	cmd.Flags().StringVar(&flags.target, "target", "", "target profile name or absolute URL")
	cmd.Flags().StringVar(&flags.totalAmount, "total-amount", "", "override totalAmount")
	cmd.Flags().StringVar(&flags.currency, "currency", "", "override currency")
	cmd.Flags().StringVar(&flags.userPaymentOptionID, "user-payment-option-id", "", "override userPaymentOptionId")
	cmd.Flags().StringVar(&flags.pppTransactionID, "ppp-transaction-id", "", "override PPP_TransactionID")
	cmd.Flags().StringVar(&flags.transactionID, "transaction-id", "", "override TransactionID")
	cmd.Flags().StringVar(&flags.clientUniqueID, "client-unique-id", "", "override clientUniqueId")
	cmd.Flags().StringVar(&flags.clientRequestID, "client-request-id", "", "override clientRequestId")
	cmd.Flags().StringVar(&flags.reason, "reason", "", "override Reason")
	cmd.Flags().StringVar(&flags.reasonCode, "reason-code", "", "override ReasonCode")
	cmd.Flags().StringVar(&flags.customField1, "custom-field-1", "", "override customField1")
	cmd.Flags().StringVar(&flags.customField2, "custom-field-2", "", "override customField2")
	cmd.Flags().StringVar(&flags.customField3, "custom-field-3", "", "override customField3")
	cmd.Flags().StringVar(&flags.customField4, "custom-field-4", "", "override customField4")
	cmd.Flags().StringVar(&flags.customField5, "custom-field-5", "", "override customField5")
	cmd.Flags().StringVar(&flags.customField6, "custom-field-6", "", "override customField6")
	cmd.Flags().StringVar(&flags.customField7, "custom-field-7", "", "override customField7")
	cmd.Flags().StringVar(&flags.customField8, "custom-field-8", "", "override customField8")
	cmd.Flags().StringVar(&flags.customField9, "custom-field-9", "", "override customField9")
	cmd.Flags().StringVar(&flags.customField10, "custom-field-10", "", "override customField10")
	cmd.Flags().StringVar(&flags.customField11, "custom-field-11", "", "override customField11")
	cmd.Flags().StringVar(&flags.customField12, "custom-field-12", "", "override customField12")
	cmd.Flags().StringVar(&flags.customField13, "custom-field-13", "", "override customField13")
	cmd.Flags().StringVar(&flags.customField14, "custom-field-14", "", "override customField14")
	cmd.Flags().StringVar(&flags.customField15, "custom-field-15", "", "override customField15")
	cmd.Flags().BoolVar(&flags.requireCorrelationFields, "require-correlation-fields", false, "require explicit amount/currency/status and correlation IDs")
	if includeAllowUntrusted {
		cmd.Flags().BoolVar(&flags.allowUntrustedTarget, "allow-untrusted-target", false, "allow unknown public targets")
	}

	_ = cmd.MarkFlagRequired("profile")
	_ = cmd.MarkFlagRequired("target")
}

func bindPaymentFromRawFlags(cmd *cobra.Command, flags *paymentPixFlags, includeAllowUntrusted bool) {
	cmd.Flags().StringVar(&flags.configPath, "config", "", "config file path (defaults to user config directory)")
	cmd.Flags().StringVar(&flags.profile, "profile", "", "merchant profile name")
	cmd.Flags().StringVar(&flags.rawFile, "file", "", "path to URL-encoded DMN payload")
	cmd.Flags().BoolVar(&flags.keepRawMerchantFields, "keep-raw-merchant-fields", false, "keep merchant_id and merchant_site_id from the raw payload")
	cmd.Flags().StringVar(&flags.status, "status", "", "override payment status: PENDING, APPROVED, or DECLINED")
	cmd.Flags().StringVar(&flags.target, "target", "", "target profile name or absolute URL")
	cmd.Flags().StringVar(&flags.totalAmount, "total-amount", "", "override totalAmount")
	cmd.Flags().StringVar(&flags.currency, "currency", "", "override currency")
	cmd.Flags().StringVar(&flags.userPaymentOptionID, "user-payment-option-id", "", "override userPaymentOptionId")
	cmd.Flags().StringVar(&flags.pppTransactionID, "ppp-transaction-id", "", "override PPP_TransactionID")
	cmd.Flags().StringVar(&flags.transactionID, "transaction-id", "", "override TransactionID")
	cmd.Flags().StringVar(&flags.clientUniqueID, "client-unique-id", "", "override clientUniqueId")
	cmd.Flags().StringVar(&flags.clientRequestID, "client-request-id", "", "override clientRequestId")
	cmd.Flags().StringVar(&flags.reason, "reason", "", "override Reason")
	cmd.Flags().StringVar(&flags.reasonCode, "reason-code", "", "override ReasonCode")
	cmd.Flags().StringVar(&flags.customField1, "custom-field-1", "", "override customField1")
	cmd.Flags().StringVar(&flags.customField2, "custom-field-2", "", "override customField2")
	cmd.Flags().StringVar(&flags.customField3, "custom-field-3", "", "override customField3")
	cmd.Flags().StringVar(&flags.customField4, "custom-field-4", "", "override customField4")
	cmd.Flags().StringVar(&flags.customField5, "custom-field-5", "", "override customField5")
	cmd.Flags().StringVar(&flags.customField6, "custom-field-6", "", "override customField6")
	cmd.Flags().StringVar(&flags.customField7, "custom-field-7", "", "override customField7")
	cmd.Flags().StringVar(&flags.customField8, "custom-field-8", "", "override customField8")
	cmd.Flags().StringVar(&flags.customField9, "custom-field-9", "", "override customField9")
	cmd.Flags().StringVar(&flags.customField10, "custom-field-10", "", "override customField10")
	cmd.Flags().StringVar(&flags.customField11, "custom-field-11", "", "override customField11")
	cmd.Flags().StringVar(&flags.customField12, "custom-field-12", "", "override customField12")
	cmd.Flags().StringVar(&flags.customField13, "custom-field-13", "", "override customField13")
	cmd.Flags().StringVar(&flags.customField14, "custom-field-14", "", "override customField14")
	cmd.Flags().StringVar(&flags.customField15, "custom-field-15", "", "override customField15")
	if includeAllowUntrusted {
		cmd.Flags().BoolVar(&flags.allowUntrustedTarget, "allow-untrusted-target", false, "allow unknown public targets")
	}

	_ = cmd.MarkFlagRequired("profile")
	_ = cmd.MarkFlagRequired("file")
	_ = cmd.MarkFlagRequired("target")
}

type resolvedPaymentPixInputs struct {
	merchantProfile appconfig.MerchantProfile
	targetURL       string
	trustedProfiles map[string]targetsafe.Profile
	payload         payment.Payload
}

func resolvePaymentPixInputs(flags paymentPixFlags) (resolvedPaymentPixInputs, error) {
	return resolvePaymentAPMInputs(flags, apm.Pix)
}

func resolvePaymentBoletoInputs(flags paymentPixFlags) (resolvedPaymentPixInputs, error) {
	return resolvePaymentAPMInputs(flags, apm.Boleto)
}

func resolvePaymentCardInputs(flags paymentPixFlags) (resolvedPaymentPixInputs, error) {
	return resolvePaymentAPMInputs(flags, apm.Card)
}

func resolvePaymentLocalPaymentsAfricaInputs(flags paymentPixFlags) (resolvedPaymentPixInputs, error) {
	return resolvePaymentAPMInputs(flags, apm.LocalPaymentsAfrica)
}

func resolvePaymentAPMInputs(flags paymentPixFlags, build func(payment.Options) (payment.Payload, error)) (resolvedPaymentPixInputs, error) {
	configPath, err := resolveConfigPath(flags.configPath)
	if err != nil {
		return resolvedPaymentPixInputs{}, err
	}

	cfg, err := appconfig.Load(configPath)
	if err != nil {
		return resolvedPaymentPixInputs{}, err
	}

	merchantProfile, ok := cfg.Merchants[flags.profile]
	if !ok {
		return resolvedPaymentPixInputs{}, fmt.Errorf("merchant profile %q not found", flags.profile)
	}
	if err := appconfig.ValidateMerchantProfile(merchantProfile); err != nil {
		return resolvedPaymentPixInputs{}, err
	}

	targetURL, err := siminput.ResolveTargetURL(cfg, flags.target)
	if err != nil {
		return resolvedPaymentPixInputs{}, err
	}

	payload, err := build(payment.Options{
		MerchantID:          merchantProfile.MerchantID,
		MerchantSiteID:      merchantProfile.MerchantSiteID,
		MerchantSecretKey:   merchantProfile.MerchantSecretKey,
		Status:              strings.ToUpper(flags.status),
		TotalAmount:         flags.totalAmount,
		Currency:            flags.currency,
		UserPaymentOptionID: flags.userPaymentOptionID,
		PPPTransactionID:    flags.pppTransactionID,
		TransactionID:       flags.transactionID,
		ClientUniqueID:      flags.clientUniqueID,
		ClientRequestID:     flags.clientRequestID,
		Reason:              flags.reason,
		ReasonCode:          flags.reasonCode,
		CustomField1:        flags.customField1,
		CustomField2:        flags.customField2,
		CustomField3:        flags.customField3,
		CustomField4:        flags.customField4,
		CustomField5:        flags.customField5,
		CustomField6:        flags.customField6,
		CustomField7:        flags.customField7,
		CustomField8:        flags.customField8,
		CustomField9:        flags.customField9,
		CustomField10:       flags.customField10,
		CustomField11:       flags.customField11,
		CustomField12:       flags.customField12,
		CustomField13:       flags.customField13,
		CustomField14:       flags.customField14,
		CustomField15:       flags.customField15,
	})
	if err != nil {
		return resolvedPaymentPixInputs{}, err
	}

	return resolvedPaymentPixInputs{
		merchantProfile: merchantProfile,
		targetURL:       targetURL,
		trustedProfiles: siminput.TrustedTargetProfiles(cfg),
		payload:         payload,
	}, nil
}

func resolvePaymentFromRawInputs(flags paymentPixFlags) (resolvedPaymentPixInputs, error) {
	configPath, err := resolveConfigPath(flags.configPath)
	if err != nil {
		return resolvedPaymentPixInputs{}, err
	}

	cfg, err := appconfig.Load(configPath)
	if err != nil {
		return resolvedPaymentPixInputs{}, err
	}

	merchantProfile, ok := cfg.Merchants[flags.profile]
	if !ok {
		return resolvedPaymentPixInputs{}, fmt.Errorf("merchant profile %q not found", flags.profile)
	}
	if err := appconfig.ValidateMerchantProfile(merchantProfile); err != nil {
		return resolvedPaymentPixInputs{}, err
	}

	targetURL, err := siminput.ResolveTargetURL(cfg, flags.target)
	if err != nil {
		return resolvedPaymentPixInputs{}, err
	}

	rawPayloadBytes, err := os.ReadFile(flags.rawFile)
	if err != nil {
		return resolvedPaymentPixInputs{}, err
	}

	payload, err := payment.ParseEncoded(string(rawPayloadBytes))
	if err != nil {
		return resolvedPaymentPixInputs{}, err
	}

	if !flags.keepRawMerchantFields {
		payload.Fields[payment.FieldMerchantID] = merchantProfile.MerchantID
		payload.Fields[payment.FieldMerchantSiteID] = merchantProfile.MerchantSiteID
	}

	if status := strings.ToUpper(strings.TrimSpace(flags.status)); status != "" {
		pppStatus, err := payment.PPPStatusForStatus(status)
		if err != nil {
			return resolvedPaymentPixInputs{}, err
		}
		payload.Fields[payment.FieldStatus] = status
		payload.Fields[payment.FieldPPPStatus] = pppStatus
	}

	setOverride(payload.Fields, payment.FieldTotalAmount, flags.totalAmount)
	setOverride(payload.Fields, payment.FieldCurrency, flags.currency)
	setOverride(payload.Fields, payment.FieldUserPaymentOptionID, flags.userPaymentOptionID)
	setOverride(payload.Fields, payment.FieldPPPTransactionID, flags.pppTransactionID)
	setOverride(payload.Fields, payment.FieldTransactionID, flags.transactionID)
	setOverride(payload.Fields, payment.FieldClientUniqueID, flags.clientUniqueID)
	setOverride(payload.Fields, payment.FieldClientRequestID, flags.clientRequestID)
	setOverride(payload.Fields, payment.FieldReason, flags.reason)
	setOverride(payload.Fields, payment.FieldReasonCode, flags.reasonCode)
	setOverride(payload.Fields, payment.FieldCustomField1, flags.customField1)
	setOverride(payload.Fields, payment.FieldCustomField2, flags.customField2)
	setOverride(payload.Fields, payment.FieldCustomField3, flags.customField3)
	setOverride(payload.Fields, payment.FieldCustomField4, flags.customField4)
	setOverride(payload.Fields, payment.FieldCustomField5, flags.customField5)
	setOverride(payload.Fields, payment.FieldCustomField6, flags.customField6)
	setOverride(payload.Fields, payment.FieldCustomField7, flags.customField7)
	setOverride(payload.Fields, payment.FieldCustomField8, flags.customField8)
	setOverride(payload.Fields, payment.FieldCustomField9, flags.customField9)
	setOverride(payload.Fields, payment.FieldCustomField10, flags.customField10)
	setOverride(payload.Fields, payment.FieldCustomField11, flags.customField11)
	setOverride(payload.Fields, payment.FieldCustomField12, flags.customField12)
	setOverride(payload.Fields, payment.FieldCustomField13, flags.customField13)
	setOverride(payload.Fields, payment.FieldCustomField14, flags.customField14)
	setOverride(payload.Fields, payment.FieldCustomField15, flags.customField15)

	payload, err = payment.RecomputeAdvanceResponseChecksum(payload, merchantProfile.MerchantSecretKey)
	if err != nil {
		return resolvedPaymentPixInputs{}, err
	}

	return resolvedPaymentPixInputs{
		merchantProfile: merchantProfile,
		targetURL:       targetURL,
		trustedProfiles: siminput.TrustedTargetProfiles(cfg),
		payload:         payload,
	}, nil
}

func setOverride(fields map[string]string, key, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	fields[key] = value
}

func toCredentialProfile(profile appconfig.MerchantProfile) credentials.Profile {
	return credentials.Profile{
		Environment:       profile.Environment,
		MerchantID:        profile.MerchantID,
		MerchantSiteID:    profile.MerchantSiteID,
		MerchantSecretKey: profile.MerchantSecretKey,
	}
}

func printTargetSummary(cmd *cobra.Command, result targetsafe.Result) {
	fmt.Fprintf(cmd.OutOrStdout(), "Target host: %s\n", result.Host)
	fmt.Fprintf(cmd.OutOrStdout(), "Target classification: %s\n", result.Classification)
	fmt.Fprintf(cmd.OutOrStdout(), "Target reason: %s\n", result.Reason)
}

func printPayloadPreview(cmd *cobra.Command, payload payment.Payload) {
	fmt.Fprintln(cmd.OutOrStdout(), "Payload fields:")
	tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	for _, key := range orderedPayloadKeys(payload.Fields) {
		fmt.Fprintf(tw, "%s\t%s\n", key, payload.Fields[key])
	}
	_ = tw.Flush()

	fmt.Fprintln(cmd.OutOrStdout(), "Raw URL-encoded payload:")
	fmt.Fprintln(cmd.OutOrStdout(), payload.Encode())
}

func orderedPayloadKeys(fields map[string]string) []string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func safetyError(result targetsafe.Result, original error) error {
	if result.Host == "" {
		return original
	}
	return fmt.Errorf("target host %q is not allowed: %s", result.Host, result.Reason)
}

func validateStrictMode(cmd *cobra.Command, flags paymentPixFlags) error {
	if !flags.requireCorrelationFields {
		return nil
	}

	if strings.TrimSpace(flags.totalAmount) == "" {
		return fmt.Errorf("strict mode (--require-correlation-fields) requires --total-amount")
	}
	if strings.TrimSpace(flags.currency) == "" {
		return fmt.Errorf("strict mode (--require-correlation-fields) requires --currency")
	}
	if strings.TrimSpace(flags.userPaymentOptionID) == "" {
		return fmt.Errorf("strict mode (--require-correlation-fields) requires --user-payment-option-id")
	}
	if !cmd.Flags().Changed("status") {
		return fmt.Errorf("strict mode (--require-correlation-fields) requires explicit --status")
	}

	return nil
}
