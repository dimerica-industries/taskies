package taskies

import (
	"github.com/hoisie/mustache"
)

func template(str string, env *Env) string {
	return mustache.Render(str, env.vals)
}
