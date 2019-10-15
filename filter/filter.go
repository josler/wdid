package filter

type Filter interface {
	Match(i interface{}) (bool, error)
}

type FilterComparison int

const (
	FilterEq FilterComparison = iota
	FilterNe
	FilterGt
	FilterLt
)

func (fc FilterComparison) String() string {
	switch fc {
	case FilterEq:
		return "="
	case FilterNe:
		return "!="
	case FilterGt:
		return ">"
	case FilterLt:
		return "<"
	}
	return ""
}
