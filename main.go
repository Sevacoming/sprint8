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

 store := NewParcelStore(db)
 svc := NewParcelService(store)

 // демо: регистрация и вывод номера
 n, err := svc.RegisterParcel(1, "г. Псков, ул. Пушкина, д. 5")
 if err != nil {
  fmt.Println(err)
  return
 }
 fmt.Printf("Новая посылка #%d зарегистрирована\n", n)
}
