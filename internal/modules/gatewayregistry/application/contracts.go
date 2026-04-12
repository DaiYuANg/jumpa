package application

import (
	"context"

	gatewaydomain "github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry/domain"
	"github.com/samber/mo"
)

type GatewayService interface {
	List(ctx context.Context) ([]gatewaydomain.Gateway, error)
	Get(ctx context.Context, id string) (mo.Option[gatewaydomain.Gateway], error)
	RegisterHeartbeat(ctx context.Context, in RegisterHeartbeatInput) (gatewaydomain.Gateway, error)
	MarkOffline(ctx context.Context, nodeKey string) error
}

type RegisterHeartbeatInput struct {
	NodeKey       string
	NodeName      string
	RuntimeType   string
	AdvertiseAddr string
	SSHListenAddr string
	Zone          string
	Tags          []string
}
