package src

import (
	"io"
)

type RunResult struct {
	out   []byte
	err   []byte
	error error
}

type RunContext interface {
	Env() *Env
	In() io.Reader
	Out() io.Writer
	Err() io.Writer
	Run(t Task) *RunResult
	Clone(*Env, io.Reader, io.Writer, io.Writer) RunContext
}

type Task interface {
	Type() string
	Name() string
	Description() string
	Run(RunContext) error
	ExportData() []map[string]interface{}
}

type baseTask struct {
	typ         string
	name        string
	description string
	exportData  []map[string]interface{}
}

func (t *baseTask) Type() string {
	return t.typ
}

func (t *baseTask) Name() string {
	return t.name
}

func (t *baseTask) Description() string {
	return t.description
}

func (t *baseTask) ExportData() []map[string]interface{} {
	return t.exportData
}
