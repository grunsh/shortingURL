package config

import (
	"flag"
)

var ServerAddress string
var ShortBaseURL string

func main() {
}

func init() {
	ServAddr := flag.String("a", "localhost:8080", "Host server address")
	ShortBaseURL := flag.String("b", "http://localhost:8080/", "Short base address")
	flag.Parse()
	ServerAddress := *ServAddr
	ShortURLBase := *ShortBaseURL
	if ServerAddress == "" {
		panic("ой")
	}
	if ShortURLBase == "" {
		panic("ой")
	}
}
