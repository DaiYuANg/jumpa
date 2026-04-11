//go:build darwin && cgo

package identity

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Foundation -framework OpenDirectory
#include <stdlib.h>
#include <Foundation/Foundation.h>
#include <OpenDirectory/OpenDirectory.h>

static int jumpa_od_authenticate(const char *node_name, const char *username, const char *password, char **err_msg) {
	@autoreleasepool {
		NSString *node = [NSString stringWithUTF8String:(node_name == NULL || node_name[0] == '\0') ? "/Search" : node_name];
		NSString *user = [NSString stringWithUTF8String:username == NULL ? "" : username];
		NSString *pass = [NSString stringWithUTF8String:password == NULL ? "" : password];

		NSError *error = nil;
		ODSession *session = [ODSession defaultSession];
		ODNode *odNode = [ODNode nodeWithSession:session name:node error:&error];
		if (odNode == nil) {
			if (err_msg != NULL && error != nil) {
				*err_msg = strdup([[error localizedDescription] UTF8String]);
			}
			return 1;
		}

		ODRecord *record = [odNode recordWithRecordType:kODRecordTypeUsers name:user attributes:nil error:&error];
		if (record == nil) {
			if (err_msg != NULL && error != nil) {
				*err_msg = strdup([[error localizedDescription] UTF8String]);
			}
			return 2;
		}

		BOOL ok = [record verifyPassword:pass error:&error];
		if (!ok) {
			if (err_msg != NULL && error != nil) {
				*err_msg = strdup([[error localizedDescription] UTF8String]);
			}
			return 3;
		}

		return 0;
	}
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
	return "opendirectory"
}

func (b platformOSPasswordBackend) Available() bool {
	return true
}

func (b platformOSPasswordBackend) AuthenticatePassword(_ context.Context, provider ProviderDescriptor, credentials PasswordCredentials) (Authentication, error) {
	node := strings.TrimSpace(b.cfg.DirectoryNode)
	if node == "" {
		node = "/Search"
	}

	cNode := C.CString(node)
	cUsername := C.CString(credentials.Username)
	cPassword := C.CString(credentials.Password)
	defer func() {
		C.free(unsafe.Pointer(cNode))
		C.free(unsafe.Pointer(cUsername))
		C.free(unsafe.Pointer(cPassword))
	}()

	var errMsg *C.char
	rc := C.jumpa_od_authenticate(cNode, cUsername, cPassword, &errMsg)
	defer C.jumpa_free_string(errMsg)

	if rc != 0 {
		message := "OpenDirectory authentication failed"
		if errMsg != nil {
			message = C.GoString(errMsg)
		}
		return Authentication{}, fmt.Errorf("%w: %s", ErrInvalidCredentials, message)
	}

	auth := newAuthentication(credentials.Username, provider, credentials.RemoteAddr)
	auth.Attributes.Set("directoryNode", node)
	return auth, nil
}
