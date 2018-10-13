package filter

type Filter interface {
	Match(i interface{}) (bool, error)
}
