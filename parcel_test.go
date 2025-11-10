package main

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func getTestParcel() Parcel {
	return Parcel{
		Client:  1,
		Address: "Псков, д. Пушкина, ул. Колотушкина, д. 5",
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", "tracker.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db
}

func TestAddGetDelete(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)
	p := getTestParcel()

	id, err := store.Add(int64(p.Client), p.Address)
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	got, err := store.Get(id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Client != p.Client || got.Address != p.Address {
		t.Fatalf("unexpected parcel: %+v", got)
	}
	if got.Status != ParcelStatusRegistered {
		t.Fatalf("expected status %q, got %q", ParcelStatusRegistered, got.Status)
	}

	if err := store.Delete(id); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestSetAddress(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)

	id, err := store.Add(1, "старый адрес")
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	newAddr := "new test address"
	if err := store.SetAddress(id, newAddr); err != nil {
		t.Fatalf("set address: %v", err)
	}

	got, err := store.Get(id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Address != newAddr {
		t.Fatalf("address not updated: %q", got.Address)
	}
}

func TestSetStatus(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	store := NewParcelStore(db)

	id, err := store.Add(1, "любой адрес")
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	if err := store.SetStatus(id, ParcelStatusSent); err != nil {
		t.Fatalf("set status sent: %v", err)
	}
	got, err := store.Get(id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != ParcelStatusSent {
		t.Fatalf("status not updated to sent: %q", got.Status)
	}

	if err := store.SetStatus(id, ParcelStatusDelivered); err != nil {
		t.Fatalf("set status delivered: %v", err)
	}
	got, err = store.Get(id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != ParcelStatusDelivered {
		t.Fatalf("status not updated to delivered: %q", got.Status)
	}
}
