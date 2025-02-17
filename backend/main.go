package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"url-shortener/internal/db"

	_ "github.com/mattn/go-sqlite3"
)

// Какие данные забираем из запроса с фронта.
type RequestData struct {
	BaseUrl string `json:"baseUrl"`
	User    string `json:"user"`
}

type URL struct {
	Original_url string `json:"original_url"`
	Short_url    string `json:"short_url"`
}

// БД
var sqliteDB = db.DB()

func main() {
	// БД
	defer sqliteDB.Close()

	createTable(sqliteDB)

	// Каждые 5 минут запускаем поиск и удаление просроченных записей.
	go func() {
		for {
			deleteOldUrls(sqliteDB)
			time.Sleep(10 * time.Second)
		}
	}()

	mux := http.NewServeMux()
	// Ручки

	// Создаем короткую ссылку
	mux.HandleFunc("POST /url-short", createShortUrl)
	// Переход по короткой ссылке
	mux.HandleFunc("GET /user-urls/{id}/", useShrotUrl)
	// Получить все ссылки юзера
	mux.HandleFunc("GET /user-urls/", getAllUserUrls)
	// Home page
	// mux.HandleFunc("GET /", homePageHandler)
	mux.Handle("/", http.FileServer(http.Dir("./dist")))

	// MiddleWare

	// Если запрос выполняется дольше 2 секунд, завершаем его выполнение и возвращаем 503
	timeoutHandler := http.TimeoutHandler(mux, 2*time.Second, "Превышено время выполнения запроса на сервере, попробуйте позже.")

	// Конфигурация сервера
	server := &http.Server{
		Addr:    ":9462",
		Handler: enableCORS(timeoutHandler),
	}

	log.Printf("Сервер URL SHORTENER %s\n", server.Addr)
	// Запуск сервера
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Ошибка, %s\n", err)
	}
}

// func homePageHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "text/plain")
// 	// resp, _ := json.Marshal("URL SHORTENED BACKEND")
// 	resp := "URL SHORTENED BACKEND"
// 	w.Write([]byte(resp))
// }

func getAllUserUrls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Читаем запрос
	queryParams := r.URL.Query()
	userID := queryParams.Get("user")

	getAllUserUrlsQuery := `SELECT short_url, original_url  FROM urls WHERE user_id = ?`
	rows, _ := sqliteDB.Query(getAllUserUrlsQuery, userID)
	var urls []URL
	for rows.Next() {
		var url URL
		err := rows.Scan(&url.Short_url, &url.Original_url)
		if err != nil {
			log.Fatal(err)
		}
		urls = append(urls, url)
	}
	resp, _ := json.Marshal(urls)
	w.Write(resp)

}

// Обработка перехода по короткому URl
// Ищем короткий URL в базе, если есть делаем редирект,
// если нету просто выкидываем сообщение
func useShrotUrl(w http.ResponseWriter, r *http.Request) {
	var originUrl string
	nf := "Похоже URL устарел, создайте новый! http://127.0.0.1:5173/"
	notFoundUrl, _ := json.Marshal(nf)
	// Тут достаем короткий url из строки запроса
	shortUrl := r.PathValue("id")

	getOriginalUrlQuery := `SELECT original_url from urls WHERE short_url = ?`
	err := sqliteDB.QueryRow(getOriginalUrlQuery, shortUrl).Scan(&originUrl)
	if err != nil {
		log.Println("Не удалось найти оригинальный URL:", err)
		w.Write(notFoundUrl)
	} else {
		// resp, _ := json.Marshal("HELLO")
		originUrl = fmt.Sprintf(`https://%s`, originUrl)
		http.Redirect(w, r, originUrl, http.StatusFound)
	}

}

// Создаем короткий url, заносим его в базу
func createShortUrl(w http.ResponseWriter, req *http.Request) {
	// Читаем запрос
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения тела запроса", http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	// Парсинг JSON
	var data RequestData
	if err := json.Unmarshal(reqBody, &data); err != nil {
		http.Error(w, "Ошибка парсинга JSON", http.StatusBadRequest)
		return
	}
	shortUrl := generateShortURL()

	err = insertURL(sqliteDB, shortUrl, data.BaseUrl, data.User)
	if err != nil {
		log.Println("Не удалось создать коротку ссылку:", err)
	}
	str := shortUrl
	resp, err := json.Marshal(str)
	if err != nil {
		log.Println("Ошибка cериализации", err)
	}
	w.Header().Set("Content-Type", "application/json") // Устанавливаем Content-Type
	w.Write(resp)
}

// Создаем таблицы если они не созданы.
func createTable(sqliteDB *sql.DB) error {
	queryUrls := `
    CREATE TABLE IF NOT EXISTS urls (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
        short_url TEXT NOT NULL UNIQUE,
        original_url TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		clicks INTEGER DEFAULT 0
    );
    `

	queryUsers := `CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id text not null unique,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := sqliteDB.Exec(queryUrls)
	if err != nil {
		log.Println("Ошибка при создании таблицы URLS")
	}
	_, err = sqliteDB.Exec(queryUsers)
	if err != nil {
		log.Println("Ошибка при создании таблицы USERS")
	}

	return err
}

// Генератор коротких URL
func generateShortURL() string {
	b := make([]byte, 6) // Генерируем 6 байт
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:8] // Укорачиваем до 8 символов
}

// Проверяем существует ли юзер в БД, если нет то создаем.
// Создаем запись в БД с коротким URl.
func insertURL(sqliteDB *sql.DB, shortURL string, originalURL string, userID string) error {
	var userExist bool

	checkUserQuery := `SELECT exists(select * FROM users WHERE user_id = ?)`

	err := sqliteDB.QueryRow(checkUserQuery, userID).Scan(&userExist)
	if err != nil {
		log.Println("Ошибка при поиске юзера в таблице USERS:", err)
		return err
	}

	if !userExist {
		createUserQuery := `INSERT INTO users (user_id) values (?)`
		_, err := sqliteDB.Exec(createUserQuery, userID)
		if err != nil {
			log.Println("Ошибка при создании юзера:", err)
			return err
		}
	}

	query := `INSERT INTO urls (user_id, short_url, original_url, expires_at) VALUES (?, ?, ?,  DATETIME('now', '+5 minutes'))`

	_, err = sqliteDB.Exec(query, userID, shortURL, originalURL)
	if err != nil {
		log.Println("Ошибка при создании короткого URL:", err)
		return err
	}
	return nil
}

// Зомби процесс, ищем и удаляем созданные короткие URL которым больше 6 минут.
func deleteOldUrls(sqliteDB *sql.DB) error {
	query := `DELETE FROM urls where expires_at < DATETIME('now', '-1 minutes') OR expires_at IS NULL`
	_, err := sqliteDB.Exec(query)
	if err != nil {
		return fmt.Errorf("ошибка удаления просроченных URL, %w", err)
	}
	return nil
}

// Это CORS.
func enableCORS(next http.Handler) http.Handler {
	allowedOrigins := map[string]bool{
		"http://127.0.0.1:5173": true,
		"http://localhost":      true,
		"http://app.localhost":  true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			return
		}

		// Если запрос разрешен, передаем его дальше.
		next.ServeHTTP(w, r)
	})
}
