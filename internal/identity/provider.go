package identity

import (
	"runtime"
	"strings"

	config2 "github.com/DaiYuANg/jumpa/internal/config"
)

type ProviderDescriptor struct {
	Kind               string   `json:"kind"`
	Platform           string   `json:"platform"`
	Backend            string   `json:"backend"`
	AuthenticationMode string   `json:"authenticationMode"`
	ProvisioningMode   string   `json:"provisioningMode"`
	SupportedFeatures  []string `json:"supportedFeatures"`
	Notes              []string `json:"notes,omitempty"`
}

func CurrentProvider(cfg config2.AppConfig) ProviderDescriptor {
	platform := runtime.GOOS
	kind := strings.ToLower(strings.TrimSpace(cfg.Identity.Provider))
	if kind == "" {
		kind = "local"
	}

	if kind == "os" || kind == "system" {
		return osProviderDescriptor(cfg, platform)
	}

	return ProviderDescriptor{
		Kind:               "local",
		Platform:           platform,
		Backend:            "application",
		AuthenticationMode: "application-password",
		ProvisioningMode:   "managed-in-db",
		SupportedFeatures: []string{
			"app-users",
			"rbac",
			"bastion-policies",
			"api-tokens",
		},
		Notes: []string{
			"Use this mode when bastion accounts should be managed independently from host OS accounts.",
		},
	}
}

func osProviderDescriptor(cfg config2.AppConfig, platform string) ProviderDescriptor {
	backend := strings.ToLower(strings.TrimSpace(cfg.Identity.OS.Backend))
	if backend == "" || backend == "auto" {
		switch platform {
		case "linux":
			backend = "pam"
		case "windows":
			backend = "winlogon"
		case "darwin":
			backend = "opendirectory"
		default:
			backend = platform
		}
	}

	return ProviderDescriptor{
		Kind:               "os",
		Platform:           platform,
		Backend:            backend,
		AuthenticationMode: "delegated-to-os",
		ProvisioningMode:   "jit-or-sync",
		SupportedFeatures: []string{
			"os-login",
			"jit-principal-mapping",
			"policy-binding",
			"session-audit",
		},
		Notes: []string{
			"Delegate authentication to PAM on Linux, local/domain accounts on Windows, or OpenDirectory on macOS, but keep bastion authorization, target mapping, and audit state inside the application database.",
			"OS groups can be imported or mapped, but they should not replace application-side bastion policies.",
		},
	}
}
