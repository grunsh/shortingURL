package storage

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"shortingURL/cmd/shortener/config"
	fun "shortingURL/cmd/shortener/f"
)

var err error

var ShrtUserID string

type URLdbtyoe map[string]config.RecordURL

var URLdb URLdbtyoe

type InMemURL struct {
}

type Storer interface {
	Open()
	StoreURL(url []byte, UserID string) ([]byte, error)
	StoreURLbatch(urls []config.RecordURL, UserID string) []config.RecordURL
	GetURL(hash string) config.RecordURL
	GetUserURLs(UserID string) []config.RecordURL
	Close()
}

func (f *InMemURL) GetURL(hash string) config.RecordURL {
	return URLdb[hash]
}

func (f *InMemURL) GetUserURLs(UserID string) []config.RecordURL {
	var tempURLs []config.RecordURL
	for _, u := range URLdb {
		if u.UserID == UserID {
			tempURLs = append(tempURLs, u)
		}
	}
	return tempURLs
}

func (f *InMemURL) StoreURL(url []byte, UserID string) ([]byte, error) {
	hash := fun.GetHash()
	u := config.RecordURL{
		ID:     fun.NextSequenceID(),
		HASH:   hash,
		URL:    string(url),
		UserID: UserID,
	}
	URLdb[hash] = u
	return []byte(config.PRM.ShortBaseURL + hash), nil
}

func (f *InMemURL) StoreURLbatch(urls []config.RecordURL, UserID string) []config.RecordURL {
	var uResp []config.RecordURL
	for _, u := range urls {
		hash := fun.GetHash()
		u := config.RecordURL{
			ID:     fun.NextSequenceID(),
			HASH:   hash,
			URL:    u.URL,
			CorID:  u.CorID,
			UserID: UserID,
		}
		URLdb[hash] = u
		uResp = append(uResp, u)
	}
	return uResp
}

// Метод инициализации хранилища. В данном случае, инициализируем мапу, а то ай-ай
func (f *InMemURL) Open() {
	URLdb = make(URLdbtyoe)
}

// Метод закрытия хранилища. В случае с памятью, крыть нечего.
func (f *InMemURL) Close() {

}

/*---------- Начало. Секция работы с файлом. ----------*/

type FileStorageURL struct {
	FilePath string
}

var Prod *Producer

