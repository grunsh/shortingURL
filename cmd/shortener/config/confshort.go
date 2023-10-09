package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"log"
)

var ServerAddress string
var ShortBaseURL string

type Cconfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

func main() {
}

func init() {
	var cfg Cconfig
	err := env.Parse(&cfg) // 👈 Parse environment variables into `Config`
	if err != nil {
		log.Fatalf("unable to parse ennvironment variables: %e", err)
	}
	ServAddr := flag.String("a", "localhost:8080", "Host server address")
	ShortURLBase := flag.String("b", "http://localhost:8080/", "Short base address")
	flag.Parse()
	ServerAddress = *ServAddr
	ShortBaseURL = *ShortURLBase
	fmt.Println("Адрес сервера: ", cfg.ServerAddress, "URL: ", cfg.BaseURL)
}
