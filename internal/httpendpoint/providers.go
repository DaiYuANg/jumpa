package httpendpoint

import (
	"fmt"
	"sort"
	"strings"

	"github.com/arcgolabs/dix"
	dixadvanced "github.com/arcgolabs/dix/advanced"
	"github.com/arcgolabs/httpx"
)

const servicePrefix = "jumpa.http.endpoint."

func Provider0[T httpx.Endpoint](name string, fn func() T) dix.ProviderFunc {
	return dixadvanced.NamedProvider0[httpx.Endpoint](serviceName(name), func() httpx.Endpoint {
		return fn()
	})
}

func Provider1[T httpx.Endpoint, D1 any](name string, fn func(D1) T) dix.ProviderFunc {
	return dixadvanced.NamedProvider1[httpx.Endpoint, D1](serviceName(name), func(d1 D1) httpx.Endpoint {
		return fn(d1)
	})
}

func Provider2[T httpx.Endpoint, D1, D2 any](name string, fn func(D1, D2) T) dix.ProviderFunc {
	return dixadvanced.NamedProvider2[httpx.Endpoint, D1, D2](serviceName(name), func(d1 D1, d2 D2) httpx.Endpoint {
		return fn(d1, d2)
	})
}

func Provider3[T httpx.Endpoint, D1, D2, D3 any](name string, fn func(D1, D2, D3) T) dix.ProviderFunc {
	return dixadvanced.NamedProvider3[httpx.Endpoint, D1, D2, D3](serviceName(name), func(d1 D1, d2 D2, d3 D3) httpx.Endpoint {
		return fn(d1, d2, d3)
	})
}

func SliceProvider() dix.ProviderFunc {
	return dix.RawProviderWithMetadata(func(c *dix.Container) {
		dix.ProvideTErr[[]httpx.Endpoint](c, func() ([]httpx.Endpoint, error) {
			return ResolveAll(c)
		})
	}, dix.ProviderMetadata{
		Label:  "HTTPEndpointSliceProvider",
		Output: dix.TypedService[[]httpx.Endpoint](),
		Raw:    true,
	})
}

func ResolveAll(c *dix.Container) ([]httpx.Endpoint, error) {
	if c == nil || c.Raw() == nil {
		return nil, fmt.Errorf("httpendpoint: container is nil")
	}

	names := endpointServiceNames(c)
	endpoints := make([]httpx.Endpoint, 0, len(names))
	for _, name := range names {
		endpoint, err := dixadvanced.ResolveNamedAs[httpx.Endpoint](c, name)
		if err != nil {
			return nil, fmt.Errorf("httpendpoint: resolve %s: %w", name, err)
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

func serviceName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		panic("httpendpoint: endpoint provider name cannot be empty")
	}
	return servicePrefix + trimmed
}

func endpointServiceNames(c *dix.Container) []string {
	services := c.Raw().ListProvidedServices()
	names := make([]string, 0, len(services))
	for _, service := range services {
		if strings.HasPrefix(service.Service, servicePrefix) {
			names = append(names, service.Service)
		}
	}
	sort.Strings(names)
	return names
}
