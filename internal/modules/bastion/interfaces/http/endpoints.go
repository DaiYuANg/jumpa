package http

func boolOr(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func normalizePageRequest(page, pageSize int) (normalizedPage int, normalizedPageSize int, offset int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize, (page - 1) * pageSize
}
