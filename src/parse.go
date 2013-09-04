package src

import (
	"fmt"
	"launchpad.net/goyaml"
	"reflect"
)

var (
	invalidTopLevelType    = fmt.Errorf("Top level must be slice")
	invalidInstructionType = fmt.Errorf("Instruction must be a map")
	invalidInstructionLen  = fmt.Errorf("Instruction map must have only one key")
	invalidInstruction     = fmt.Errorf("Invalid instruction key found")
	invalidTaskType        = fmt.Errorf("Task must be a map")
	invalidTaskKey         = fmt.Errorf("Invalid task key found")
	invalidRunType         = fmt.Errorf("Run must be a map")
	invalidRunKey          = fmt.Errorf("Invalid run key found")
	invalidSetType         = fmt.Errorf("Set must be a map")
)

func parseBytes(contents []byte) (*ast, error) {
	var yaml interface{}

	err := goyaml.Unmarshal(contents, &yaml)

	if err != nil {
		return nil, err
	}

	yaml = clean(yaml)
	return parseYaml(yaml)
}

func parseYaml(data interface{}) (*ast, error) {
	ast := &ast{
		instructions: make([]instruction, 0),
	}

	if err := ast.decode(reflect.ValueOf(data)); err != nil {
		return nil, err
	}

	return ast, nil
}

type ast struct {
	instructions []instruction
}

func (a *ast) decode(data reflect.Value) error {
	k := data.Kind()

	if k != reflect.Slice {
		return invalidTopLevelType
	}

	l := data.Len()

	for i := 0; i < l; i++ {
		v := data.Index(i).Elem()

		if err := a.decodeInstruction(v); err != nil {
			return err
		}
	}

	return nil
}

func (a *ast) decodeInstruction(r reflect.Value) error {
	k := r.Kind()

    var (
        key string
        val reflect.Value
    )

    if k == reflect.String {
        key = "run"
        val = r
    } else {
        if k != reflect.Map {
            return invalidInstructionType
        }

        if r.Len() != 1 {
            return invalidInstructionLen
        }

        k := r.MapKeys()[0]
        key = k.String()
        val = r.MapIndex(k).Elem()
    }

	ins, err := decodeInstruction(key, val)

	if err != nil {
		return err
	}

	a.instructions = append(a.instructions, ins)

	return nil
}

func clean(val interface{}) interface{} {
	rv, ok := val.(reflect.Value)

	if !ok {
		rv = reflect.ValueOf(val)
	}

	for rv.Kind() == reflect.Interface || rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Map {
		m2 := make(map[string]interface{})
		keys := rv.MapKeys()

		for _, rk := range keys {
			ks := fmt.Sprintf("%v", rk.Interface())
			m2[ks] = clean(rv.MapIndex(rk))
		}

		return m2
	}

	if rv.Kind() == reflect.Slice {
		l := rv.Len()
		sl := make([]interface{}, l)

		for i := 0; i < l; i++ {
			sl[i] = clean(rv.Index(i))
		}

		return sl
	}

	return fmt.Sprintf("%v", rv.Interface())
}
