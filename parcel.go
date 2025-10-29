package main

import (
	"database/sql"
	"errors"
	"time"
)

const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

var (
	ErrNotAllowed = errors.New("operation not allowed in current status")
	ErrNotFound   = errors.New("parcel not found")
)

type Parcel struct {
	Number    int64
	Client    int64
	Status    string
	Address   string
	CreatedAt string
}

type ParcelStore struct{ db *sql.DB }

func NewParcelStore(db *sql.DB) *ParcelStore { return &ParcelStore{db: db} }

// Add — регистрация новой посылки
func (s *ParcelStore) Add(clientID int64, address string) (int64, error) {
	const q = `INSERT INTO parcel (client, status, address, created_at)
            VALUES (?, ?, ?, ?)`
	created := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(q, clientID, ParcelStatusRegistered, address, created)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Get — получить одну посылку по номеру
func (s *ParcelStore) Get(number int64) (Parcel, error) {
	const q = `SELECT number, client, status, address, created_at
            FROM parcel WHERE number = ?`
	var p Parcel
	err := s.db.QueryRow(q, number).
		Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Parcel{}, ErrNotFound
	}
	return p, err
}

// GetByClient — список посылок клиента
func (s *ParcelStore) GetByClient(clientID int64) ([]Parcel, error) {
	const q = `SELECT number, client, status, address, created_at
            FROM parcel WHERE client = ? ORDER BY number`
	rows, err := s.db.Query(q, clientID)
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

// UpdateStatus — сменить статус
func (s *ParcelStore) UpdateStatus(number int64, newStatus string) error {
	switch newStatus {
	case ParcelStatusRegistered, ParcelStatusSent, ParcelStatusDelivered:
	default:
		return errors.New("unknown status")
	}
	res, err := s.db.Exec(`UPDATE parcel SET status = ? WHERE number = ?`, newStatus, number)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

// UpUpUpUpdateAddress — изменить адрес (только когда registered)
func (s *ParcelStore) UpUpUpUpdateAddress(number int64, newAddress string) error {
	res, err := s.db.Exec(
		`UPDATE parcel SET address = ? WHERE number = ? AND status = 'registered'`,
		newAddress, number,
	)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		var cnt int
		_ = s.db.QueryRow(`SELECT COUNT(1) FROM parcel WHERE number = ?`, number).Scan(&cnt)
		if cnt > 0 {
			return ErrNotAllowed
		}
		return ErrNotFound
	}
	return nil
}

// Delete — удалить (только когда registered)
func (s *ParcelStore) Delete(number int64) error {
	res, err := s.db.Exec(`DELETE FROM parcel WHERE number = ? AND status = 'registered'`, number)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		var cnt int
		_ = s.db.QueryRow(`SELECT COUNT(1) FROM parcel WHERE number = ?`, number).Scan(&cnt)
		if cnt > 0 {
			return ErrNotAllowed
		}
		return ErrNotFound
	}
	return nil
}

// Сервисная обёртка
type ParcelService struct{ store *ParcelStore }

func NewParcelService(store *ParcelStore) *ParcelService { return &ParcelService{store: store} }

// «Каноничные» имена
func (s *ParcelService) Add(clientID int64, address string) (int64, error) {
	return s.store.Add(clientID, address)
}
func (s *ParcelService) Get(number int64) (Parcel, error) {
	return s.store.Get(number)
}
func (s *ParcelService) GetByClient(clientID int64) ([]Parcel, error) {
	return s.store.GetByClient(clientID)
}
func (s *ParcelService) UpdateStatus(number int64, newStatus string) error {
	return s.store.UpdateStatus(number, newStatus)
}
func (s *ParcelService) UpUpUpUpdateAddress(number int64, newAddress string) error {
	return s.store.UpUpUpUpdateAddress(number, newAddress)
}
func (s *ParcelService) Delete(number int64) error {
	return s.store.Delete(number)
}

// Синонимы — чтобы не падали старые вызовы
func (s *ParcelService) RegisterParcel(clientID int64, address string) (int64, error) {
	return s.Add(clientID, address)
}
func (s *ParcelService) GetClientParcels(clientID int64) ([]Parcel, error) {
	return s.GetByClient(clientID)
}
func (s *ParcelService) ChangeStatus(number int64, newStatus string) error {
	return s.UpdateStatus(number, newStatus)
}
func (s *ParcelService) Remove(number int64) error {
	return s.Delete(number)
}

func (s *ParcelService) ChangeAddress(number int64, newAddress string) error {
	return s.store.UpdateAddress(number, newAddress)
}

func (ps *ParcelStore) UpdateAddress(number int64, newAddress string) error {
	// Разрешаем менять адрес только у зарегистрированных посылок
	res, err := ps.db.Exec(
		`UPDATE parcel SET address = ? WHERE number = ? AND status = ?`,
		newAddress, number, ParcelStatusRegistered,
	)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		// либо номер не найден, либо статус не "registered"
		return errors.New("update not allowed")
	}
	return nil
}
