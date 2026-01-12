package cmd

import (
	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type EnvConfig struct {
	Enviroment string `koanf:"ENVIRONMENT"`
	Port       int    `koanf:"Port"`
}

var k = koanf.New(".")

func LoadConfig() (*EnvConfig, error) {
	env := &EnvConfig{}

	if err := k.Load(file.Provider(".env"), dotenv.Parser()); err != nil {
		return nil, err
	}

	if err := k.Unmarshal("", env); err != nil {
		return nil, err
	}

	if err := validateConfig(env); err != nil {
		return nil, err
	}

	return env, nil
}

func validateConfig(env *EnvConfig) error {
	// PORT validation
	return v.ValidateStruct(env,
		v.Field(&env.Enviroment,
			v.Required,
			v.In("DEV", "PROD"),
		),
		v.Field(&env.Port,
			v.Required,
			v.Min(1),
			v.Max(65535),
		),
	)
}
