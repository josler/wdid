package filter

type Filter interface {
	Match(i Matchable) (bool, error)
}

type Matchable interface {
	Data() string
	Status() string
	Datetime() int64
	Kind() int64
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
