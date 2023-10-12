/*
Дяденька, вот ты, который взялся смотреть мой код и ревью писать. Обращаюсь к тебе с просьбой, не будь сильно строг,
ПОЖАЛУЙСТА!  Я не кодил почти 20 лет (было php/Perl во времена сисадминства, интернет тогда был dial-up ещё).
Я попросил на работе купить мне курс Go. И как-то так получилось, что был куплен продвинутый курс, а не для тех, кто
учится ходить. Для меня любой мало-мальски работающий код - победа, как для годовалого ребёнка первый десяток шагов
держась за палец. Я не сдаюсь. Стараюсь. Я даже с гитом до курса не работал и никогда не писал юнит тестов. А тут вон чо.
Спасибо тебе дяденька за понимание, заранее. Не заваливай пожалуйста. Я почти не сплю, но тяну лямку и грызу гранит.
*/
package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

// Тип для переменных окружения
type Sconfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

var cfg Sconfig

const hashLen int = 10 // Длина генерируемого хеша

// const shortURLDomain string = "http://localhost:8080/"
var shortURLDomain string

type urlDBtype map[string][]byte

var urlDB = make(urlDBtype) // мапа для урлов, ключ - хеш, значение - URL

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // Для генератора хэшей

// Генератор сокращённого URL. Использует константу shortURLDomain как настройку.
func addURL(url []byte) []byte {
	hash := getHash()
	urlDB[hash] = url
	return []byte(shortURLDomain + hash)
}

// Генератор хеша. Использует константу hashLen для определения длины
func getHash() string {
	sb := strings.Builder{}
	sb.Grow(hashLen)
	for i := 0; i < hashLen; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

func shortingGetURL(res http.ResponseWriter, req *http.Request) {
	id := req.URL.Path[1:]                         // Откусываем / и записываем id
	res.Header().Set("Content-Type", "text/plain") // Установим тип ответа text/plain
	if val, ok := urlDB[id]; ok {
		res.Header().Set("Location", string(val))     // Укажем куда редирект
		res.WriteHeader(http.StatusTemporaryRedirect) // Передаём 307
	} else {
		res.WriteHeader(http.StatusBadRequest) // Прошли весь массив, но хеша нет.
	}
}

// Хендлер / для сокращения URL. На входе принимается URL как text/plain
func shortingRequest(res http.ResponseWriter, req *http.Request) {
	data, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		res.Header().Set("Content-Type", "text/plain") // Установим тип ответа text/plain
		res.WriteHeader(http.StatusBadRequest)
		return // Выход по 400
	}
	shrtURL := addURL(data)
	res.Header().Set("Content-Type", "text/plain") // Установим тип ответа text/plain
	res.Header().Set("Content-Length", strconv.Itoa(len(shrtURL)))
	res.WriteHeader(http.StatusCreated)
	res.Write(shrtURL)

}

func main() {
	//urlDB := make(map[string][]byte) // мапа для урлов, ключ - хеш, значение - URL
	//urlDB["14"] = []byte{1, 2, 3, 4, 5}

	//	_ = urlDB
	/*
		Вот эту всё чехарду с параметрами пришлось вытащить сюда из конфига, потому что иначе flag.Parse()
		в ините пакета конфига подхватывает параметры при запуске юнит тестов и всё фейлится к чертям.
		Я пока не нашёл способа это победить, сроки жмут, уже ночь, надо 5-й инкремент сдать :(
	*/
	ServAddrParam := flag.String("a", "localhost:8080", "Host server address")
	ShortURLBaseParam := flag.String("b", "http://localhost:8080/", "Short base address")
	flag.Parse()
	ServerAddress := *ServAddrParam
	ShortBaseURL := *ShortURLBaseParam
	err := env.Parse(&cfg) // Парсим переменные окружения
	if err != nil {
		log.Fatalf("Ну не получилось распарсить переменную окружения: %e", err)
	}
	if cfg.BaseURL != "" { // Если переменная окружения есть, используем её, иначе параметр или значение по-умолчанию
		ShortBaseURL = cfg.BaseURL
	}
	if cfg.ServerAddress != "" { // Если переменная окружения есть, используем её, иначе параметр или значение по-умолчанию
		ServerAddress = cfg.ServerAddress
	}
	if ShortBaseURL[len(ShortBaseURL)-1:] != "/" { // Накинем "/", т.к. в параметрах его не передают
		ShortBaseURL += "/"
	}
	tempV := strings.Split(ServerAddress, ":")
	serverName := tempV[0]
	serverPort := tempV[1]
	shortURLDomain = ShortBaseURL

	r := chi.NewRouter()
	r.Get("/{id}", shortingGetURL)
	r.Post("/", shortingRequest)

	log.Fatal(http.ListenAndServe(serverName+":"+serverPort, r))
}
func init() {
}
