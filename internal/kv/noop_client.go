package kv

import (
	"context"
	"errors"
	"time"

	"github.com/DaiYuANg/arcgo/collectionx"
	"github.com/DaiYuANg/arcgo/kvx"
)

var errKVDisabled = errors.New("kv client is disabled")

type noopClient struct{}

type noopSubscription struct {
	ch chan []byte
}

type noopPipeline struct{}

func newNoopClient() kvx.Client {
	return &noopClient{}
}

func (c *noopClient) Get(_ context.Context, _ string) ([]byte, error) {
	return nil, kvx.ErrNil
}

func (c *noopClient) MGet(_ context.Context, _ []string) (map[string][]byte, error) {
	return map[string][]byte{}, nil
}

func (c *noopClient) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error {
	return nil
}

func (c *noopClient) MSet(_ context.Context, _ map[string][]byte, _ time.Duration) error {
	return nil
}

func (c *noopClient) Delete(_ context.Context, _ string) error {
	return nil
}

func (c *noopClient) DeleteMulti(_ context.Context, _ []string) error {
	return nil
}

func (c *noopClient) Exists(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (c *noopClient) ExistsMulti(_ context.Context, keys []string) (map[string]bool, error) {
	result := make(map[string]bool, len(keys))
	return result, nil
}

func (c *noopClient) Expire(_ context.Context, _ string, _ time.Duration) error {
	return nil
}

func (c *noopClient) TTL(_ context.Context, _ string) (time.Duration, error) {
	return 0, nil
}

func (c *noopClient) Scan(_ context.Context, _ string, _ uint64, _ int64) (collectionx.List[string], uint64, error) {
	return collectionx.NewList[string](), 0, nil
}

func (c *noopClient) Keys(_ context.Context, _ string) (collectionx.List[string], error) {
	return collectionx.NewList[string](), nil
}

func (c *noopClient) HGet(_ context.Context, _, _ string) ([]byte, error) {
	return nil, kvx.ErrNil
}

func (c *noopClient) HMGet(_ context.Context, _ string, _ []string) (map[string][]byte, error) {
	return map[string][]byte{}, nil
}

func (c *noopClient) HSet(_ context.Context, _ string, _ map[string][]byte) error {
	return nil
}

func (c *noopClient) HMSet(_ context.Context, _ string, _ map[string][]byte) error {
	return nil
}

func (c *noopClient) HGetAll(_ context.Context, _ string) (map[string][]byte, error) {
	return map[string][]byte{}, nil
}

func (c *noopClient) HDel(_ context.Context, _ string, _ ...string) error {
	return nil
}

func (c *noopClient) HExists(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

func (c *noopClient) HKeys(_ context.Context, _ string) (collectionx.List[string], error) {
	return collectionx.NewList[string](), nil
}

func (c *noopClient) HVals(_ context.Context, _ string) (collectionx.List[[]byte], error) {
	return collectionx.NewList[[]byte](), nil
}

func (c *noopClient) HLen(_ context.Context, _ string) (int64, error) {
	return 0, nil
}

func (c *noopClient) HIncrBy(_ context.Context, _, _ string, increment int64) (int64, error) {
	return increment, nil
}

func (c *noopClient) Publish(_ context.Context, _ string, _ []byte) error {
	return errKVDisabled
}

func (c *noopClient) Subscribe(_ context.Context, _ string) (kvx.Subscription, error) {
	return newNoopSubscription(), nil
}

func (c *noopClient) PSubscribe(_ context.Context, _ string) (kvx.Subscription, error) {
	return newNoopSubscription(), nil
}

func (c *noopClient) XAdd(_ context.Context, _, _ string, _ map[string][]byte) (string, error) {
	return "", errKVDisabled
}

func (c *noopClient) XRead(_ context.Context, _, _ string, _ int64) (collectionx.List[kvx.StreamEntry], error) {
	return collectionx.NewList[kvx.StreamEntry](), errKVDisabled
}

func (c *noopClient) XReadMultiple(_ context.Context, _ map[string]string, _ int64, _ time.Duration) (collectionx.MultiMap[string, kvx.StreamEntry], error) {
	return collectionx.NewMultiMap[string, kvx.StreamEntry](), errKVDisabled
}

func (c *noopClient) XRange(_ context.Context, _, _, _ string) (collectionx.List[kvx.StreamEntry], error) {
	return collectionx.NewList[kvx.StreamEntry](), errKVDisabled
}

func (c *noopClient) XRevRange(_ context.Context, _, _, _ string) (collectionx.List[kvx.StreamEntry], error) {
	return collectionx.NewList[kvx.StreamEntry](), errKVDisabled
}

func (c *noopClient) XLen(_ context.Context, _ string) (int64, error) {
	return 0, errKVDisabled
}

func (c *noopClient) XTrim(_ context.Context, _ string, _ int64) error {
	return errKVDisabled
}

func (c *noopClient) XDel(_ context.Context, _ string, _ []string) error {
	return errKVDisabled
}

func (c *noopClient) XGroupCreate(_ context.Context, _, _, _ string) error {
	return errKVDisabled
}

func (c *noopClient) XGroupDestroy(_ context.Context, _, _ string) error {
	return errKVDisabled
}

func (c *noopClient) XGroupCreateConsumer(_ context.Context, _, _, _ string) error {
	return errKVDisabled
}

func (c *noopClient) XGroupDelConsumer(_ context.Context, _, _, _ string) error {
	return errKVDisabled
}

func (c *noopClient) XReadGroup(_ context.Context, _, _ string, _ map[string]string, _ int64, _ time.Duration) (collectionx.MultiMap[string, kvx.StreamEntry], error) {
	return collectionx.NewMultiMap[string, kvx.StreamEntry](), errKVDisabled
}

func (c *noopClient) XAck(_ context.Context, _, _ string, _ []string) error {
	return errKVDisabled
}

func (c *noopClient) XPending(_ context.Context, _, _ string) (*kvx.PendingInfo, error) {
	return nil, errKVDisabled
}

func (c *noopClient) XPendingRange(_ context.Context, _, _, _, _ string, _ int64) (collectionx.List[kvx.PendingEntry], error) {
	return collectionx.NewList[kvx.PendingEntry](), errKVDisabled
}

func (c *noopClient) XClaim(_ context.Context, _, _, _ string, _ time.Duration, _ []string) (collectionx.List[kvx.StreamEntry], error) {
	return collectionx.NewList[kvx.StreamEntry](), errKVDisabled
}

func (c *noopClient) XAutoClaim(_ context.Context, _, _, _ string, _ time.Duration, _ string, _ int64) (string, collectionx.List[kvx.StreamEntry], error) {
	return "", collectionx.NewList[kvx.StreamEntry](), errKVDisabled
}

func (c *noopClient) XInfoGroups(_ context.Context, _ string) (collectionx.List[kvx.GroupInfo], error) {
	return collectionx.NewList[kvx.GroupInfo](), errKVDisabled
}

func (c *noopClient) XInfoConsumers(_ context.Context, _, _ string) (collectionx.List[kvx.ConsumerInfo], error) {
	return collectionx.NewList[kvx.ConsumerInfo](), errKVDisabled
}

func (c *noopClient) XInfoStream(_ context.Context, _ string) (*kvx.StreamInfo, error) {
	return nil, errKVDisabled
}

func (c *noopClient) Load(_ context.Context, _ string) (string, error) {
	return "", errKVDisabled
}

func (c *noopClient) Eval(_ context.Context, _ string, _ []string, _ [][]byte) ([]byte, error) {
	return nil, errKVDisabled
}

func (c *noopClient) EvalSHA(_ context.Context, _ string, _ []string, _ [][]byte) ([]byte, error) {
	return nil, errKVDisabled
}

func (c *noopClient) JSONSet(_ context.Context, _, _ string, _ []byte, _ time.Duration) error {
	return nil
}

func (c *noopClient) JSONGet(_ context.Context, _, _ string) ([]byte, error) {
	return nil, kvx.ErrNil
}

func (c *noopClient) JSONSetField(_ context.Context, _, _ string, _ []byte) error {
	return nil
}

func (c *noopClient) JSONGetField(_ context.Context, _, _ string) ([]byte, error) {
	return nil, kvx.ErrNil
}

func (c *noopClient) JSONDelete(_ context.Context, _, _ string) error {
	return nil
}

func (c *noopClient) CreateIndex(_ context.Context, _, _ string, _ []kvx.SchemaField) error {
	return errKVDisabled
}

func (c *noopClient) DropIndex(_ context.Context, _ string) error {
	return errKVDisabled
}

func (c *noopClient) Search(_ context.Context, _, _ string, _ int) (collectionx.List[string], error) {
	return collectionx.NewList[string](), errKVDisabled
}

func (c *noopClient) SearchWithSort(_ context.Context, _, _, _ string, _ bool, _ int) (collectionx.List[string], error) {
	return collectionx.NewList[string](), errKVDisabled
}

func (c *noopClient) SearchAggregate(_ context.Context, _, _ string, _ int) ([]map[string]any, error) {
	return nil, errKVDisabled
}

func (c *noopClient) Acquire(_ context.Context, _, _ string, _ time.Duration) (bool, error) {
	return false, errKVDisabled
}

func (c *noopClient) Release(_ context.Context, _, _ string) (bool, error) {
	return false, errKVDisabled
}

func (c *noopClient) Extend(_ context.Context, _, _ string, _ time.Duration) (bool, error) {
	return false, errKVDisabled
}

func (c *noopClient) Pipeline() kvx.Pipeline {
	return noopPipeline{}
}

func (c *noopClient) Close() error {
	return nil
}

func newNoopSubscription() kvx.Subscription {
	return &noopSubscription{ch: make(chan []byte)}
}

func (s *noopSubscription) Channel() <-chan []byte {
	return s.ch
}

func (s *noopSubscription) Close() error {
	close(s.ch)
	return nil
}

func (p noopPipeline) Enqueue(_ string, _ ...[]byte) error {
	return nil
}

func (p noopPipeline) Exec(_ context.Context) ([][]byte, error) {
	return [][]byte{}, nil
}

func (p noopPipeline) Close() error {
	return nil
}
