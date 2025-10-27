package internal

type Set[T comparable] map[T]struct{}

func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

func NewSetFromSlice[T comparable](s []T) Set[T] {
	set := make(Set[T])
	for _, v := range s {
		set[v] = struct{}{}
	}

	return set
}

func (s Set[T]) Add(v T) {
	s[v] = struct{}{}
}

func (s Set[T]) Remove(v T) {
	delete(s, v)
}

func (s Set[T]) Has(v T) bool {
	_, ok := s[v]
	return ok
}

func (s Set[T]) ToSlice() []T {
	x := make([]T, 0, len(s))
	for e := range s {
		x = append(x, e)
	}

	return x
}

func (s *Set[T]) Clear() {
	*s = make(Set[T])
}
