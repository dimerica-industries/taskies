package taskies

import (
    "io"
)

type Task func(env Env, in io.Reader, out, err io.Writer) error

func NewRunner(tasks map[string]Task, env Env, in io.Reader, out, err io.Writer) *Runner {
    return &Runner {
        tasks: tasks,
        env: env,
        in: in,
        out: out,
        err: err,
    }
}

type Runner struct {
    tasks map[string]Task
    env Env
    in io.Reader
    out io.Writer
    err io.Writer
}

func (r *Runner) Run() error {
    for _, t := range r.tasks {
        if err := t(r.env, r.in, r.out, r.err); err != nil {
            return err
        }
    }

    return nil
}
