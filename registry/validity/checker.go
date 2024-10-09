package validity

type Checker interface {
	Check(schema, schemaType, mode string) (bool, error)
}

type CheckerFunc func(schema, schemaType, mode string) (bool, error)

func (f CheckerFunc) Check(schema, schemaType, mode string) (bool, error) {
	return f(schema, schemaType, mode)
}
