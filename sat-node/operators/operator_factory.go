package operators

type OperatorFactory func() OperatorBase

var FactoryMap = map[string]OperatorFactory{
	"SimpleSource":    func() OperatorBase { return new(SimpleSource) },
	"SimpleSink":      func() OperatorBase { return new(SimpleSink) },
	"KeyByInputOp":    func() OperatorBase { return new(KeyByInputOp) },
	"FFTOp":           func() OperatorBase { return new(FFTOp) },
	"HaFilterOp":      func() OperatorBase { return new(HaFilterOp) },
	"HaSumOp":         func() OperatorBase { return new(HaSumOp) },
	"LaFilterOp":      func() OperatorBase { return new(LaFilterOp) },
	"LaSumOp":         func() OperatorBase { return new(LaSumOp) },
	"SimpleFFTSource": func() OperatorBase { return new(SimpleFFTSource) },
}
