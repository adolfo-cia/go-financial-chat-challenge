package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDriver            string
	DBSource            string
	ServerPort          string
	TokenSymmetricKey   string
	AccessTokenDuration time.Duration
	RabbitUrl           string
}

func Load() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("error loading .env file", err)
	}

	accesTokenDuration, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_DURATION"))
	if err != nil {
		log.Fatalln("error loading ACCESS_TOKEN_DURATION", err)
	}

	return &Config{
		DBDriver:            os.Getenv("DB_DRIVER"),
		DBSource:            os.Getenv("DB_SOURCE"),
		ServerPort:          os.Getenv("SERVER_PORT"),
		TokenSymmetricKey:   os.Getenv("TOKEN_SYMMETRIC_KEY"),
		AccessTokenDuration: accesTokenDuration,
		RabbitUrl:           os.Getenv("RABBIT_URL")}

}
