package yajirobe

// Scanner Scanner
type Scanner interface {
	Scan() ([]*Stock, []*Fund, error)
}
