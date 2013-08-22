package src

import (
	"fmt"
	"reflect"
)

// Run multiple tasks in serial.
// In YAML, this is represented as a "- tasks:" block
type CompositeTask struct {
	*baseTask
	tasks []Task
}

func (t *CompositeTask) Run(ctxt RunContext) error {
	for _, t := range t.tasks {
		if res := ctxt.Run(t); res.error != nil {
			return res.error
		}
	}

	return nil
}

func compositeProvider(ps providerSet, data *taskData) (Task, error) {
	val := reflect.ValueOf(data.taskData)

	if val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("CompositeProvider expects slice, %s found", val.Kind())
	}

	tasks := make([]Task, val.Len())

	for i := 0; i < val.Len(); i++ {
		d := val.Index(i).Elem().Interface()
		task, err := ps.provide(d)

		if err != nil {
			return nil, err
		}

		tasks[i] = task
	}

	return &CompositeTask{
		baseTask: baseTaskFromTaskData(data),
		tasks:    tasks,
	}, nil
}
