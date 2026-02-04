package cmd

import (
	"errors"
	"net/url"
	"strings"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type EnvConfig struct {
	Environment  string `koanf:"ENVIRONMENT"`
	Port         int    `koanf:"PORT"`
	DatabaseURL  string `koanf:"GOOSE_DBSTRING"`
	ClientDomain string `koanf:"CLIENT_DOMAIN"`
	CookieDomain string `koanf:"COOKIE_DOMAIN"`
	CookieSecure bool   `koanf:"COOKIE_SECURE"`
}

var k = koanf.New(".")
var Env *EnvConfig

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
		v.Field(&env.Environment,
			v.Required,
			v.In("DEV", "PROD", "TEST"),
		),
		v.Field(&env.Port,
			v.Required,
			v.Min(1),
			v.Max(65535),
		),
		v.Field(&env.DatabaseURL,
			v.Required,
			v.By(func(value any) error {
				s, ok := value.(string)
				if !ok {
					return errors.New("database URL must be a string")
				}
				if !strings.HasPrefix(s, "postgres://") {
					return errors.New("database URL must start with 'postgres://'")
				}
				parsed, err := url.Parse(s)
				if err != nil || parsed.Scheme == "" || parsed.Host == "" {
					return errors.New("database URL must be a valid postgres URI")
				}
				return nil
			}),
		),
	)
}
