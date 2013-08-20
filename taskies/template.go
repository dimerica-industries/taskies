package taskies

import (
	"github.com/dimerica-industries/taskies/mustache"
)

func template(str string, env *Env) string {
	ctxt := []interface{}{env.vals}

	for !env.IsRoot() {
		env = env.Parent()
		ctxt = append(ctxt, env.vals)
	}

	return mustache.Render(str, ctxt...)
}
