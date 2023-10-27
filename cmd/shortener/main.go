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
	"bufio"
	_ "bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net/http"
	"os"
	"shortingURL/cmd/shortener/config"
	"strings"
	"time"
)

// Тип для переменных окружения
type Sconfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

type urlType struct {
	uuid uint
	url  []byte
}

type urlDBtype map[string]*urlType

var (
	cfg            Sconfig               // Переменная для объекта конфигурирования
	shortURLDomain string                // Переменная используется в коде в разных местах, значение присваивается в начале работы их cfg
	urlDB          = make(urlDBtype)     // мапа для урлов, ключ - хеш
	sugar          zap.SugaredLogger     // регистратор журналов
	SeaqunceId     uint              = 0 // Используем для генератора uuid URL
	urlFile        *URLsDB
)

const hashLen int = 10 // Длина генерируемого хеша

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" // Для генератора хэшей

// Для файло с урлами
type URLsDB struct {
	file    *os.File
	encoder *json.Encoder
	decoder *json.Decoder
	sc      *bufio.Scanner
}

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

	//  Разжиматор сжатого
	gzipReader struct {
		r  io.ReadCloser
		zr *gzip.Reader
	}

	// Структура хранения данных URL в файле
	URL struct {
		uuid     uint   `json:"uuid"`
		shortURL string `json:"short_url"`
		URL      string `json:"original_url"`
	}
)

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func newCompressReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &gzipReader{
		r:  r,
		zr: zr,
	}, nil
}

func nextId() uint { // Генерилка сиквенса
	SeaqunceId += 1
	return SeaqunceId
}

func (c gzipReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
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
	urlDB[hash].url = url
	urlDB[hash].uuid = nextId()
	//urlFile.WriteURL(URL{urlDB[hash].uuid, hash, string(url)})
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
	id := req.URL.Path[1:] // Откусываем / и записываем id
	//fmt.Println("shortingGetURL")
	res.Header().Del("Content-Encoding")
	res.Header().Set("Content-Type", "text/plain") // Установим тип ответа text/plain
	if val, ok := urlDB[id]; ok {
		res.Header().Set("Location", string(val.url)) // Укажем куда редирект
		res.WriteHeader(http.StatusTemporaryRedirect) // Передаём 307
	} else {
		res.WriteHeader(http.StatusBadRequest) // Прошли весь массив, но хеша нет.
	}
}

// Хендлер для сокращения URL. На входе принимается URL как text/plain
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
	res.WriteHeader(http.StatusCreated)
	fmt.Println("shotringRequest: ", string(shrtURL))
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
	//	res.Header().Set("Content-Length", strconv.Itoa(len(resp)))
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

// миддлварь-сжиматор тельца ответа и разжиматор тельца запросов
func compressExchange(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") { // если вдруг нам передали сжатое, разжимаем
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr // меняем тело запроса на новое
			defer cr.Close()
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") { // если нам не передали свою готовность к принятию сжатого, просто выходим.
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed) // создаём gzip.Writer поверх текущего w
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()
		w.Header().Set("Content-Encoding", "gzip")
		// передаём обработчику страницы переменную типа gzipWriter для вывода данных
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func URLsDBfile(fileName string) (*URLsDB, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	fmt.Println("debug: ", file)
	return &URLsDB{
		file:    file,
		encoder: json.NewEncoder(file),
		decoder: json.NewDecoder(file),
		sc:      bufio.NewScanner(file),
	}, nil
}

func (u *URLsDB) WriteURL(url URL) error {
	return u.encoder.Encode(&url)
}

func (u *URLsDB) ReadURL() ([]byte, error) {
	// одиночное сканирование до следующей строки
	if !u.sc.Scan() {
		return nil, u.sc.Err()
	}
	// читаем данные из scanner
	url := u.sc.Bytes()
	return url, nil
}
func (u *URLsDB) Close() error {
	// закрываем файл
	return u.file.Close()
}

func main() {

	Parameters := config.GetParams()
	shortURLDomain = Parameters.ShortBaseURL
	storageFileName := Parameters.FileStoragePath

	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()
	sugar = *logger.Sugar() // делаем регистратор SugaredLogger

	// Пытаемся читать урлы, если они вообще есть
	urlFile, _ = URLsDBfile(storageFileName)
	fmt.Println(storageFileName)
	//	sc := bufio.NewScanner(urlFile)
	fmt.Println(urlFile)
	//for sc.Scan() {
	//	fmt.Println(sc.Text())
	//}
	if err != nil {
		sugar.Infow(
			"Очень всё плохо с тем, чтобы прочитать урлы: ",
			"путь: ", Parameters.FileStoragePath,
			"косяк: ", err,
		)
	}
	urlFile.Close()

	sugar.Infow(
		"Starting server",
		"addr", Parameters.ServerAddress,
	)

	r := chi.NewRouter()
	r.Use(compressExchange) // Встраиваем сжиматор-разжиматор
	r.Use(logHTTPInfo)      // Встраиваем логгер в роутер
	r.Get("/{id}", shortingGetURL)
	r.Post("/", shortingRequest)
	r.Post("/api/shorten", shortingJSON)
	http.ListenAndServe(Parameters.ServerAddress, r)
	//sugar.Infow(http.ListenAndServe(serverName+":"+serverPort, r).Error().)
}
