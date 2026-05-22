package siminput

import (
	"fmt"

	appconfig "github.com/parkerscobey/nuvei-dmn-simulator/internal/config"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/targetsafe"
)

func ResolveTargetURL(cfg appconfig.Config, targetArg string) (string, error) {
	if profile, ok := cfg.Targets[targetArg]; ok {
		if err := appconfig.ValidateTargetProfile(profile); err != nil {
			return "", err
		}
		return profile.URL, nil
	}

	target := appconfig.TargetProfile{URL: targetArg, Kind: "trusted"}
	if err := appconfig.ValidateTargetProfile(target); err != nil {
		return "", fmt.Errorf("target %q is not a configured profile and is not a valid absolute URL", targetArg)
	}

	return targetArg, nil
}

func TrustedTargetProfiles(cfg appconfig.Config) map[string]targetsafe.Profile {
	profiles := make(map[string]targetsafe.Profile, len(cfg.Targets))
	for name, target := range cfg.Targets {
		profiles[name] = targetsafe.Profile{
			URL:             target.URL,
			Kind:            target.Kind,
			RequiresConfirm: target.RequiresConfirm,
		}
	}
	return profiles
}
