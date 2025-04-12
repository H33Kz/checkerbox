package device

type Device interface {
	FunctionResolver()
	Print()
}
