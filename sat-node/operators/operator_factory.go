package operators

type OperatorFactory func() OperatorBase

var FactoryMap = map[string]OperatorFactory{
	"SimpleSource": func() OperatorBase { return new(SimpleSource) },
	"SimpleSink":   func() OperatorBase { return new(SinkOperator) },
	"KeyByInputOp": func() OperatorBase { return new(KeyByInputOp) },
}
