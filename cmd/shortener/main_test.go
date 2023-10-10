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
			name:    "hash которого нет",
			request: shortURLDomain + "/GGGGGGGGGG",
			method:  http.MethodGet,
			want: want{
				response:    "",
				contentType: "text/plain",
				status:      http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.request, nil)
			w := httptest.NewRecorder()
			if tt.name[0] == 'H' { // H - метка для тестов, которые для обычных запросов с ожидаемым нормальным поведением.
				shortingRequest(w, request)
				res := w.Result()
				data, err := io.ReadAll(res.Body)
				res.Body.Close()
				require.NoError(t, err)
				assert.Equal(t, tt.want.response, string(data)[:len(shortURLDomain)])
			} else {
				shortingGetURL(w, request)
				res := w.Result()
				data, err := io.ReadAll(res.Body)
				res.Body.Close()
				require.NoError(t, err)
				assert.Equal(t, tt.want.response, string(data))
				assert.Equal(t, tt.want.status, res.StatusCode)
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
