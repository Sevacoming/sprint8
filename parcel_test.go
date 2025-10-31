package main

import (
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

// helper для тестов: использует openDB() из main.go и возвращает только *sql.DB
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := openDB()
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func resetRow(t *testing.T, db *sql.DB, num int64) {
	t.Helper()
	db.Exec(`DELETE FROM parcel WHERE number = ?`, num)
}

func TestRegisterAndList(t *testing.T) {
	db := openTestDB(t)
	store := NewParcelStore(db)
	svc := NewParcelService(store)

	num, err := svc.RegisterParcel(1, "г. Псков, ул. Пушкина, д. 5")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	t.Cleanup(func() { resetRow(t, db, num) })

	list, err := svc.GetClientParcels(1)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	found := false
	for _, p := range list {
		if p.Number == num && p.Status == ParcelStatusRegistered {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("registered parcel not found in client list")
	}
}

func TestStatusFlow(t *testing.T) {
	db := openTestDB(t)
	store := NewParcelStore(db)
	svc := NewParcelService(store)

	num, err := svc.RegisterParcel(2, "г. Саратов, ул. Верхние Зори, д. 25")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	t.Cleanup(func() { resetRow(t, db, num) })

	if err := svc.ChangeStatus(num, ParcelStatusSent); err != nil {
		t.Fatalf("to sent: %v", err)
	}
	if err := svc.ChangeStatus(num, ParcelStatusDelivered); err != nil {
		t.Fatalf("to delivered: %v", err)
	}
}

func TestUpdateAddressRules(t *testing.T) {
	db := openTestDB(t)
	store := NewParcelStore(db)
	svc := NewParcelService(store)

	num, err := svc.RegisterParcel(3, "первый адрес")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	t.Cleanup(func() { resetRow(t, db, num) })

	if err := svc.ChangeAddress(num, "обновлённый адрес"); err != nil {
		t.Fatalf("update address registered: %v", err)
	}

	if err := svc.ChangeStatus(num, ParcelStatusSent); err != nil {
		t.Fatalf("to sent: %v", err)
	}
	if err := svc.ChangeAddress(num, "нельзя уже"); err == nil {
		t.Fatalf("expected error when updating address after sent")
	}
}

func TestDeleteRules(t *testing.T) {
	db := openTestDB(t)
	store := NewParcelStore(db)
	svc := NewParcelService(store)

	num, err := svc.RegisterParcel(4, "адрес")
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if err := svc.ChangeStatus(num, ParcelStatusSent); err != nil {
		t.Fatalf("to sent: %v", err)
	}
	if err := svc.Remove(num); err == nil {
		t.Fatalf("expected error: delete not allowed when sent")
	}
	resetRow(t, db, num)
}

func TestMain(m *testing.M) { os.Exit(m.Run()) }
