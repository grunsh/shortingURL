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
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"io"
	"net/http"
	"shortingURL/cmd/shortener/config"
	"shortingURL/cmd/shortener/storage"
	"strings"
	"time"
)

// Блок зла. Глобальные переменные и константы
var (
	shortURLDomain string            // Переменная используется в коде в разных местах, значение присваивается в начале работы их cfg
	sugar          zap.SugaredLogger // регистратор журналов
	parameters     config.Parameters //Глобалочка для параметров
	err            error
	db             *sql.DB
	URLstorage     storage.Storer
)

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

	// Тип с интерфейсом для компрессии
	gzipReader struct {
		r  io.ReadCloser
		zr *gzip.Reader
	}
)

/*---------- Начало. Блок про сжиматорство ----------*/
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

func (c gzipReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

/*---------- Конец. Блок про сжиматорство ----------*/

// Функции журналирования
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

/*---------- Начало. Секция хендлеров ----------*/
// Хендлер пинга БД. 500, если не успели поингануться за 2 сек или вообще не дорвались, или умер контекст http-запроса.
func ping(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()
	ps := config.PRM.DatabaseDSN
	db, err = sql.Open("pgx", ps)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = db.PingContext(ctx)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
	db.Close()
}

// Хендлер получения сокращённого URL. 307 и редирект, или ошибка.
func shortingGetURL(res http.ResponseWriter, req *http.Request) {
	id := req.URL.Path[1:] // Откусываем / и записываем id
	req.Body.Close()
	res.Header().Del("Content-Encoding")
	res.Header().Set("Content-Type", "text/plain") // Установим тип ответа text/plain
	u := URLstorage.GetURL(id)
	if u.ID != 0 {
		res.Header().Set("Location", u.URL)           // Укажем куда редирект
		res.WriteHeader(http.StatusTemporaryRedirect) // Передаём 307
	} else {
		res.WriteHeader(http.StatusBadRequest) // Прошли весь массив, но хеша нет.
	}
}

// Хендлер для сокращения URL. На входе принимается URL как text/plain
func shortingRequest(res http.ResponseWriter, req *http.Request) {
	data, err := io.ReadAll(req.Body)
	req.Body.Close()
	res.Header().Set("Content-Type", "text/plain") // Установим тип ответа text/plain
	if err != nil {                                // Если что-то не то с чтением запроса, выходим!
		res.WriteHeader(http.StatusBadRequest)
		return // Выход по 400
	}
	shrtURL, er := URLstorage.StoreURL(data)
	var sqErr *storage.ErrorsSQL
	if errors.As(er, &sqErr) {
		res.WriteHeader(sqErr.Code)
		res.Write(shrtURL)
		return
	}
	res.WriteHeader(http.StatusCreated)
	res.Write(shrtURL)
}

// JSON хендлер для сокращения URL. На входе принимается URL как JSON
func shortingJSON(res http.ResponseWriter, req *http.Request) {
	type URLReq struct { // Тип для запроса с тегом url
		URL string `json:"url"`
	}
	type URLResp struct { // Тип для ответа с тегом result
		Result string `json:"result"`
	}
	var reqURL URLReq
	var respURL URLResp
	res.Header().Set("Content-Type", "application/json") // Установим тип ответа application/json
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body) // Чтение тела запроса в буфер buf
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return // Выход по 400 (ошибка чтения тела запроса)
	}
	if err = json.Unmarshal(buf.Bytes(), &reqURL); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return // Выход по 400
	}
	r, err := URLstorage.StoreURL([]byte(reqURL.URL))
	respURL.Result = string(r) //StoreURL возвращает []byte для хендлера с plain/text, а тут нам строка нужна
	var sqErr *storage.ErrorsSQL
	if errors.As(err, &sqErr) { // Смотрим, ошибка нам наша вернулась? Если да, то выведем сокращённый урл, и 409 ошибку
		res.WriteHeader(sqErr.Code)
		resp, err := json.Marshal(respURL)
		if err != nil {
			res.WriteHeader(sqErr.Code)
			return // Выход по 400
		}
		res.Write(resp)
		return
	}
	resp, err := json.Marshal(respURL)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return // Выход по 400
	}
	res.WriteHeader(http.StatusCreated)
	res.Write(resp)
}

// JSON хендлер для пакетного сокращения URL. На входе принимается URL как JSON
func shortingJSONbatch(res http.ResponseWriter, req *http.Request) {
	type URLResp struct { // Тип для ответа с тегом result
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}
	var respURL []URLResp
	var reqURL []config.RecordURL
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)                     // Чтение тела запроса в буфер buf
	res.Header().Set("Content-Type", "application/json") // Установим тип ответа application/json
	if err != nil {                                      // Если не удалось прочитать запрос, то выходить надо по 400
		res.WriteHeader(http.StatusBadRequest)
		return // Выход по 400 (ошибка чтения тела запроса)
	}
	if err = json.Unmarshal(buf.Bytes(), &reqURL); err != nil { // Если не удалось распарсить JSON, вылетаем по 400
		res.WriteHeader(http.StatusBadRequest)
		return // Выход по 400
	}
	for _, u := range URLstorage.StoreURLbatch(reqURL) { //В цикле формируем массив элементов с тегированной структурой, как того требует задание.
		respURL = append(respURL, URLResp{
			CorrelationID: u.CorID,
			ShortURL:      config.PRM.ShortBaseURL + u.HASH,
		})
	}
	resp, err := json.Marshal(respURL)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return // Выход по 400
	}
	res.WriteHeader(http.StatusCreated)
	res.Write(resp)
}

/*---------- Конец. Секция хендлеров ----------*/

/*---------- Начало. Секция миддлаварей. ----------*/
// миддлварь-обёртка для журналирования запросов
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

/*---------- Конец. Секция миддлаварей. ----------*/

func main() {
	config.PRM = config.GetParams()              // Для начала получаем все параметры
	URLstorage = storage.InitStorage(config.PRM) // В инит входит логика выбора хранилища + создание таблиц в БД, если БД.
	URLstorage.Open()
	defer URLstorage.Close()
	shortURLDomain = parameters.ShortBaseURL

	// Раскручиваем маховик журналирования
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()
	sugar = *logger.Sugar() // делаем регистратор SugaredLogger
	sugar.Infow(
		"Starting server",
		"addr", config.PRM.ServerAddress,
	)

	// Роутер. Регистрируем миддлвари, хендлеры и запускаемся.
	r := chi.NewRouter()
	r.Use(compressExchange) // Встраиваем сжиматор-разжиматор
	r.Use(logHTTPInfo)      // Встраиваем логгер в роутер
	r.Get("/{id}", shortingGetURL)
	r.Post("/", shortingRequest)
	r.Post("/api/shorten", shortingJSON)
	r.Post("/api/shorten/batch", shortingJSONbatch)
	r.Get("/ping", ping)
	err = http.ListenAndServe(config.PRM.ServerAddress, r)
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
}
