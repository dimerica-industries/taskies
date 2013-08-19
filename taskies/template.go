package taskies

import (
	"github.com/hoisie/mustache"
)

func template(str string, env *Env) string {
	ctxt := []interface{}{env.vals}

	for !env.IsRoot() {
		env = env.Parent()
		ctxt = append(ctxt, env.vals)
	}

	return mustache.Render(str, ctxt...)
}
