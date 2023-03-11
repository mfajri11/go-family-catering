package config

import (
	"family-catering/pkg/utils"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	cfg Config
	tmp Config
)

func InitMock() {
	tmp = cfg
	cfg = MockConfig{}
}

func DestroyMock() {
	cfg = tmp
}

type (
	Config struct {
		path     string
		App      app      `yaml:"app"`
		Web      web      `yaml:"web"`
		Server   server   `yaml:"server"`
		Log      log      `yaml:"log"`
		Postgres postgres `yaml:"postgres"`
		Redis    redis    `yaml:"redis"`
		Mailer   mailer   `yaml:"mailer"`
	}

	app struct {
		Name    string `yaml:"name" env-required:"true"`
		Version string `yaml:"version" env-required:"true"`
	}

	web struct {
		PaginationLimit       int           `yaml:"pagination-limit" env-default:"10" env-layout:"int"`
		AllowedOrigins        []string      `yaml:"allowed-origins" env-default:"https://*, http://*" env-layout:"slice"`
		AllowedMethods        []string      `yaml:"allowed-methods" env-default:"GET,POST,PUT,DELETE,OPTIONS" env-layout:"slice"`
		AllowedHeaders        []string      `yaml:"allowed-headers" env-default:"Accept,Authorization,Content-Type" env-layout:"slice"`
		MaxAge                int           `yaml:"max-age"`
		GeneralRequestLimit   int           `yaml:"limit-general-request-per-minute"`
		AccessTokenSecretKey  string        `env:"SECRET_KEY_ACCESS_TOKEN"`
		RefreshTokenSecretKey string        `env:"SECRET_KEY_REFRESH_TOKEN"`
		AccessTokenTTL        time.Duration `yaml:"access-token-ttl" env-layout:"time.Duration"`
		RefreshTokenTTL       time.Duration `yaml:"refresh-token-ttl" env-layout:"time.Duration"`
	}

	server struct {
		Host            string        `yaml:"host" env-required:"true"`
		Port            int           `yaml:"port" env-default:"9000" env-layout:"int"`
		ReadTimeout     time.Duration `yaml:"read-timeout" env-default:"10s" env-layout:"time.Duration"`
		WriteTimeout    time.Duration `yaml:"write-timeout" env-default:"10s" env-layout:"time.Duration"`
		ShutDownTimeout time.Duration `yaml:"shutdown-timeout" env-default:"3s" env-layout:"time.Duration"`
	}

	log struct {
		Level string `yaml:"level" env-default:"info"`
	}

	postgres struct {
		Host                  string        `yaml:"host" env-required:"true"`
		Port                  int           `yaml:"port" env-required:"true"`
		Username              string        `env:"PG_USER"`
		Password              string        `env:"PG_PASSWORD" env-layout:"string"`
		DataBaseName          string        `env:"PG_DATABASE" env-layout:"string"`
		OpenConnection        int           `yaml:"max-open-connection" env-layout:"int"`
		IdleConnection        int           `yaml:"max-idle" env-layout:"int"`
		ConnectionMaxLifeTime time.Duration `yaml:"max-lifetime" env-layout:"time.Duration"`
	}

	redis struct {
		Host         string `yaml:"host" env-required:"true"`
		Port         string `yaml:"port" env-required:"true"`
		Password     string `env:"REDIS_PASSWORD" env-layout:"string"`
		DataBaseName int    `env:"database-name" env-default:"0" env-layout:"int"`
		PoolSize     int    `yaml:"pool-size" env-default:"10" env-layout:"int"`
		MaxRetries   int    `yaml:"max-retries" env-default:"3" env-layout:"int"`
	}

	mailer struct {
		Host                   string `yaml:"host" env-required:"true"`
		Port                   int    `yaml:"port" env-default:"1025" env-layout:"int"`
		Email                  string `env:"MAILER_EMAIL" env-layout:"string"`
		Password               string `env:"MAILER_PASSWORD" env-layout:"string"`
		SupportEmail           string `yaml:"support-email"`
		TemplateForgotPassword string `yaml:"template-forgot-password" env-default:"forgot_password_template.txt" env-layout:"string"`
		Identity               string `yaml:"identity"`
	}
)

func (s server) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (pg postgres) URL() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", pg.Username, pg.Password, pg.Host, pg.Port, pg.DataBaseName)
}

func (r redis) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

func (m mailer) Addr() string {
	return fmt.Sprintf("%s:%d", m.Host, m.Port)
}

func init() {

	env := utils.GetEnv("FCAT_ENV", "development")
	configFile := "config." + env + ".yaml"
	// little trick to change to working directory, really useful for testing
	// see https://brandur.org/fragments/testing-go-project-root
	cfg = Config{}
	_, fname, _, _ := runtime.Caller(0)
	dir := filepath.Join(fname, "..")
	cfg.path = dir
	configPath := filepath.Join(dir, configFile)
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		panic(err)
	}

}

func Cfg() *Config {
	return &cfg
}

func Path() string {
	return cfg.path
}

type MockConfig = Config
