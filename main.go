package main

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)        // объект хранилища
	service := NewParcelService(store) // сервис

	// маленькая демонстрация: зарегистрируем посылку
	if _, err := service.Add(1, "г. Псков, ул. Пушкина, д. 5"); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("OK")
}
