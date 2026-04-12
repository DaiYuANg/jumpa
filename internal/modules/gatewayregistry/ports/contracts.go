package ports

import (
	"context"
	"time"

	"github.com/samber/mo"
)

type GatewayRecord struct {
	ID            string
	NodeKey       string
	NodeName      string
	RuntimeType   string
	AdvertiseAddr string
	SSHListenAddr string
	Zone          string
	TagsCSV       string
	State         string
	RegisteredAt  time.Time
	LastSeenAt    time.Time
	UpdatedAt     time.Time
}

type CreateGatewayInput struct {
	NodeKey       string
	NodeName      string
	RuntimeType   string
	AdvertiseAddr string
	SSHListenAddr string
	Zone          string
	TagsCSV       string
	State         string
	RegisteredAt  time.Time
	LastSeenAt    time.Time
	UpdatedAt     time.Time
}

type UpdateGatewayHeartbeatInput struct {
	NodeName      string
	AdvertiseAddr string
	SSHListenAddr string
	Zone          string
	TagsCSV       string
	State         string
	LastSeenAt    time.Time
	UpdatedAt     time.Time
}

type GatewayRepository interface {
	ListGateways(ctx context.Context) ([]GatewayRecord, error)
	GetGatewayByID(ctx context.Context, id string) (mo.Option[GatewayRecord], error)
	GetGatewayByNodeKey(ctx context.Context, nodeKey string) (mo.Option[GatewayRecord], error)
	CreateGateway(ctx context.Context, in CreateGatewayInput) (GatewayRecord, error)
	UpdateGatewayHeartbeat(ctx context.Context, id string, in UpdateGatewayHeartbeatInput) (mo.Option[GatewayRecord], error)
	UpdateGatewayState(ctx context.Context, id, state string, updatedAt time.Time) (mo.Option[GatewayRecord], error)
}
