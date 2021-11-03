package main

type Body map[Pos]bool

func GroupBodies(s State, ps []Pos) []Body {
	m := map[Pos]Body{}
	for _, p := range ps {
		nb := Body{p: true}
		for _, np := range s.Around(p) {
			for bp := range m[np] {
				nb[bp] = true
				m[bp] = nb
			}
		}
		m[p] = nb
	}
	unique := map[Pos]void{}
	var ret []Body
outer:
	for _, b := range m {
		for p := range b {
			if _, found := unique[p]; found {
				continue outer
			}
			unique[p] = void{}
		}
		ret = append(ret, b)
	}
	return ret
}

func mapEquals(a, b map[Pos]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func (b Body) Any() Pos {
	for p := range b {
		return p
	}
	panic("empty Body")
}
