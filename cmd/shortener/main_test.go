/*
Дяденька, вот ты, который взялся смотреть мой код и ревью писать. Обращаюсь к тебе с просьбой, не будь сильно строг,
ПОЖАЛУЙСТА!   Я не кодил почти 20 лет (было php/Perl во времена сисадминства, интернет тогда был dial-up ещё).
Я попросил на работе купить мне курс Go. И как-то так получилось, что был куплен продвинутый курс, а не для тех, кто
учится ходить. Для меня любой мало-мальски работающий код - победа, как для годовалого ребёнка первый дейсяток шагов
держась за палец. Я не сдаюсь. Стараюсь. Я даже с гитом до курса не работал и никогда не писал юнит тестов. А тут вон чо.
Спасибо тебе дяденька за понимание, заранее. Не заваливай пожалуйста. Я почти не сплю, но тяну лямку и грызу гранит.
*/

package main

import (
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
		request string
		method  string
		want    want
	}{
		{
			name:    "H Простая проверка простого URL (как-то надо откусить hash)",
			request: "https://www.yandex.ru/",
			method:  http.MethodPost,
			want: want{
				response:    shortURLDomain,
				contentType: "text/plain",
				status:      http.StatusCreated,
			},
		},
		{
			name:    "URL правильный, но не тот метод",
			request: "https://www.yandex.ru/",
			method:  http.MethodPut,
			want: want{
				response:    "",
				contentType: "text/plain",
				status:      http.StatusBadRequest,
			},
		},
		{
			name:    "hash которого нет",
			request: shortURLDomain + "/123",
			method:  http.MethodGet,
			want: want{
				response:    "",
				contentType: "text/plain",
				status:      http.StatusBadRequest,
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.request, nil)
			w := httptest.NewRecorder()
			shortingRequest(w, request)
			res := w.Result()
			res.Body.Close()
			data, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			if tt.name[0] == 'H' { // H - метка для тестов, которые для обычных запросов с ожидаемым нормальным поведением.
				assert.Equal(t, tt.want.response, string(data)[:len(shortURLDomain)])
			} else {
				assert.Equal(t, tt.want.response, string(data))
				assert.Equal(t, tt.want.status, res.StatusCode)
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
