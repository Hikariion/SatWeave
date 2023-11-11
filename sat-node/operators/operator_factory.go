package operators

type OperatorFactory func() OperatorBase

var FactoryMap = map[string]OperatorFactory{
	"KeyByInputOp": func() OperatorBase { return new(KeyByInputOp) },
}
