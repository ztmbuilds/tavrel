package store

type ListOptions[C any] struct {
	Cursor *C
	Limit  int
}

type Page[T, C any] struct {
	Items      []T
	NextCursor *C
	HasMore    bool
}

func paginate[T any](list []T, limit int) (items []T, hasMore bool, last *T) {

	if len(list) <= limit || limit < 0 {
		return list, false, nil
	}
	return list[:limit], true, &list[limit-1]

}
