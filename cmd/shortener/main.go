package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
)

const hashLen int = 10
const shortUrlDomain string = "http://localhosst:8080/"

var urlStorage [][]string //слайс для хранения URL и их хешей, первый индекс - запись, второй: 0 - URL, 1 - хеш

// Генератор сокращённого URL. Использует константу shortUrlDomain как настройку.
func addURL(url []byte) []byte {
	//	tempVar := string(url)
	hashStr := getHash() // Сохраним хэш в переменную. Понадобится для сохранения в массиве и для формирования короткого URL
	urlVar := make([]string, 0)
	// Сформируем слайс-строку в слайс урлов. Колонка 0 - сокращённый URL, 1 - хеш
	urlVar = append(urlVar, string(url))
	urlVar = append(urlVar, hashStr)
	// ...и сохраним её в слайс строк
	urlStorage = append(urlStorage, urlVar)
	return []byte(shortUrlDomain + hashStr)
}

// Генератор хеша. Использует константу hashLen для определения длины
func getHash() string {
	var letters = string("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") //словарик для генерации хешей
	var b string = ""
	for i := 0; i < hashLen; i++ {
		b += string(letters[rand.Intn(len(letters))])
	}
	return b
}

// Хендлер / для сокращения URL. На входе принимается URL как text/plain
func shortingRequest(res http.ResponseWriter, req *http.Request) {
	// Проверяем на тему метода POST, если что, кричим нервно
	if req.Method == http.MethodGet {
		//		http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
		id := req.URL.Path[1:]                      // Откусываем / и записываем id
		for i := len(urlStorage) - 1; i >= 0; i-- { //По слайсу идём с конца, ищем самый свежий редирект
			if urlStorage[i][1] == id {
				res.Header().Set("Location", urlStorage[i][0]) // Укажем куда редирект
				res.WriteHeader(http.StatusTemporaryRedirect)
				fmt.Println(urlStorage[i][0], urlStorage[i][1])
				return // Нашли нужный хеш и выдали редирект. Выходим.
			}
		}
		res.WriteHeader(http.StatusBadRequest) // Прошли весь массив, но хеша нет. Ошибка
		return
	} else if req.Method == http.MethodPost {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
		}
		shrtUrl := addURL(data)
		fmt.Println(urlStorage) // Для отладки, пишет в консоль весь массив скохранённого
		res.WriteHeader(http.StatusCreated)
		res.Header().Set("Content-Type", "text/plain") // Установим тип ответа text/plain
		res.Header().Set("Content-Length", string(len(shrtUrl)))
		res.Write(shrtUrl)

	} else {
		res.WriteHeader(http.StatusBadRequest) // Ошибка запроса (не пост, и не гет)
	}

}

func returnRedirect(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("Это страница /api."))
}

func main() {
	shorting := http.NewServeMux()
	shorting.HandleFunc(`/`, shortingRequest)

	err := http.ListenAndServe(`:8080`, shorting)
	if err != nil {
		panic(err)
	}
}
