package src

import (
	"fmt"
	"reflect"
	"strings"
)

type taskDecoder func(*taskData) (Task, error)
type taskDecoderSet map[string]taskDecoder

func defaultTaskDecoders() taskDecoderSet {
	ts := make(taskDecoderSet)

	ts["shell"] = shellDecoder
	ts["pipe"] = pipeDecoder(ts)
	ts["tasks"] = compositeDecoder(ts)

	return ts
}

func newDecoder(ns *Namespace) *decoder {
	return &decoder{
		ns:           ns,
		taskDecoders: defaultTaskDecoders(),
	}
}

type decoder struct {
	ns           *Namespace
	taskDecoders taskDecoderSet
}

type decodeFn func(reflect.Value) error

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

func (ps taskDecoderSet) decode(data interface{}) (Task, error) {
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

	t, err := prov(td)
	Debugf("[TASK DECODER] [task=%#v] [data=%#v]", t, td)

	return t, err
}

func (d *decoder) decode(val reflect.Value) error {
	if val.Kind() != reflect.Map {
		return fmt.Errorf("Expecting map, found %s", val.Kind())
	}

	keys := val.MapKeys()

	for _, k := range keys {
		v := val.MapIndex(k).Elem()
		var fn decodeFn

		switch k.String() {
		case "include":
			fn = d.decodeIncludes
		case "tasks":
			fn = d.decodeTasks
		case "env":
			fn = d.decodeEnv
		default:
			return fmt.Errorf("Invalid key: %s", k.Elem().String())
		}

		err := fn(v)

		if err != nil {
			return err
		}
	}

	return nil
}

func (d *decoder) decodeIncludes(val reflect.Value) error {
	return nil
}

func (d *decoder) decodeEnv(val reflect.Value) error {
	if val.Kind() != reflect.Map {
		return fmt.Errorf("Expecting map, found %s", val.Kind())
	}

	keys := val.MapKeys()

	for _, k := range keys {
		ks := k.String()
		vs := fmt.Sprintf("%v", val.MapIndex(k).Elem().Interface())

		d.ns.env.Set(ks, vs)
	}

	return nil
}

func (d *decoder) decodeTasks(val reflect.Value) error {
	if val.Kind() != reflect.Slice {
		return fmt.Errorf("Expecting slice, found %s", val.Kind())
	}

	l := val.Len()

	for i := 0; i < l; i++ {
		tval := val.Index(i).Elem()

		if err := d.decodeTask(tval); err != nil {
			return err
		}
	}

	return nil
}

func (d *decoder) decodeTask(val reflect.Value) error {
	t, err := d.taskDecoders.decode(val.Interface())

	if err != nil {
		return err
	}

	name := t.Name()
	lname := strings.ToLower(name)

	if name == "" {
		return fmt.Errorf("No name found")
	}

	d.ns.AddTask(t)
	d.taskDecoders[lname] = proxyDecoder(t)

	return nil
}
