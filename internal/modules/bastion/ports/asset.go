package ports

import (
	"context"
	"time"

	"github.com/samber/mo"
)

type HostRecord struct {
	ID                 string
	Name               string
	Address            string
	Port               int
	Protocol           string
	Environment        *string
	Platform           *string
	AuthenticationType string
	CredentialRef      *string
	JumpEnabled        bool
	RecordingPolicy    string
	CreatedAt          time.Time
}

type HostAccountRecord struct {
	ID                 string
	HostID             string
	AccountName        string
	AuthenticationType string
	CredentialRef      *string
	CreatedAt          time.Time
}

type CreateHostRecordInput struct {
	Name               string
	Address            string
	Port               int
	Protocol           string
	Environment        *string
	Platform           *string
	AuthenticationType string
	CredentialRef      *string
	JumpEnabled        bool
	RecordingPolicy    string
	CreatedAt          time.Time
}

type PatchHostRecordInput struct {
	Name               *string
	Address            *string
	Port               *int
	Protocol           *string
	Environment        *string
	Platform           *string
	AuthenticationType *string
	CredentialRef      *string
	JumpEnabled        *bool
	RecordingPolicy    *string
}

type CreateHostAccountRecordInput struct {
	HostID             string
	AccountName        string
	AuthenticationType string
	CredentialRef      *string
	CreatedAt          time.Time
}

type PatchHostAccountRecordInput struct {
	AccountName        *string
	AuthenticationType *string
	CredentialRef      *string
}

type HostRepository interface {
	ListHosts(ctx context.Context) ([]HostRecord, error)
	GetHostByID(ctx context.Context, id string) (mo.Option[HostRecord], error)
	GetHostByName(ctx context.Context, name string) (mo.Option[HostRecord], error)
	CreateHost(ctx context.Context, in CreateHostRecordInput) (HostRecord, error)
	UpdateHost(ctx context.Context, id string, in PatchHostRecordInput) (mo.Option[HostRecord], error)
	DeleteHost(ctx context.Context, id string) (bool, error)
}

type HostAccountRepository interface {
	GetHostAccountByID(ctx context.Context, hostID, accountID string) (mo.Option[HostAccountRecord], error)
	GetHostAccountByName(ctx context.Context, hostID, accountName string) (mo.Option[HostAccountRecord], error)
	ListHostAccountsByHostID(ctx context.Context, hostID string) ([]HostAccountRecord, error)
	CreateHostAccount(ctx context.Context, in CreateHostAccountRecordInput) (HostAccountRecord, error)
	UpdateHostAccount(ctx context.Context, hostID, accountID string, in PatchHostAccountRecordInput) (mo.Option[HostAccountRecord], error)
	DeleteHostAccount(ctx context.Context, hostID, accountID string) (bool, error)
}
