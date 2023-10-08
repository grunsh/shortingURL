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
	"io"
	"math/rand"
	"net/http"
	"strconv"
)

const hashLen int = 10
const shortURLDomain string = "http://localhost:8080/"

var urlStorage [][]string //слайс для хранения URL и их хешей, первый индекс - запись, второй: 0 - URL, 1 - хеш

// Генератор сокращённого URL. Использует константу shortURLDomain как настройку.
func addURL(url []byte) []byte {
	hashStr := getHash() // Сохраним хэш в переменную. Понадобится для сохранения в массиве и для формирования короткого URL
	urlVar := make([]string, 0)
	// Сформируем слайс-строку в слайс урлов. Колонка 0 - сокращённый URL, 1 - хеш
	urlVar = append(urlVar, string(url))
	urlVar = append(urlVar, hashStr)
	urlStorage = append(urlStorage, urlVar) // ...и сохраним её в слайс строк
	return []byte(shortURLDomain + hashStr)
}

// Генератор хеша. Использует константу hashLen для определения длины
func getHash() string {
	var letters = string("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") //словарик для генерации хешей
	var b = ""
	for i := 0; i < hashLen; i++ {
		b += string(letters[rand.Intn(len(letters))])
	}
	return b
}

// Хендлер / для сокращения URL. На входе принимается URL как text/plain
func shortingRequest(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet { // Если GET / , то вернём редирект и сокращённый URL
		id := req.URL.Path[1:]                      // Откусываем / и записываем id
		for i := len(urlStorage) - 1; i >= 0; i-- { //По слайсу идём с конца, ищем самый свежий редирект
			if urlStorage[i][1] == id {
				res.Header().Set("Location", urlStorage[i][0]) // Укажем куда редирект
				res.Header().Set("Content-Type", "text/plain") // Установим тип ответа text/plain
				res.WriteHeader(http.StatusTemporaryRedirect)  // Передаём 307
				return                                         // Нашли нужный хеш и выдали редирект. Завершаем работу хендлера.
			}
		} //for... поиск хеша в памяти
		res.Header().Set("Content-Type", "text/plain") // Установим тип ответа text/plain
		res.WriteHeader(http.StatusBadRequest)         // Прошли весь массив, но хеша нет. Ошибка 400
		return                                         // Выход по 400
	} else if req.Method == http.MethodPost {
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
	} else {
		res.Header().Set("Content-Type", "text/plain") // Установим тип ответа text/plain
		res.WriteHeader(http.StatusBadRequest)         // Ошибка запроса (не пост, и не гет)
	}

}

func main() {
	shorting := http.NewServeMux()
	shorting.HandleFunc(`/`, shortingRequest)

	err := http.ListenAndServe(`:8080`, shorting)
	if err != nil {
		panic(err)
	}
}
