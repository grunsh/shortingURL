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
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net/http"
	"shortingURL/cmd/shortener/config"
	"strconv"
	"strings"
	"time"
)

// Тип для переменных окружения
type Sconfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

type urlDBtype map[string][]byte

var (
	cfg            Sconfig           // Переменная для объекта конфигурирования
	shortURLDomain string            // Переменная используется в коде в разных местах, значение присваивается в начале работы их cfg
	urlDB          = make(urlDBtype) // мапа для урлов, ключ - хеш, значение - URL
	sugar          zap.SugaredLogger // регистратор журналов
)

const hashLen int = 10 // Длина генерируемого хеша

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // Для генератора хэшей

type (
	// структура для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}

	// Тип с интерфейсом ResponseWriter для компрессии
	gzipWriter struct {
		http.ResponseWriter
		Writer io.Writer
	}
)

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

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

// Хендлер получения сокращённого URL. 307 и редирект, или ошибка.
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

func shortingJSON(res http.ResponseWriter, req *http.Request) {
	type URLReq struct { // Тип для запроса с тегом url
		URL string `json:"url"`
	}
	type URLResp struct { // Тип для ответа с тегом result
		Result string `json:"result"`
	}
	var reqURL URLReq
	var respURL URLResp
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body) // Чтение тела запроса в буфер buf
	if err != nil {
		res.Header().Set("Content-Type", "application/json") // Установим тип ответа application/json
		res.WriteHeader(http.StatusBadRequest)
		return // Выход по 400 (ошибка чтения тела запроса)
	}
	if err = json.Unmarshal(buf.Bytes(), &reqURL); err != nil {
		res.Header().Set("Content-Type", "application/json") // Установим тип ответа application/json
		res.WriteHeader(http.StatusBadRequest)
		return // Выход по 400
	}
	respURL.Result = string(addURL([]byte(reqURL.URL)))
	resp, err := json.Marshal(respURL)
	if err != nil {
		res.Header().Set("Content-Type", "application/json") // Установим тип ответа application/json
		res.WriteHeader(http.StatusBadRequest)
		return // Выход по 400
	}
	res.Header().Set("Content-Type", "application/json") // Установим тип ответа application/json
	res.Header().Set("Content-Length", strconv.Itoa(len(resp)))
	res.WriteHeader(http.StatusCreated)
	res.Write(resp)
}

// Обёртка для журналирования запросов
func logHTTPInfo(h http.Handler) http.Handler {
	logHTTPRequests := func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r) // внедряем реализацию http.ResponseWriter
		sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", time.Since(startTime),
			"size", responseData.size,
		)
	}
	return http.HandlerFunc(logHTTPRequests)
}

func compressExchange(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// проверяем, что клиент поддерживает gzip-сжатие
		// это упрощённый пример. В реальном приложении следует проверять все
		// значения r.Header.Values("Accept-Encoding") и разбирать строку
		// на составные части, чтобы избежать неожиданных результатов
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// если gzip не поддерживается, передаём управление
			// дальше без изменений
			h.ServeHTTP(rw, r)
			return
		}

		// создаём gzip.Writer поверх текущего w
		gz, err := gzip.NewWriterLevel(rw, gzip.BestSpeed)
		if err != nil {
			io.WriteString(rw, err.Error())
			return
		}
		defer gz.Close()

		rw.Header().Set("Content-Encoding", "gzip")
		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		h.ServeHTTP(gzipWriter{ResponseWriter: rw, Writer: gz}, r)
	})
}

func main() {

	// Где-то тут надо вызвать пакетову фнукцию и получить параметры.
	Parameters := config.GetParams()
	shortURLDomain = Parameters.ShortBaseURL

	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()

	sugar = *logger.Sugar() // делаем регистратор SugaredLogger
	sugar.Infow(
		"Starting server",
		"addr", Parameters.ServerAddress,
	)

	r := chi.NewRouter()
	r.Use(logHTTPInfo) // Встраиваем логгер в роутер
	r.Use(compressExchange)
	r.Get("/{id}", shortingGetURL)
	r.Post("/", shortingRequest)
	r.Post("/api/shorten", shortingJSON)
	http.ListenAndServe(Parameters.ServerAddress, r)
	//	sugar.Infow(http.ListenAndServe(serverName+":"+serverPort, r).Error().)
}
