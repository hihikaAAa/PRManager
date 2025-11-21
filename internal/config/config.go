package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
    Env string `yaml:"env" env-default:"local"`

    HTTPServer struct {
        Address      string        `yaml:"address" env-default:":8080"`
        ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"5s"`
        WriteTimeout time.Duration `yaml:"write_timeout" env-default:"5s"`
        IdleTimeout  time.Duration `yaml:"idle_timeout" env-default:"60s"`
    } `yaml:"http_server"`

    DB struct {
		DSN string `yaml:"dsn" env-required:"true"`
    } `yaml:"db"`
}

func MustLoad() *Config{
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("There is no CONFIG_PATH")
	}

	if _,err := os.Stat(configPath); os.IsNotExist(err){
		log.Fatalf("config file %s does not exist",configPath)
	}

	var cfg Config
	
	if err:= cleanenv.ReadConfig(configPath, &cfg);err!=nil{
		log.Fatalf("cannot read config file: %s", err)
	}

	return &cfg
}
