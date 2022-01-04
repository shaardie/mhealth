package interfaces

type Check interface {
	Run() error
	Name() string
	Type() string
}
