package filter

type Filter interface {
	Match(i interface{}) (bool, error)
}

type FilterComparison int

const (
	FilterEq FilterComparison = iota
	FilterNe
)
