package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // драйвер SQLite
)

func openDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	db, err := openDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	service := NewParcelService(store) // сервис

	// маленькая демонстрация: зарегистрируем посылку
	num, err := service.Register(1, "г. Псков, ул. Пушкина, д. 5")
	if err != nil {
		fmt.Println(err)
		return
	}

	// сменим адрес в статусе registered (разрешено)
	if err := service.ChangeAddress(num, "обновлённый адрес"); err != nil {
		fmt.Println(err)
		return
	}

	// переведём в sent
	if err := service.ChangeStatus(num, ParcelStatusSent); err != nil {
		fmt.Println(err)
		return
	}

	// попробуем удалить в sent (должна прийти ошибка)
	if err := service.Remove(num); err == nil {
		fmt.Println("expected error: delete not allowed when sent")
	}

	fmt.Println("OK")
}
