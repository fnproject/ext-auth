package simple

type originalErrorer interface {
	Original() error
}

type authErr struct {
	code     int
	reason   string
	original error
}

func (ae authErr) Code() int {
	return ae.code
}

func (ae authErr) Error() string {
	return ae.reason
}

func (ae authErr) Original() error {
	return ae.original
}

type coder interface {
	Code() int
}
