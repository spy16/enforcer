package rule

import (
	"context"
	"sync"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

// New returns a fully-initialised rule engine instance.
func New() *Engine {
	return &Engine{vm: &vm.VM{}}
}

// Engine represents a rule engine and provides function for executing
// rules.
type Engine struct {
	mu sync.Mutex
	vm *vm.VM
}

// Exec executes a rule with given data as env and returns true if the
// result is truthy (non-nil and non-false).
func (en *Engine) Exec(_ context.Context, rule string, data interface{}) (bool, error) {
	en.mu.Lock()
	defer en.mu.Unlock()

	p, err := expr.Compile(rule, expr.Env(data))
	if err != nil {
		return false, err
	}

	out, err := en.vm.Run(p, data)
	if err != nil {
		return false, err
	}
	return isTruthy(out), nil
}

func isTruthy(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return v != nil
}
