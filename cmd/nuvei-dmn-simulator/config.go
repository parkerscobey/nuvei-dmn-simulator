package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	appconfig "github.com/parkerscobey/nuvei-dmn-simulator/internal/config"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/credentials"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newConfigCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage merchant and target profiles",
	}
	cmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path (defaults to user config directory)")

	cmd.AddCommand(newConfigListCommand(&configPath))
	cmd.AddCommand(newConfigSetMerchantCommand(&configPath))
	cmd.AddCommand(newConfigSetTargetCommand(&configPath))
	cmd.AddCommand(newConfigVerifyCommand(&configPath))

	return cmd
}

func newConfigListCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured merchant and target profiles with secrets redacted",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveConfigPath(*configPath)
			if err != nil {
				return err
			}

			cfg, err := appconfig.Load(path)
			if err != nil {
				return err
			}

			formatted := appconfig.Format(appconfig.Redacted(cfg))
			if formatted == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "No merchant or target profiles configured.")
				return nil
			}

			fmt.Fprint(cmd.OutOrStdout(), formatted)
			return nil
		},
	}
}

func newConfigSetMerchantCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "set-merchant <profile>",
		Short: "Add or update a merchant profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			path, err := resolveConfigPath(*configPath)
			if err != nil {
				return err
			}

			cfg, err := appconfig.Load(path)
			if err != nil {
				return err
			}

			reader := bufio.NewReader(cmd.InOrStdin())
			out := cmd.OutOrStdout()

			profile := appconfig.MerchantProfile{}
			profile.Environment, err = promptString(out, reader, "Nuvei environment (test/prod)")
			if err != nil {
				return err
			}
			profile.MerchantID, err = promptString(out, reader, "Merchant ID")
			if err != nil {
				return err
			}
			profile.MerchantSiteID, err = promptString(out, reader, "Merchant Site ID")
			if err != nil {
				return err
			}
			profile.MerchantSecretKey, err = promptSecret(out, cmd.InOrStdin(), reader, "Merchant Secret Key")
			if err != nil {
				return err
			}

			if err := appconfig.ValidateMerchantProfile(profile); err != nil {
				return err
			}

			cfg.Merchants[profileName] = profile
			if err := appconfig.Save(path, cfg); err != nil {
				return err
			}

			fmt.Fprintf(out, "Saved merchant profile %q to %s\n", profileName, path)
			return nil
		},
	}
}

func newConfigSetTargetCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "set-target <name>",
		Short: "Add or update a trusted target profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetName := args[0]
			path, err := resolveConfigPath(*configPath)
			if err != nil {
				return err
			}

			cfg, err := appconfig.Load(path)
			if err != nil {
				return err
			}

			reader := bufio.NewReader(cmd.InOrStdin())
			out := cmd.OutOrStdout()

			target := appconfig.TargetProfile{}
			target.URL, err = promptString(out, reader, "Target URL")
			if err != nil {
				return err
			}
			target.Kind, err = promptString(out, reader, "Target kind (local/staging/sandbox/production-hosted-sandbox)")
			if err != nil {
				return err
			}
			target.RequiresConfirm, err = promptBool(out, reader, "Requires confirmation before send (true/false)")
			if err != nil {
				return err
			}

			if err := appconfig.ValidateTargetProfile(target); err != nil {
				return err
			}

			cfg.Targets[targetName] = target
			if err := appconfig.Save(path, cfg); err != nil {
				return err
			}

			fmt.Fprintf(out, "Saved target profile %q to %s\n", targetName, path)
			return nil
		},
	}
}

func newConfigVerifyCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "verify <profile>",
		Short: "Verify a merchant profile with Nuvei",
		Long: `Verify a merchant profile by calling Nuvei /getSessionToken.

This authenticates merchant credentials with Nuvei without sending a payment request
or opening an order. The merchant secret is loaded from local config and is never
printed by this command.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			path, err := resolveConfigPath(*configPath)
			if err != nil {
				return err
			}

			cfg, err := appconfig.Load(path)
			if err != nil {
				return err
			}

			profile, ok := cfg.Merchants[profileName]
			if !ok {
				return fmt.Errorf("merchant profile %q not found", profileName)
			}
			if err := appconfig.ValidateMerchantProfile(profile); err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			verification, err := credentials.NewClient(nil).Verify(ctx, credentials.Profile{
				Environment:       profile.Environment,
				MerchantID:        profile.MerchantID,
				MerchantSiteID:    profile.MerchantSiteID,
				MerchantSecretKey: profile.MerchantSecretKey,
			})
			if err != nil {
				return err
			}

			cacheNote := ""
			if verification.Cached {
				cacheNote = " (cached)"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Verified merchant profile %q with Nuvei %s environment using /getSessionToken%s.\n", profileName, verification.Environment, cacheNote)
			return nil
		},
	}
}

func resolveConfigPath(path string) (string, error) {
	if path != "" {
		return path, nil
	}

	return appconfig.DefaultPath()
}

func promptString(out io.Writer, reader *bufio.Reader, label string) (string, error) {
	fmt.Fprintf(out, "%s: ", label)
	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	return strings.TrimSpace(value), nil
}

func promptBool(out io.Writer, reader *bufio.Reader, label string) (bool, error) {
	value, err := promptString(out, reader, label)
	if err != nil {
		return false, err
	}

	switch strings.ToLower(value) {
	case "true", "t", "yes", "y", "1":
		return true, nil
	case "false", "f", "no", "n", "0", "":
		return false, nil
	default:
		return false, fmt.Errorf("expected true or false")
	}
}

func promptSecret(out io.Writer, in io.Reader, reader *bufio.Reader, label string) (string, error) {
	fmt.Fprintf(out, "%s: ", label)
	if file, ok := in.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		secret, err := term.ReadPassword(int(file.Fd()))
		fmt.Fprintln(out)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(secret)), nil
	}

	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	return strings.TrimSpace(value), nil
}
