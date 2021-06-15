package util

import "fmt"

type comboErr struct {
	fns []func() error
}

func NewComboErr() *comboErr {
	return &comboErr{}
}

func (e *comboErr) Add(fn func() error) {
	e.fns = append(e.fns, fn)
}

func (e comboErr) Run() error {
	var es []error
	for _, fn := range e.fns {
		if err := fn(); err != nil {
			es = append(es, err)
		}
	}
	if len(es) == 0 {
		return nil
	}
	return fmt.Errorf("%#v", es)
}
