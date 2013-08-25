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

func compositeDecoder(td taskDecoderSet) taskDecoder {
	return func(data *taskData) (Task, error) {
		val := reflect.ValueOf(data.taskData)

		if val.Kind() != reflect.Slice {
			return nil, fmt.Errorf("CompositeDecoder expects slice, %s found", val.Kind())
		}

		tasks := make([]Task, val.Len())

		for i := 0; i < val.Len(); i++ {
			d := val.Index(i).Elem().Interface()
			task, err := td.decode(d)

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
}
