package db

import (
	"database/sql"
	"fmt"
	"log"

	 _ "github.com/mattn/go-sqlite3"
)

func DB() *sql.DB {
	    // Открываем базу данных (если файла нет, он будет создан)
		db, err := sql.Open("sqlite3", "./urlShortener.db")
		if err != nil {
			fmt.Println(err)
		}

		// Проверяем подключение
		err = db.Ping()
		if err != nil {
			fmt.Println(err)
		}
	
		log.Println("Подключение к базе данных успешно установлено!")
		return db
}