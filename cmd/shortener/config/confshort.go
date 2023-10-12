package config

var ServerAddress string
var ShortBaseURL string

type FlagString struct {
	name     string
	defValue string
	usage    string
}

var Prms []struct {
	description string
	param       FlagString
}

func init() {
	Prms = []struct {
		description string
		param       FlagString
	}{
		{
			description: "Параметр адреса сервера на котором он должен запуститься",
			param: FlagString{
				name:     "a",
				defValue: "localhost:8080",
				usage:    "Host server address",
			},
		},
		{
			description: "Параметр базового сокращённого URL",
			param: FlagString{
				name:     "b",
				defValue: "http://localhost:8080/",
				usage:    "Short base address",
			},
		},
	}
	if len((Prms)) < 1 {
		panic("Ok")
	}
}
