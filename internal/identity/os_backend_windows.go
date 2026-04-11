//go:build windows

package identity

import (
	"context"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

type platformOSPasswordBackend struct{}

var (
	modAdvapi32   = windows.NewLazySystemDLL("advapi32.dll")
	procLogonUser = modAdvapi32.NewProc("LogonUserW")
)

func newOSPasswordBackend(_ osBackendConfig) osPasswordBackend {
	return platformOSPasswordBackend{}
}

func (platformOSPasswordBackend) Name() string {
	return "winlogon"
}

func (platformOSPasswordBackend) Available() bool {
	return true
}

func (platformOSPasswordBackend) AuthenticatePassword(_ context.Context, provider ProviderDescriptor, credentials PasswordCredentials) (Authentication, error) {
	username, domain := splitWindowsPrincipal(credentials.Username)

	userPtr, err := windows.UTF16PtrFromString(username)
	if err != nil {
		return Authentication{}, ErrInvalidCredentials
	}
	passwordPtr, err := windows.UTF16PtrFromString(credentials.Password)
	if err != nil {
		return Authentication{}, ErrInvalidCredentials
	}

	var domainPtr *uint16
	if strings.TrimSpace(domain) != "" {
		domainPtr, err = windows.UTF16PtrFromString(domain)
		if err != nil {
			return Authentication{}, ErrInvalidCredentials
		}
	}

	var token windows.Handle
	r1, _, _ := procLogonUser.Call(
		uintptr(unsafe.Pointer(userPtr)),
		uintptr(unsafe.Pointer(domainPtr)),
		uintptr(unsafe.Pointer(passwordPtr)),
		uintptr(3),
		uintptr(0),
		uintptr(unsafe.Pointer(&token)),
	)
	if r1 == 0 {
		return Authentication{}, ErrInvalidCredentials
	}
	defer windows.CloseHandle(token)

	auth := newAuthentication(username, provider, credentials.RemoteAddr)
	if domain != "" {
		auth.Attributes.Set("domain", domain)
	}
	return auth, nil
}

func splitWindowsPrincipal(input string) (username, domain string) {
	value := strings.TrimSpace(input)
	if value == "" {
		return "", ""
	}
	if strings.Contains(value, `\`) {
		parts := strings.SplitN(value, `\`, 2)
		return parts[1], parts[0]
	}
	if strings.Contains(value, "@") {
		parts := strings.SplitN(value, "@", 2)
		return parts[0], parts[1]
	}
	return value, "."
}
