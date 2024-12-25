package fmi

type sMatch struct {
	query      []byte
	start, end int
}

type Stack []sMatch

func (s Stack) Empty() bool {
	return len(s) == 0
}

func (s Stack) Peek() sMatch {
	return s[len(s)-1]
}

func (s *Stack) Put(i sMatch) {
	(*s) = append((*s), i)
}

func (s *Stack) Pop() sMatch {
	d := (*s)[len(*s)-1]
	(*s) = (*s)[:len(*s)-1]
	return d
}
