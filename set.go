package findctx

type set[T comparable] map[T]struct{}

func newSet[T comparable]() set[T] {
	return make(set[T])
}

func (s set[T]) add(v T) {
	s[v] = struct{}{}
}

func (s set[T]) remove(v T) {
	delete(s, v)
}

func (s set[T]) has(v T) bool {
	_, ok := s[v]
	return ok
}

func (s set[T]) toSlice() []T {
	x := make([]T, 0, len(s))
	for e := range s {
		x = append(x, e)
	}

	return x
}

func (s *set[T]) clear() {
	*s = make(set[T])
}
