package test

type ResultType int

const (
	Fail ResultType = iota
	Pass
	Done
	Error
	InProgress
)

func (rt ResultType) String() string {
	return [...]string{"Fail", "Pass", "Done", "Error", "InProgress"}[rt]
}

type Result struct {
	Site    int
	Id      uint
	Label   string
	Result  ResultType
	Message string
	Retried int
}
