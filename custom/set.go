package custom

type Set[T comparable] map[T]struct{}

func (s Set[T]) Has(key T) bool {
	_, ok := s[key]
	return ok
}

func (s Set[T]) Add(key T) {
	s[key] = struct{}{}
}

func (s Set[T]) Del(key T) {
	delete(s, key)
}
