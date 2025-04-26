package test

type ResultType int

const (
	Fail ResultType = iota
	Pass
	Done
	Error
)

func (rt ResultType) String() string {
	return [...]string{"Fail", "Pass", "Done", "Error"}[rt]
}

type Result struct {
	Result  ResultType
	Message string
}
