package endpoints

func ok[T any](data T) Result[T] {
	return Result[T]{Success: true, Data: data}
}

func okPage[T any](items []T, total, page, pageSize int) Result[PageResult[T]] {
	return ok(PageResult[T]{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
