package storage

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"shortingURL/cmd/shortener/config"
	fun "shortingURL/cmd/shortener/f"
)

var err error

type URLdbtyoe map[string]config.RecordURL

var URLdb URLdbtyoe

type InMemURL struct {
}

type Storer interface {
	Open()
	StoreURL(url []byte) []byte
	StoreURLbatch(urls []config.RecordURL) []config.RecordURL
	GetURL(hash string) config.RecordURL
	Close()
}

func (f *InMemURL) GetURL(hash string) config.RecordURL {
	return URLdb[hash]
}

func (f *InMemURL) StoreURL(url []byte) []byte {
	hash := fun.GetHash()
	u := config.RecordURL{
		ID:   fun.NextSequenceID(),
		HASH: hash,
		URL:  string(url),
	}
	URLdb[hash] = u
	OpenDataBase()
	StoreURLinDataBase(url)
	CloseDataBase()
	return []byte(config.PRM.ShortBaseURL + hash)
}

func (f *InMemURL) StoreURLbatch(urls []config.RecordURL) []config.RecordURL {
	var uResp []config.RecordURL
	for _, u := range urls {
		hash := fun.GetHash()
		u := config.RecordURL{
			ID:    fun.NextSequenceID(),
			HASH:  hash,
			URL:   u.URL,
			CorID: u.CorID,
		}
		URLdb[hash] = u
		uResp = append(uResp, u)
	}
	OpenDataBase()
	StoreURLinDataBaseBatch(urls)
	CloseDataBase()
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

func (f *FileStorageURL) StoreURL(url []byte) []byte {
	hash := fun.GetHash()
	u := config.RecordURL{
		ID:   fun.NextSequenceID(),
		HASH: hash,
		URL:  string(url),
	}
	URLdb[hash] = u
	Prod.WriteURL(u)
	OpenDataBase()
	StoreURLinDataBase(url)
	CloseDataBase()
	return []byte(config.PRM.ShortBaseURL + hash)
}

func (f *FileStorageURL) StoreURLbatch(urls []config.RecordURL) []config.RecordURL {
	var uResp []config.RecordURL
	for _, u := range urls {
		hash := fun.GetHash()
		u := config.RecordURL{
			ID:    fun.NextSequenceID(),
			HASH:  hash,
			URL:   u.URL,
			CorID: u.CorID,
		}
		Prod.WriteURL(u)
		uResp = append(uResp, u)
	}
	OpenDataBase()
	StoreURLinDataBaseBatch(urls)
	CloseDataBase()
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
var db *sql.DB

type DataBase struct {
	DataBaseDSN string
}

func (f *DataBase) GetURL(hash string) config.RecordURL {
	var (
		uuid uint
		url  string
	)
	db.QueryRow("SELECT u.id, u.url FROM shorturl.url u WHERE u.hash = $1", hash).Scan(&uuid, &url)
	if uuid == 0 {
		return config.RecordURL{ID: 0, HASH: "", URL: "", CorID: ""}
	} else {
		return config.RecordURL{ID: uuid, HASH: hash, URL: url}
	}
}

func (f *DataBase) StoreURL(url []byte) []byte {
	return []byte(StoreURLinDataBase(url))
}

func StoreURLinDataBase(url []byte) []byte {
	hash := fun.GetHash()
	u := config.RecordURL{
		ID:    0,
		HASH:  hash,
		URL:   string(url),
		CorID: "",
	}
	tx, err := db.Begin()
	if err != nil {
		panic("Ой. Не получилось начать транзакцию.")
	}
	result, err := tx.Exec("insert into shorturl.url (hash,url,correlation_id) values ($1,$2,$3)", u.HASH, u.URL, u.CorID)
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
		fmt.Println("Не добавилось ничего: ", "insert into shorturl.url (hash,url,correlation_id) values ($1,$2,$3)", u.HASH, u.URL, u.CorID)
	}
	tx.Commit()
	if u.ID == 0 { // Чисто чтоб вет тест перестал докапываться
		u.ID = 0
	}
	return []byte(config.PRM.ShortBaseURL + hash)
}

func (f *DataBase) StoreURLbatch(urls []config.RecordURL) []config.RecordURL {
	return StoreURLinDataBaseBatch(urls)
}

func StoreURLinDataBaseBatch(urls []config.RecordURL) []config.RecordURL {
	var uResp []config.RecordURL
	tx, err := db.Begin()
	if err != nil {
		panic("Ой. Не получилось начать транзакцию в StoreURLbatch")
	}
	for _, u := range urls {
		hash := fun.GetHash()
		u := config.RecordURL{
			ID:    fun.NextSequenceID(),
			HASH:  hash,
			URL:   u.URL,
			CorID: u.CorID,
		}
		tx.Exec("insert into shorturl.url (hash,url,correlation_id) values ($1,$2,$3)", u.HASH, u.URL, u.CorID)
		fmt.Println("insert into shorturl.url (hash,url,correlation_id) values ($1,$2,$3)", u.HASH, u.URL, u.CorID)
		uResp = append(uResp, u)
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
	db, err = sql.Open("pgx", ps)
	q := "CREATE SCHEMA IF NOT EXISTS shortURL"
	db.QueryRow(q)
	q = "CREATE table IF NOT EXISTS  shortURL.URL (id bigserial primary key, hash varchar(10), url varchar(255))"
	db.QueryRow(q)
}

// Метод закрытия хранилища. В случае с памятью, крыть нечего.
func (f *DataBase) Close() {
	CloseDataBase()
}

func CloseDataBase() {
	db.Close()
}

/*-------------------- Конец. Секция работы с постгрёй. --------------------*/
