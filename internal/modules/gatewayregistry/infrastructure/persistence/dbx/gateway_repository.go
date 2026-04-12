package dbx

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/DaiYuANg/arcgo/dbx"
	"github.com/DaiYuANg/arcgo/dbx/repository"
	"github.com/DaiYuANg/jumpa/internal/modules/gatewayregistry/ports"
	"github.com/samber/mo"
)

type gatewayRow struct {
	ID            int64     `dbx:"id"`
	NodeKey       string    `dbx:"node_key"`
	NodeName      string    `dbx:"node_name"`
	RuntimeType   string    `dbx:"runtime_type"`
	AdvertiseAddr string    `dbx:"advertise_addr"`
	SSHListenAddr string    `dbx:"ssh_listen_addr"`
	ZoneName      string    `dbx:"zone_name"`
	TagsCSV       *string   `dbx:"tags_csv"`
	State         string    `dbx:"state"`
	RegisteredAt  time.Time `dbx:"registered_at"`
	LastSeenAt    time.Time `dbx:"last_seen_at"`
	UpdatedAt     time.Time `dbx:"updated_at"`
}

type gatewaySchema struct {
	dbx.Schema[gatewayRow]
	ID            dbx.IDColumn[gatewayRow, int64, dbx.IDSnowflake] `dbx:"id,pk"`
	NodeKey       dbx.Column[gatewayRow, string]                   `dbx:"node_key"`
	NodeName      dbx.Column[gatewayRow, string]                   `dbx:"node_name"`
	RuntimeType   dbx.Column[gatewayRow, string]                   `dbx:"runtime_type"`
	AdvertiseAddr dbx.Column[gatewayRow, string]                   `dbx:"advertise_addr"`
	SSHListenAddr dbx.Column[gatewayRow, string]                   `dbx:"ssh_listen_addr"`
	ZoneName      dbx.Column[gatewayRow, string]                   `dbx:"zone_name"`
	TagsCSV       dbx.Column[gatewayRow, *string]                  `dbx:"tags_csv"`
	State         dbx.Column[gatewayRow, string]                   `dbx:"state"`
	RegisteredAt  dbx.Column[gatewayRow, time.Time]                `dbx:"registered_at"`
	LastSeenAt    dbx.Column[gatewayRow, time.Time]                `dbx:"last_seen_at"`
	UpdatedAt     dbx.Column[gatewayRow, time.Time]                `dbx:"updated_at"`
}

type gatewayRepo struct {
	gs   gatewaySchema
	repo *repository.Base[gatewayRow, gatewaySchema]
}

func NewGatewayRepository(db *dbx.DB) ports.GatewayRepository {
	gs := dbx.MustSchema("gateway_registry_nodes", gatewaySchema{})
	return &gatewayRepo{gs: gs, repo: repository.New[gatewayRow](db, gs)}
}

func (r *gatewayRepo) ListGateways(ctx context.Context) ([]ports.GatewayRecord, error) {
	rows, err := r.repo.ListSpec(ctx, repository.OrderBy(r.gs.NodeName.Asc(), r.gs.LastSeenAt.Desc()))
	if err != nil {
		return nil, err
	}
	return collectionx.MapList(rows, func(_ int, row gatewayRow) ports.GatewayRecord {
		return toGatewayRecord(row)
	}).Values(), nil
}

func (r *gatewayRepo) GetGatewayByID(ctx context.Context, id string) (mo.Option[ports.GatewayRecord], error) {
	value, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return mo.None[ports.GatewayRecord](), err
	}
	row, err := r.repo.FirstSpec(ctx, repository.Where(r.gs.ID.Eq(value)))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return mo.None[ports.GatewayRecord](), nil
		}
		return mo.None[ports.GatewayRecord](), err
	}
	return mo.Some(toGatewayRecord(row)), nil
}

func (r *gatewayRepo) GetGatewayByNodeKey(ctx context.Context, nodeKey string) (mo.Option[ports.GatewayRecord], error) {
	row, err := r.repo.FirstSpec(ctx, repository.Where(r.gs.NodeKey.Eq(nodeKey)))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return mo.None[ports.GatewayRecord](), nil
		}
		return mo.None[ports.GatewayRecord](), err
	}
	return mo.Some(toGatewayRecord(row)), nil
}

func (r *gatewayRepo) CreateGateway(ctx context.Context, in ports.CreateGatewayInput) (ports.GatewayRecord, error) {
	row := &gatewayRow{
		NodeKey:       in.NodeKey,
		NodeName:      in.NodeName,
		RuntimeType:   in.RuntimeType,
		AdvertiseAddr: in.AdvertiseAddr,
		SSHListenAddr: in.SSHListenAddr,
		ZoneName:      in.Zone,
		TagsCSV:       nilIfBlank(in.TagsCSV),
		State:         in.State,
		RegisteredAt:  in.RegisteredAt,
		LastSeenAt:    in.LastSeenAt,
		UpdatedAt:     in.UpdatedAt,
	}
	if err := r.repo.Create(ctx, row); err != nil {
		return ports.GatewayRecord{}, err
	}
	return toGatewayRecord(*row), nil
}

func (r *gatewayRepo) UpdateGatewayHeartbeat(ctx context.Context, id string, in ports.UpdateGatewayHeartbeatInput) (mo.Option[ports.GatewayRecord], error) {
	value, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return mo.None[ports.GatewayRecord](), err
	}
	res, err := r.repo.UpdateByID(ctx, value,
		r.gs.NodeName.Set(in.NodeName),
		r.gs.AdvertiseAddr.Set(in.AdvertiseAddr),
		r.gs.SSHListenAddr.Set(in.SSHListenAddr),
		r.gs.ZoneName.Set(in.Zone),
		r.gs.TagsCSV.Set(nilIfBlank(in.TagsCSV)),
		r.gs.State.Set(in.State),
		r.gs.LastSeenAt.Set(in.LastSeenAt),
		r.gs.UpdatedAt.Set(in.UpdatedAt),
	)
	if err != nil {
		return mo.None[ports.GatewayRecord](), err
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return mo.None[ports.GatewayRecord](), nil
	}
	return r.GetGatewayByID(ctx, id)
}

func (r *gatewayRepo) UpdateGatewayState(ctx context.Context, id, state string, updatedAt time.Time) (mo.Option[ports.GatewayRecord], error) {
	value, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return mo.None[ports.GatewayRecord](), err
	}
	res, err := r.repo.UpdateByID(ctx, value, r.gs.State.Set(state), r.gs.UpdatedAt.Set(updatedAt))
	if err != nil {
		return mo.None[ports.GatewayRecord](), err
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		return mo.None[ports.GatewayRecord](), nil
	}
	return r.GetGatewayByID(ctx, id)
}

func toGatewayRecord(row gatewayRow) ports.GatewayRecord {
	return ports.GatewayRecord{
		ID:            strconv.FormatInt(row.ID, 10),
		NodeKey:       row.NodeKey,
		NodeName:      row.NodeName,
		RuntimeType:   row.RuntimeType,
		AdvertiseAddr: row.AdvertiseAddr,
		SSHListenAddr: row.SSHListenAddr,
		Zone:          row.ZoneName,
		TagsCSV:       stringOrEmpty(row.TagsCSV),
		State:         row.State,
		RegisteredAt:  row.RegisteredAt,
		LastSeenAt:    row.LastSeenAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func nilIfBlank(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func stringOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
