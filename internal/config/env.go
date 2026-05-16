package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

func LoadDotEnv() error {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}
