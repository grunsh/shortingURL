package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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
				response:    shortUrlDomain,
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
			request: shortUrlDomain + "/123",
			method:  http.MethodGet,
			want: want{
				response:    shortUrlDomain,
				contentType: "text/plain",
				status:      http.StatusCreated,
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.request, nil)
			w := httptest.NewRecorder()
			shortingRequest(w, request)
			data, _ := io.ReadAll(w.Body)
			if tt.name[0] == 'H' {
				assert.Equal(t, tt.want.response, string(data)[:len(shortUrlDomain)])
				fmt.Println(string(data)[:len(shortUrlDomain)])
			}
		})
	}
}
