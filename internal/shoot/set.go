package shoot

type Set[T comparable] map[T]struct{}

func MakeSet[T comparable](elems ...T) Set[T] {
	s := make(map[T]struct{}, len(elems))
	for _, e := range elems {
		s[e] = struct{}{}
	}
	return s
}

func (s Set[T]) Adds(elem T) {
	s[elem] = struct{}{}
}

func (s Set[T]) Has(elem T) bool {
	_, ok := s[elem]
	return ok
}

// func (s Set[T]) Removes(elem T) {
// 	delete(s, elem)
// }
