//go:build linux && cgo

package identity

/*
#cgo LDFLAGS: -lpam
#include <stdlib.h>
#include <string.h>
#include <security/pam_appl.h>

typedef struct {
	const char* password;
} jumpa_pam_context;

static int jumpa_pam_conv(int num_msg, const struct pam_message **msg, struct pam_response **resp, void *appdata_ptr) {
	jumpa_pam_context *ctx = (jumpa_pam_context*)appdata_ptr;
	struct pam_response *responses = (struct pam_response*)calloc((size_t)num_msg, sizeof(struct pam_response));
	if (responses == NULL) {
		return PAM_BUF_ERR;
	}

	for (int i = 0; i < num_msg; i++) {
		switch (msg[i]->msg_style) {
		case PAM_PROMPT_ECHO_OFF:
		case PAM_PROMPT_ECHO_ON:
			responses[i].resp = strdup(ctx->password == NULL ? "" : ctx->password);
			if (responses[i].resp == NULL) {
				for (int j = 0; j < i; j++) {
					free(responses[j].resp);
				}
				free(responses);
				return PAM_BUF_ERR;
			}
			responses[i].resp_retcode = 0;
			break;
		case PAM_TEXT_INFO:
		case PAM_ERROR_MSG:
			responses[i].resp = NULL;
			responses[i].resp_retcode = 0;
			break;
		default:
			for (int j = 0; j <= i; j++) {
				free(responses[j].resp);
			}
			free(responses);
			return PAM_CONV_ERR;
		}
	}

	*resp = responses;
	return PAM_SUCCESS;
}

static int jumpa_pam_authenticate(const char *service, const char *username, const char *password, char **err_msg) {
	jumpa_pam_context ctx = { password };
	struct pam_conv conv = { jumpa_pam_conv, &ctx };
	pam_handle_t *pamh = NULL;

	int rc = pam_start(service, username, &conv, &pamh);
	if (rc != PAM_SUCCESS) {
		if (err_msg != NULL) {
			*err_msg = strdup("pam_start failed");
		}
		return rc;
	}

	rc = pam_authenticate(pamh, PAM_DISALLOW_NULL_AUTHTOK);
	if (rc == PAM_SUCCESS) {
		rc = pam_acct_mgmt(pamh, PAM_DISALLOW_NULL_AUTHTOK);
	}

	if (rc != PAM_SUCCESS && err_msg != NULL) {
		const char *msg = pam_strerror(pamh, rc);
		*err_msg = strdup(msg == NULL ? "pam authentication failed" : msg);
	}

	pam_end(pamh, rc);
	return rc;
}

static void jumpa_free_string(char *value) {
	if (value != NULL) {
		free(value);
	}
}
*/
import "C"

import (
	"context"
	"fmt"
	"strings"
	"unsafe"
)

type platformOSPasswordBackend struct {
	cfg osBackendConfig
}

func newOSPasswordBackend(cfg osBackendConfig) osPasswordBackend {
	return platformOSPasswordBackend{cfg: cfg}
}

func (b platformOSPasswordBackend) Name() string {
	return "pam"
}

func (b platformOSPasswordBackend) Available() bool {
	return true
}

func (b platformOSPasswordBackend) AuthenticatePassword(_ context.Context, provider ProviderDescriptor, credentials PasswordCredentials) (Authentication, error) {
	service := strings.TrimSpace(b.cfg.PAMService)
	if service == "" {
		service = "sshd"
	}

	cService := C.CString(service)
	cUsername := C.CString(credentials.Username)
	cPassword := C.CString(credentials.Password)
	defer func() {
		C.free(unsafe.Pointer(cService))
		C.free(unsafe.Pointer(cUsername))
		C.free(unsafe.Pointer(cPassword))
	}()

	var errMsg *C.char
	rc := C.jumpa_pam_authenticate(cService, cUsername, cPassword, &errMsg)
	defer C.jumpa_free_string(errMsg)

	if rc != C.PAM_SUCCESS {
		message := "pam authentication failed"
		if errMsg != nil {
			message = C.GoString(errMsg)
		}
		return Authentication{}, fmt.Errorf("%w: %s", ErrInvalidCredentials, message)
	}

	auth := newAuthentication(credentials.Username, provider, credentials.RemoteAddr)
	auth.Attributes.Set("pamService", service)
	return auth, nil
}
