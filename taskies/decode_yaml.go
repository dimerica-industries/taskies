package taskies

import (
	"fmt"
	"launchpad.net/goyaml"
	"reflect"
)

type namedTask struct {
	name string
	task Task
}

type provider func(providerSet, interface{}) (Task, error)
type providerSet map[string]provider

func (ps providerSet) provide(data interface{}) (*namedTask, error) {
	val := reflect.ValueOf(data)

	if val.Kind() != reflect.Map {
		return nil, fmt.Errorf("Expecting map, found%s", val.Kind())
	}

	keys := val.MapKeys()

	var (
		name     string
		task     string
		contents interface{}
	)

	for _, k := range keys {
		v := val.MapIndex(k).Elem()
		ks := k.Elem().String()

		switch ks {
		case "name":
			name = v.String()
		default:
			task = ks
			contents = v.Interface()
		}
	}

	if task == "" {
		return nil, fmt.Errorf("No task provided")
	}

	prov, ok := ps[task]

	if !ok {
		return nil, fmt.Errorf("No task named \"%s\" found", task)
	}

	t, err := prov(ps, contents)

	if err != nil {
		return nil, err
	}

	return &namedTask{
		name: name,
		task: t,
	}, nil
}

func NewTaskSet() *TaskSet {
	return &TaskSet{
		Env: NewEnv(),
		providers: map[string]provider{
			"shell": shellProvider,
			"pipe":  pipeProvider,
			"tasks": compositeProvider,
		},
		Tasks: make(map[string]Task),
	}
}

type TaskSet struct {
	Env       *Env
	providers providerSet
	Tasks     map[string]Task
}

func DecodeYAML(contents []byte, ts *TaskSet) error {
	var data interface{}
	err := goyaml.Unmarshal(contents, &data)

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

		switch k.Elem().String() {
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
		ts.Env.Set(k.Elem().String(), fmt.Sprintf("%v", val.MapIndex(k).Elem().Interface()))
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

	if t.name == "" {
		return fmt.Errorf("No name found")
	}

	ts.Tasks[t.name] = t.task
	ts.providers[t.name] = func(ps providerSet, data interface{}) (Task, error) {
		return t.task, nil
	}

	return nil
}
