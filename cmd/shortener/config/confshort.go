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
	DataBaseDSN     string `env:"DATABASE_DSN"`
}

// Структура для параметров всего кода
type Parameters struct {
	ServerAddress   string
	ShortBaseURL    string
	FileStoragePath string
	DatabaseDSN     string
}

// FlagString Тип для определения имени параметра, значения по-умолчанию и описания использования
type FlagString struct {
	name     string
	defValue string
	usage    string
}

// Prms Структура, собирающая в себе все параметры
/*var Prms []struct {
	description string
	param       FlagString
}*/

var (
	Prms = map[string]struct {
		description string
		param       FlagString
	}{
		"servAddr": {
			description: "Параметр адреса сервера на котором он должен запуститься",
			param: FlagString{
				name:     "a",
				defValue: "localhost:8080",
				usage:    "Host server address",
			},
		},
		"baseUrl": {
			description: "Параметр базового сокращённого URL",
			param: FlagString{
				name:     "b",
				defValue: "http://localhost:8080/",
				usage:    "Short base address",
			},
		},
		"storageFile": {
			description: "Параметр файла хранения URL",
			param: FlagString{
				name:     "f",
				defValue: "short-url-db.json",
				usage:    "файл хранитель урлов",
			},
		},
		"dataBase": {
			description: "Строка подключения к БД",
			param: FlagString{
				name:     "d",
				defValue: "host=localhost user=shortener password=shortener dbname=shortener sslmode=disable",
				usage:    "Строка подключения к Бд формате:  host=%s user=%s password=%s dbname=%s sslmode=disable",
			},
		},
	}
)

func GetParams() Parameters {
	var p Parameters
	serverAddress := flag.String(Prms["servAddr"].param.name, Prms["servAddr"].param.defValue, Prms["servAddr"].param.usage)
	shortURLBaseParam := flag.String(Prms["baseUrl"].param.name, Prms["baseUrl"].param.defValue, Prms["baseUrl"].param.usage)
	fileStoragePath := flag.String(Prms["storageFile"].param.name, Prms["storageFile"].param.defValue, Prms["storageFile"].param.usage)
	dataBaseString := flag.String(Prms["dataBase"].param.name, Prms["dataBase"].param.defValue, Prms["dataBase"].param.usage)
	flag.Parse()
	p.ServerAddress = *serverAddress
	p.ShortBaseURL = *shortURLBaseParam
	p.FileStoragePath = *fileStoragePath
	p.DatabaseDSN = *dataBaseString
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
	if cfg.DataBaseDSN != "" { // Если переменная окружения есть, используем её, иначе параметр или значение по-умолчанию
		p.DatabaseDSN = cfg.DataBaseDSN
	}

	return p
}
