package operators

type KeyOperatorBase interface {
	Compute(data []byte) (string, error)
}
