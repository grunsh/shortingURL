/*
Дяденька, вот ты, который взялся смотреть мой код и ревью писать. Обращаюсь к тебе с просьбой, не будь сильно строг,
ПОЖАЛУЙСТА!  Я не кодил почти 20 лет (было php/Perl во времена сисадминства, интернет тогда был dial-up ещё).
Я попросил на работе купить мне курс Go. И как-то так получилось, что был куплен продвинутый курс, а не для тех, кто
учится ходить. Для меня любой мало-мальски работающий код - победа, как для годовалого ребёнка первый дейсяток шагов
держась за палец. Я не сдаюсь. Стараюсь. Я даже с гитом до курса не работал и никогда не писал юнит тестов. А тут вон чо.
Спасибо тебе дяденька за понимание, заранее. Не заваливай пожалуйста. Я почти не сплю, но тяну лямку и грызу гранит.
*/

package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_shortingRequest(t *testing.T) {
	type want struct {
		response    string
		contentType string
		status      int
	}

	tests := []struct {
		name    string
		normal  bool
		api     string
		request string
		method  string
		want    want
	}{
		{
			name:    "Простая проверка простого URL. Добавление",
			normal:  true,
			api:     "/",
			request: "https://www.yandex.ru/",
			method:  http.MethodPost,
			want: want{
				response:    shortURLDomain,
				contentType: "text/plain",
				status:      http.StatusCreated,
			},
		},
		{
			name:    "hash которого нет",
			normal:  false,
			api:     "/{id}",
			request: shortURLDomain + "/GGGGGGGGGG",
			method:  http.MethodGet,
			want: want{
				response:    "",
				contentType: "text/plain",
				status:      http.StatusBadRequest,
			},
		},
		{
			name:    "hash которого нет",
			normal:  false,
			api:     "/{id}",
			request: shortURLDomain + "/GGGG",
			method:  http.MethodGet,
			want: want{
				response:    "",
				contentType: "text/plain",
				status:      http.StatusBadRequest,
			},
		},
		{
			name:    "hash которого нет",
			normal:  false,
			api:     "/{id}",
			request: shortURLDomain + "/",
			method:  http.MethodGet,
			want: want{
				response:    "",
				contentType: "text/plain",
				status:      http.StatusBadRequest,
			},
		},
		{
			name:    "Тест 01 для /api/shorten",
			normal:  false,
			api:     "/api/shorten",
			request: `{"url": "https://practicum.yandex.ru"}`,
			method:  http.MethodPost,
			want: want{
				response:    "",
				contentType: "application/json",
				status:      http.StatusCreated,
			},
		},
	}
	fileStorage = "short-url-db.json"
	parameters.DatabaseDSN = "host=localhost user=shortener password=shortener dbname=shortener sslmode=disable"
	storeURL, getURL = NewStorageDrivers()

	//t.Setenv("FILE_STORAGE_PATH", "short-url-db.json")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.api {
			case "/api/shorten":
				request := httptest.NewRequest(tt.method, shortURLDomain+tt.api, bytes.NewReader([]byte(tt.request)))
				w := httptest.NewRecorder()
				shortingJSON(w, request)
				res := w.Result()
				//				fmt.Println(tt.name, ": ", tt.request, tt.want, "  --  ", tt, request)
				res.Body.Close()
				assert.Equal(t, tt.want.status, res.StatusCode)
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			case "/{id}":
				request := httptest.NewRequest(tt.method, tt.request, nil)
				w := httptest.NewRecorder()
				shortingGetURL(w, request)
				res := w.Result()
				data, err := io.ReadAll(res.Body)
				res.Body.Close()
				require.NoError(t, err)
				assert.Equal(t, tt.want.response, string(data))
				assert.Equal(t, tt.want.status, res.StatusCode)
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			case "/":
				request := httptest.NewRequest(tt.method, tt.request, nil)
				w := httptest.NewRecorder()
				shortingRequest(w, request)
				res := w.Result()
				data, err := io.ReadAll(res.Body)
				res.Body.Close()
				require.NoError(t, err)
				assert.Equal(t, tt.want.response, string(data)[:len(shortURLDomain)])
			}
		})
	}
}

func Test_fileStorage(t *testing.T) {

	fileStorage = "short-url-db-test.json"
	u := recordURL{
		ID:   14,
		HASH: "asdasdasdf",
		URL:  "http://ya.ru",
	}
	t.Run("Попытка записать и прочитать записанное.", func(t *testing.T) {
		p, err := NewProducer(fileStorage)
		require.NoError(t, err)
		p.WriteURL(u)
		p.Close()
		c, err := NewConsumer(fileStorage)
		require.NoError(t, err)
		nu, err := c.ReadURL()
		require.NoError(t, err)
		c.Close()
		assert.Equal(t, u.ID, nu.ID)
		assert.Equal(t, u.HASH, nu.HASH)
		assert.Equal(t, u.URL, nu.URL)
	})
}
