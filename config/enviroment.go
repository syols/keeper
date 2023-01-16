package config

import (
	"flag"
	"os"
)

type Variable struct {
	env   string
	name  string
	value *string
}

type EnvironmentVariables map[Variable]func(string) Option

func newVariable(env, name string) Variable {
	value := flag.String(name, "", env)
	return Variable{
		env:   env,
		name:  name,
		value: value,
	}
}

func NewEnvironmentVariables() EnvironmentVariables {
	return EnvironmentVariables{
		newVariable("SIGN", "sn"): withSign,
	}
}

func (e EnvironmentVariables) Options() (options []Option) {
	flag.Parse()
	for k, v := range e {
		variable, isOk := os.LookupEnv(k.env)
		if !isOk {
			variable = *k.value
		}

		if variable != "" {
			options = append(options, v(variable))
		}
	}
	return
}

func withSign(value string) Option {
	return func(s *Config) {
		s.Server.Sign = value
	}
}
