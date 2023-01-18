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
		newVariable("SIGN", "sn"):        withSign,
		newVariable("PRIVATE_KEY", "p"):  withPrivateKey,
		newVariable("SERVER_CERT", "cr"): withServerCert,
		newVariable("CLIENT_CERT", "c"):  withClientCert,
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

func withPrivateKey(value string) Option {
	return func(s *Config) {
		s.Server.PrivateKey = &value
	}
}

func withServerCert(value string) Option {
	return func(s *Config) {
		s.Server.Certificate = &value
	}
}

func withClientCert(value string) Option {
	return func(s *Config) {
		s.Client.Certificate = &value
	}
}