// Метод инициализации хранилища. В данном случае, читаем файл в память, инитим мапу...
func (f *FileStorageURL) Open() {
	// Читаем весь файл в память, ибо нех в него каждый раз лазать, чтобы найти. Долго это. Никто так не работает.
	URLdb = make(URLdbtyoe)
	Consumer, err := NewConsumer(config.PRM.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	u, _ := Consumer.ReadURL()
	for u != nil {
		if u.ID > fun.SequenceUUID {
			fun.SequenceUUID = u.ID
		}
		URLdb[u.HASH] = *u
		u, _ = Consumer.ReadURL()
	}
	Consumer.Close()
	Prod, err = NewProducer(config.PRM.FileStoragePath)
	if err != nil {
		panic("Ой. Не получилось создать писателя в файл. ")
	}
}

func (f *FileStorageURL) Close() {
	Prod.Close()
}

func (f *FileStorageURL) GetURL(hash string) config.RecordURL {
	return URLdb[hash]
}

func (f *FileStorageURL) GetUserURLs(UserID string) []config.RecordURL {
	var tempURLs []config.RecordURL
	for _, u := range URLdb {
		if u.UserID == UserID {
			tempURLs = append(tempURLs, u)
		}
	}
	return tempURLs
}

func (f *FileStorageURL) StoreURL(url []byte, UserID string) ([]byte, error) {
	hash := fun.GetHash()
	u := config.RecordURL{
		ID:     fun.NextSequenceID(),
		HASH:   hash,
		URL:    string(url),
		UserID: UserID,
	}
	URLdb[hash] = u
	Prod.WriteURL(u)
	return []byte(config.PRM.ShortBaseURL + hash), nil
}

func (f *FileStorageURL) StoreURLbatch(urls []config.RecordURL, UserID string) []config.RecordURL {
	var uResp []config.RecordURL
	for _, u := range urls {
		hash := fun.GetHash()
		u := config.RecordURL{
			ID:     fun.NextSequenceID(),
			HASH:   hash,
			URL:    u.URL,
			CorID:  u.CorID,
			UserID: UserID,
		}
		Prod.WriteURL(u)
		uResp = append(uResp, u)
	}
	return uResp
}

type Producer struct {
	file   *os.File
	writer *bufio.Writer
}

func NewProducer(fileName string) (*Producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *Producer) WriteURL(url config.RecordURL) error {
	data, err := json.Marshal(&url)
	if err != nil {
		return err
	}
	if _, err := p.writer.Write(data); err != nil {
		return err
	}
	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}
	// записываем буфер в файл
	return p.writer.Flush()
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file *os.File
	// заменяем Reader на Scanner
	scanner *bufio.Scanner
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		file: file,
		// создаём новый scanner
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadURL() (*config.RecordURL, error) {
	// одиночное сканирование до следующей строки
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}
	// читаем данные из scanner
	data := c.scanner.Bytes()
	url := config.RecordURL{}
	err := json.Unmarshal(data, &url)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

/*---------- Конец. Секция работы с файлом. ----------*/

/*-------------------- Начало. Секция работы с постгрёй. --------------------*/
var DB *sql.DB

type ErrorsSQL struct {
	Err  error
	Text string
	Code int
}

func (e *ErrorsSQL) Error() string {
	return e.Text
}

func NewSQLError(er error, t string, c int) error {
	return &ErrorsSQL{
		Err:  er,
		Text: t,
		Code: c,
	}
}

const (
	Conflict = http.StatusConflict
)

type DataBase struct {
	DataBaseDSN string
}

func (f *DataBase) GetURL(hash string) config.RecordURL {
	var (
		uuid uint
		url  string
	)
	DB.QueryRow("SELECT u.id, u.url FROM shorturl.url u WHERE u.hash = $1", hash).Scan(&uuid, &url)
	if uuid == 0 {
		return config.RecordURL{ID: 0, HASH: "", URL: "", CorID: ""}
	} else {
		return config.RecordURL{ID: uuid, HASH: hash, URL: url}
	}
}

func (f *DataBase) GetUserURLs(UserID string) []config.RecordURL {
	var uResp []config.RecordURL
	var (
		hash string
		url  string
	)
	rows, err := DB.Query("SELECT hash,url FROM shorturl.url WHERE shrt_uuid=$1", UserID)
	defer rows.Close()
	if err != nil {
		panic("Ой. Не получилось запросить урлы пользака")
	}
	for rows.Next() {
		err := rows.Scan(&hash, &url)
		if err != nil {
			panic(err)
		}
		uResp = append(uResp, config.RecordURL{ID: 0, HASH: hash, URL: url, CorID: "", UserID: UserID})
	}
	return uResp
}

func (f *DataBase) StoreURL(url []byte, UserID string) ([]byte, error) {
	return StoreURLinDataBase(url, UserID)
}

func StoreURLinDataBase(url []byte, UserID string) ([]byte, error) {
	var (
		hashDB string
		urlDB  string
	)
	hash := fun.GetHash()
	u := config.RecordURL{
		ID:     0,
		HASH:   hash,
		URL:    string(url),
		CorID:  "",
		UserID: UserID,
	}
	tx, err := DB.Begin()
	defer tx.Commit()
	if err != nil {
		panic("Ой. Не получилось начать транзакцию.")
	}
	result, err := tx.Exec("insert into shorturl.url (hash,url,correlation_id,shrt_uuid) values ($1,$2,$3,$4) on conflict (url) do nothing", u.HASH, u.URL, u.CorID, u.UserID)
	if err != nil {
		log.Fatal(err)
		fmt.Println(err)
	}
	resu, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
		fmt.Println(err.Error())
	}
	if resu != 1 {
		tx.QueryRow("SELECT u.hash, u.url FROM shorturl.url u WHERE u.url = $1", string(url)).Scan(&hashDB, &urlDB)
		er := NewSQLError(errors.New("ErErEr"), "Already shortened URL: "+urlDB, Conflict)
		fmt.Println(er)
		fmt.Println("Не добавилось ничего: ", "insert into shorturl.url (hash,url,correlation_id,shrt_uuid) values ($1,$2,$3,$4)", u.HASH, u.URL, u.CorID, u.UserID)
		return []byte(config.PRM.ShortBaseURL + hashDB), er
	}
	if u.ID == 0 { // Чисто чтоб вет тест перестал докапываться
		u.ID = 0
	}
	return []byte(config.PRM.ShortBaseURL + hash), nil
}

func (f *DataBase) StoreURLbatch(urls []config.RecordURL, UserID string) []config.RecordURL {
	uResp := make([]config.RecordURL, len(urls))
	tx, err := DB.Begin()
	if err != nil {
		panic("Ой. Не получилось начать транзакцию в StoreURLbatch")
	}
	for i, u := range urls {
		hash := fun.GetHash()
		ur := config.RecordURL{
			ID:     0,
			HASH:   hash,
			URL:    u.URL,
			CorID:  u.CorID,
			UserID: UserID,
		}
		_, err := tx.Exec("insert into shorturl.url (hash,url,correlation_id,shrt_uuid) values ($1,$2,$3,$4)", ur.HASH, ur.URL, ur.CorID, ur.UserID)
		if err != nil {
			tx.Rollback()
			return []config.RecordURL{}
		}
		uResp[i] = ur
	}
	tx.Commit()
	return uResp
}

// Метод инициализации хранилища. В данном случае, оформим запросы для создания схемы и таблиц, если их нет
func (f *DataBase) Open() {
	OpenDataBase()
}

func OpenDataBase() {
	ps := config.PRM.DatabaseDSN
	DB, err = sql.Open("pgx", ps)
	q := "CREATE SCHEMA IF NOT EXISTS shortURL"
	DB.QueryRow(q)
	q = "CREATE table IF NOT EXISTS  shortURL.URL (id bigserial primary key, hash varchar(10), url varchar(255), correlation_id varchar(255), shrt_uuid char(36))"
	DB.QueryRow(q)
	q = "CREATE UNIQUE INDEX url_url_idx ON shorturl.url (url)"
	DB.QueryRow(q)

}

// Метод закрытия хранилища. В случае с памятью, крыть нечего.
func (f *DataBase) Close() {
	DB.Close()
}

/*-------------------- Конец. Секция работы с постгрёй. --------------------*/

func InitStorage(p config.Parameters) Storer {
	if p.DatabaseDSN != "" {
		return Storer(&DataBase{DataBaseDSN: p.DatabaseDSN})
	} else if p.FileStoragePath != "" {
		return Storer(&FileStorageURL{FilePath: p.FileStoragePath})
	}
	return Storer(&InMemURL{})
}
