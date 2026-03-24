package endpoints

import (
	"strings"

	"github.com/samber/lo"
)

func parseIDsCSV(ids string) []string {
	return lo.FilterMap(strings.Split(ids, ","), func(p string, _ int) (string, bool) {
		t := strings.TrimSpace(p)
		return t, t != ""
	})
}

func paginate[T any](items []T, page, pageSize int) pageResponse[T] {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	total := len(items)
	start := (page - 1) * pageSize
	if start >= total {
		return pageResponse[T]{Items: []T{}, Total: total, Page: page, PageSize: pageSize}
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return pageResponse[T]{Items: items[start:end], Total: total, Page: page, PageSize: pageSize}
}

func containsFold(s, q string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(strings.TrimSpace(q)))
}
