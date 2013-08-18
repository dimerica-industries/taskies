package taskies

import (
    "strings"
)

func FromArray(arr []string) Env {
    env := make(Env)

    for _, v := range arr {
        parts := strings.SplitN(v, "=", 2)
        key := parts[0]
        val := ""

        if len(parts) == 2 {
            val = parts[1]
        }

        env[key] = val
    }

    return env
}

type Env map[string]string

func (v Env) Array() []string {
    a := make([]string, len(v))
    i := 0

    for k, v := range v {
        a[i] = k + "=" + v
        i++
    }

    return a
}

func MergeEnv(one Env, others ...Env) Env {
    env := make(Env)

    for k, v := range one {
        env[k] = v
    }

    for _, env2 := range others {
        for k, v := range env2 {
            env[k] = v
        }
    }

    return env
}
