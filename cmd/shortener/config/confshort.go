package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

var ServerAddress string
var ShortBaseURL string

type Sconfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	PATH          string `env:"PATH"`
}

func main() {
}

func init() {
	var cfg Sconfig
	err := env.Parse(&cfg) // 👈 Parse environment variables into `Config`
	if err != nil {
		log.Fatalf("unable to parse ennvironment variables: %e", err)
	}
	ServAddr := flag.String("a", "localhost:8080", "Host server address")
	ShortURLBase := flag.String("b", "http://localhost:8080/", "Short base address")
	flag.Parse()
	ServerAddress = *ServAddr
	//	fmt.Println(ServerAddress)
	ShortBaseURL = *ShortURLBase
	//	fmt.Println(ShortBaseURL)
	//	fmt.Println(cfg)
	//	fmt.Println("Адрес сервера: ", cfg.ServerAddress, "URL: ", cfg.BaseURL)
}
