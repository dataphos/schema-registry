package compatibility

type Checker interface {
	Check(schema string, history []string, mode string) (bool, error)
}

type CheckerFunc func(schema string, history []string, mode string) (bool, error)

func (f CheckerFunc) Check(schema string, history []string, mode string) (bool, error) {
	return f(schema, history, mode)
}
