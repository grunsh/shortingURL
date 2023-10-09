package config

import (
	"flag"
	"fmt"
)

var ServerAddress string
var ShortBaseURL string

func main() {
}

func init() {
	ServAddr := flag.String("a", "localhost:8080", "Host server address")
	ShortURLBase := flag.String("b", "http://localhost:8080/", "Short base address")
	flag.Parse()
	ServerAddress = *ServAddr
	ShortBaseURL = *ShortURLBase
	fmt.Println("Адрес сервера: ", ServerAddress)
	fmt.Println("Базовый URL:", ShortBaseURL)
}
