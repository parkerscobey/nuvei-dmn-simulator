package targetsafe

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type CheckOptions struct {
	TrustedProfiles map[string]Profile
	AllowUntrusted  bool
	Confirmer       Confirmer
}

func Check(targetURL string, opts CheckOptions) (Result, error) {
	deniedPrefixes := []string{}
	result := Classify(targetURL, opts.TrustedProfiles, deniedPrefixes)
	return result, RequireAllowed(result, opts.AllowUntrusted, opts.Confirmer)
}

type Confirmer interface {
	Confirm(prompt string) (bool, error)
}

type ConsoleConfirmer struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func NewConsoleConfirmer() *ConsoleConfirmer {
	return &ConsoleConfirmer{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

func (c *ConsoleConfirmer) Confirm(prompt string) (bool, error) {
	if c.Stdout == nil {
		c.Stdout = os.Stdout
	}
	if c.Stdin == nil {
		c.Stdin = os.Stdin
	}
	fmt.Fprintf(c.Stdout, "%s [y/N] ", prompt)
	reader := bufio.NewReader(c.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	switch answer {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func RequireAllowed(result Result, allowUntrusted bool, confirmer Confirmer) error {
	if result.Classification == ClassificationDenied {
		return fmt.Errorf("target %q is denied: %s", result.URL, result.Reason)
	}
	if result.Classification == ClassificationUnknown {
		return fmt.Errorf("target %q has unknown safety classification", result.URL)
	}
	if result.Classification == ClassificationUntrusted {
		if allowUntrusted {
			return nil
		}

		suggestion := " either add it as a trusted target with `config set-target <name>`, pass `--allow-untrusted-target`, or use a localhost/.test/.local URL"
		return fmt.Errorf("target %q (%s) is not allowed by default%s. Use --allow-untrusted-target to send anyway.", result.URL, result.Reason, suggestion)
	}

	if result.RequiresConfirm && confirmer != nil {
		prompt := fmt.Sprintf("Target %q (%s) is a production-hosted sandbox and may be a live endpoint. Send anyway?", result.URL, result.Reason)
		confirmed, err := confirmer.Confirm(prompt)
		if err != nil {
			return fmt.Errorf("confirm prompt: %w", err)
		}
		if !confirmed {
			return fmt.Errorf("send cancelled")
		}
		return nil
	}
	if result.RequiresConfirm {
		return fmt.Errorf("target %q (%s) requires confirmation before sending", result.URL, result.Reason)
	}

	return nil
}
