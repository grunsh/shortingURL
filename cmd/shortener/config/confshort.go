package config

import (
	"flag"
)

var ServerAddress string
var ShortBaseURL string

func main() {
}

func init() {
	ServAddrParam := flag.String("a", "localhost:8080", "Host server address")
	ShortURLBaseParam := flag.String("b", "http://localhost:8080/", "Short base address")
	ServerAddress = *ServAddrParam
	ShortBaseURL = *ShortURLBaseParam
}
