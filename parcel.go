package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

/************ Store: работа с БД ************/

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore { return ParcelStore{db: db} }

// Add — создаёт запись со статусом "registered" и текущим временем.
func (s ParcelStore) Add(client int64, address string) (int64, error) {
	createdAt := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(
		`INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)`,
		client, ParcelStatusRegistered, address, createdAt,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s ParcelStore) Get(number int64) (Parcel, error) {
	var p Parcel
	err := s.db.QueryRow(
		`SELECT number, client, status, address, created_at
     FROM parcel
    WHERE number = ?`,
		number,
	).Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	return p, err
}

func (s ParcelStore) ListByClient(client int64) ([]Parcel, error) {
	rows, err := s.db.Query(
		`SELECT number, client, status, address, created_at
     FROM parcel
    WHERE client = ?
    ORDER BY number`,
		client,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Parcel
	for rows.Next() {
		var p Parcel
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s ParcelStore) SetStatus(number int64, status string) error {
	_, err := s.db.Exec(`UPDATE parcel SET status = ? WHERE number = ?`, status, number)
	return err
}

func (s ParcelStore) SetAddress(number int64, address string) error {
	_, err := s.db.Exec(`UPDATE parcel SET address = ? WHERE number = ?`, address, number)
	return err
}

func (s ParcelStore) Delete(number int64) error {
	_, err := s.db.Exec(`DELETE FROM parcel WHERE number = ?`, number)
	return err
}

/************ Service: логика и алиасы под вызовы из main.go ************/

type ParcelService struct {
	store ParcelStore
}

func NewParcelService(store ParcelStore) ParcelService { return ParcelService{store: store} }

func isValidStatus(st string) bool {
	return st == ParcelStatusRegistered || st == ParcelStatusSent || st == ParcelStatusDelivered
}

func (s ParcelService) Register(client int, address string) (Parcel, error) {
	id, err := s.store.Add(int64(client), address)
	if err != nil {
		return Parcel{}, err
	}
	return s.store.Get(id)
}

func (s ParcelService) Get(number int) (Parcel, error) {
	return s.store.Get(int64(number))
}

func (s ParcelService) List(client int) ([]Parcel, error) {
	return s.store.ListByClient(int64(client))
}

func (s ParcelService) SetStatus(number int, status string) error {
	if !isValidStatus(status) {
		return errors.New("invalid status")
	}
	return s.store.SetStatus(int64(number), status)
}

func (s ParcelService) SetAddress(number int, address string) error {
	p, err := s.store.Get(int64(number))
	if err != nil {
		return err
	}
	if p.Status != ParcelStatusRegistered {
		return errors.New("address can be changed only for registered parcels")
	}
	return s.store.SetAddress(int64(number), address)
}

func (s ParcelService) Delete(number int) error {
	p, err := s.store.Get(int64(number))
	if err != nil {
		return err
	}
	if p.Status != ParcelStatusRegistered {
		return errors.New("parcel can be deleted only in 'registered' status")
	}
	return s.store.Delete(int64(number))
}

/* --- алиасы, чтобы main.go мог вызывать ChangeAddress/ve --- */

func asNumber(id interface{}) (int, error) {
	switch v := id.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case Parcel:
		return v.Number, nil
	case *Parcel:
		return v.Number, nil
	default:
		return 0, fmt.Errorf("unsupported id type %T", id)
	}
}

func (s ParcelService) ChangeAddress(num interface{}, address string) error {
	n, err := asNumber(num)
	if err != nil {
		return err
	}
	return s.SetAddress(n, address)
}

func (s ParcelService) ChangeStatus(num interface{}, status string) error {
	n, err := asNumber(num)
	if err != nil {
		return err
	}
	return s.SetStatus(n, status)
}

func (s ParcelService) Remove(num interface{}) error {
	n, err := asNumber(num)
	if err != nil {
		return err
	}
	return s.Delete(n)
}
