package httpendpoint

import (
	"strings"

	"github.com/arcgolabs/dix"
	"github.com/arcgolabs/httpx"
)

func Provider0[T httpx.Endpoint](name string, fn func() T) dix.ProviderFunc {
	return dix.Contribute0[httpx.Endpoint](func() httpx.Endpoint {
		return fn()
	}, dix.Key(contributionKey(name)))
}

func Provider1[T httpx.Endpoint, D1 any](name string, fn func(D1) T) dix.ProviderFunc {
	return dix.Contribute1[httpx.Endpoint](func(d1 D1) httpx.Endpoint {
		return fn(d1)
	}, dix.Key(contributionKey(name)))
}

func Provider2[T httpx.Endpoint, D1, D2 any](name string, fn func(D1, D2) T) dix.ProviderFunc {
	return dix.Contribute2[httpx.Endpoint](func(d1 D1, d2 D2) httpx.Endpoint {
		return fn(d1, d2)
	}, dix.Key(contributionKey(name)))
}

func Provider3[T httpx.Endpoint, D1, D2, D3 any](name string, fn func(D1, D2, D3) T) dix.ProviderFunc {
	return dix.Contribute3[httpx.Endpoint](func(d1 D1, d2 D2, d3 D3) httpx.Endpoint {
		return fn(d1, d2, d3)
	}, dix.Key(contributionKey(name)))
}

func contributionKey(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		panic("httpendpoint: endpoint provider name cannot be empty")
	}
	return trimmed
}
