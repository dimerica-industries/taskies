package src

import (
	"fmt"
	"launchpad.net/goyaml"
	"reflect"
	"strings"
)

type taskData struct {
	task        string
	name        string
	description string
	exportData  map[string]interface{}
	taskData    interface{}
}

func baseTaskFromTaskData(td *taskData) *baseTask {
	return &baseTask{
		typ:         td.task,
		name:        td.name,
		description: td.description,
		exportData:  []map[string]interface{}{td.exportData},
	}
}

type provider func(providerSet, *taskData) (Task, error)
type providerSet map[string]provider

func (ps providerSet) provide(data interface{}) (Task, error) {
	val := reflect.ValueOf(data)

	if val.Kind() == reflect.String {
		val = reflect.ValueOf(map[string]interface{}{
			data.(string): nil,
		})
	}

	if val.Kind() != reflect.Map {
		return nil, fmt.Errorf("Expecting map, found %s", val.Kind())
	}

	var (
		td   = &taskData{exportData: make(map[string]interface{})}
		keys = val.MapKeys()
	)

	for _, k := range keys {
		v := val.MapIndex(k).Elem()
		ks := k.String()

		switch ks {
		case "name":
			td.name = v.String()
		case "description":
			td.description = v.String()
		case "task":
			td.task = v.String()
		case "export":
			if v.Kind() != reflect.Map {
				return nil, fmt.Errorf("set section must be a map, %s found", v.Kind())
			}

			skeys := v.MapKeys()

			for _, sk := range skeys {
				sv := v.MapIndex(sk).Elem().Interface()
				sks := sk.String()

				td.exportData[sks] = sv
			}

		default:
			td.task = ks

			if v.IsValid() {
				td.taskData = v.Interface()
			}
		}
	}

	if td.task == "" {
		return nil, fmt.Errorf("No task provided")
	}

	prov, ok := ps[td.task]

	if !ok {
		return nil, fmt.Errorf("No task named \"%s\" found", td.task)
	}

	t, err := prov(ps, td)
	Debugf("[PROVIDER] [task=%#v] [data=%#v]", t, td)

	return t, err
}

func NewTaskSet() *TaskSet {
	return &TaskSet{
		Env: NewEnv(),
		providers: map[string]provider{
			"shell": shellProvider,
			"pipe":  pipeProvider,
			"tasks": compositeProvider,
		},
		Tasks:           make(map[string]Task),
		ExportedTasks:   make(map[string]Task),
		UnexportedTasks: make(map[string]Task),
	}
}

type TaskSet struct {
	Env             *Env
	providers       providerSet
	Tasks           map[string]Task
	ExportedTasks   map[string]Task
	UnexportedTasks map[string]Task
}

func DecodeYAML(contents []byte, ts *TaskSet) error {
	var data interface{}
	err := goyaml.Unmarshal(contents, &data)

	data = clean(data)

	if err != nil {
		return err
	}

	val := reflect.ValueOf(data)

	if val.Kind() != reflect.Map {
		return fmt.Errorf("Expecting map, found %s", val.Kind())
	}

	keys := val.MapKeys()

	for _, k := range keys {
		v := val.MapIndex(k).Elem()

		switch k.String() {
		case "tasks":
			err = decodeTasks(v, ts)

			if err != nil {
				return err
			}
		case "env":
			err = decodeEnv(v, ts)

			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("Invalid key: %s", k.Elem().String())
		}
	}

	return nil
}

func decodeEnv(val reflect.Value, ts *TaskSet) error {
	if val.Kind() != reflect.Map {
		return fmt.Errorf("Expecting map, found %s", val.Kind())
	}

	keys := val.MapKeys()

	for _, k := range keys {
		ks := k.String()
		vs := fmt.Sprintf("%v", val.MapIndex(k).Elem().Interface())

		ts.Env.Set(ks, vs)
	}

	return nil
}

func decodeTasks(val reflect.Value, ts *TaskSet) error {
	if val.Kind() != reflect.Slice {
		return fmt.Errorf("Expecting slice, found %s", val.Kind())
	}

	l := val.Len()

	for i := 0; i < l; i++ {
		tval := val.Index(i).Elem()

		if err := decodeTask(tval, ts); err != nil {
			return err
		}
	}

	return nil
}

func decodeTask(val reflect.Value, ts *TaskSet) error {
	t, err := ts.providers.provide(val.Interface())

	if err != nil {
		return err
	}

	name := t.Name()
	lname := strings.ToLower(name)

	if name == "" {
		return fmt.Errorf("No name found")
	}

	if name[0] == lname[0] {
		ts.UnexportedTasks[lname] = t
	} else {
		ts.ExportedTasks[lname] = t
	}

	ts.Tasks[lname] = t
	ts.providers[lname] = proxyProviderFunc(t)

	return nil
}

func clean(val interface{}) interface{} {
	if m, ok := val.(map[interface{}]interface{}); ok {
		m2 := make(map[string]interface{})

		for k, v := range m {
			ks := fmt.Sprintf("%v", k)
			m2[ks] = clean(v)
		}

		return m2
	}

	if sl, ok := val.([]interface{}); ok {
		for i, v := range sl {
			sl[i] = clean(v)
		}

		return sl
	}

	return fmt.Sprintf("%v", val)
}
