package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

// Переменная для парсинга переменных окружения. Они в приоритете.
var cfg struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

// Структура для параметров всего кода
type Parameters struct {
	ServerAddress   string
	ShortBaseURL    string
	FileStoragePath string
}

// FlagString Тип для определения имени параметра, значения по-умолчанию и описания использования
type FlagString struct {
	name     string
	defValue string
	usage    string
}

// Prms Структура, собирающая в себе все параметры
var Prms []struct {
	description string
	param       FlagString
}

func GetParams() Parameters {
	var p Parameters
	serverAddress := flag.String(Prms[0].param.name, Prms[0].param.defValue, Prms[0].param.usage)
	shortURLBaseParam := flag.String(Prms[1].param.name, Prms[1].param.defValue, Prms[1].param.usage)
	fileStoragePath := flag.String(Prms[2].param.name, Prms[2].param.defValue, Prms[2].param.usage)
	flag.Parse()
	p.ServerAddress = *serverAddress
	p.ShortBaseURL = *shortURLBaseParam
	p.FileStoragePath = *fileStoragePath
	err := env.Parse(&cfg) // Парсим переменные окружения
	if err != nil {
		log.Fatalf("Ну не получилось распарсить переменную окружения: %e", err)
	}
	if cfg.BaseURL != "" { // Если переменная окружения есть, используем её, иначе параметр или значение по-умолчанию
		p.ShortBaseURL = cfg.BaseURL
	}
	if cfg.ServerAddress != "" { // Если переменная окружения есть, используем её, иначе параметр или значение по-умолчанию
		p.ServerAddress = cfg.ServerAddress
	}
	if cfg.FileStoragePath != "" { // Если переменная окружения есть, используем её, иначе параметр или значение по-умолчанию
		p.FileStoragePath = cfg.FileStoragePath
	}
	if p.ShortBaseURL[len(p.ShortBaseURL)-1:] != "/" { // Накинем "/", т.к. в параметрах его не передают
		p.ShortBaseURL += "/"
	}
	return p
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
		{
			description: "Параметр файла хранения URL",
			param: FlagString{
				name:     "f",
				defValue: "short-url-db.json",
				usage:    "файл хранитель урлов",
			},
		},
	}
	if len(Prms) < 1 {
		panic("Ok")
	}
}
