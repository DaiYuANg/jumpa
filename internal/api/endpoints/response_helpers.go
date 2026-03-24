package endpoints

func ok[T any](data T) Result[T] {
	return Result[T]{Success: true, Data: data}
}
