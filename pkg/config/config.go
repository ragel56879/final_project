package config

import (
	"os"
)

type Config struct {
	TODO_PORT   string
	TODO_DBFILE string
}

func Load() *Config {
	return &Config{
		TODO_PORT:   os.Getenv("PORT"),
		TODO_DBFILE: os.Getenv("DBFILE"),
	}
}
