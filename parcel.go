package main

import (
	"database/sql"
	"errors"
	"time"
)

/*************** Доменные константы ***************/
const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

/*************** Модель ***************/
type Parcel struct {
	Number    int64
	Client    int64
	Status    string
	Address   string
	CreatedAt string
}

/*************** Хранилище (SQL) ***************/
type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) *ParcelStore {
	return &ParcelStore{db: db}
}

func (s *ParcelStore) Add(clientID int64, address string) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(
		`INSERT INTO parcel (client, status, address, created_at)
         VALUES (:client, :status, :address, :created_at)`,
		sql.Named("client", clientID),
		sql.Named("status", ParcelStatusRegistered),
		sql.Named("address", address),
		sql.Named("created_at", now),
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return id, err
}

func (s *ParcelStore) Get(number int64) (Parcel, error) {
	var p Parcel
	row := s.db.QueryRow(
		`SELECT number, client, status, address, created_at
         FROM parcel WHERE number = :number`,
		sql.Named("number", number),
	)
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	return p, err
}

func (s *ParcelStore) GetByClient(clientID int64) ([]Parcel, error) {
	rows, err := s.db.Query(
		`SELECT number, client, status, address, created_at
         FROM parcel WHERE client = :client ORDER BY number`,
		sql.Named("client", clientID),
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

func (s *ParcelStore) UpdateStatus(number int64, newStatus string) error {
	_, err := s.db.Exec(
		`UPDATE parcel SET status = :status WHERE number = :number`,
		sql.Named("status", newStatus),
		sql.Named("number", number),
	)
	return err
}

func (s *ParcelStore) UpdateAddress(number int64, newAddress string) error {
	_, err := s.db.Exec(
		`UPDATE parcel SET address = :address WHERE number = :number`,
		sql.Named("address", newAddress),
		sql.Named("number", number),
	)
	return err
}

func (s *ParcelStore) Delete(number int64) error {
	_, err := s.db.Exec(
		`DELETE FROM parcel WHERE number = :number`,
		sql.Named("number", number),
	)
	return err
}

/*************** Сервис (бизнес-правила) ***************/
type ParcelService struct {
	store *ParcelStore
}

func NewParcelService(store *ParcelStore) *ParcelService {
	return &ParcelService{store: store}
}

// Регистрация новой посылки
func (s *ParcelService) RegisterParcel(clientID int64, address string) (int64, error) {
	return s.store.Add(clientID, address)
}

// Для совместимости с main.go
func (s *ParcelService) Register(clientID int64, address string) (int64, error) {
	return s.RegisterParcel(clientID, address)
}

// Список посылок клиента
func (s *ParcelService) GetClientParcels(clientID int64) ([]Parcel, error) {
	return s.store.GetByClient(clientID)
}

// Смена статуса с валидацией переходов
func (s *ParcelService) ChangeStatus(number int64, newStatus string) error {
	p, err := s.store.Get(number)
	if err != nil {
		return err
	}
	switch {
	case p.Status == ParcelStatusRegistered && newStatus == ParcelStatusSent:
		return s.store.UpdateStatus(number, newStatus)
	case p.Status == ParcelStatusSent && newStatus == ParcelStatusDelivered:
		return s.store.UpdateStatus(number, newStatus)
	default:
		return errors.New("status change not allowed")
	}
}

// Смена адреса (только в registered)
func (s *ParcelService) ChangeAddress(number int64, newAddress string) error {
	p, err := s.store.Get(number)
	if err != nil {
		return err
	}
	if p.Status != ParcelStatusRegistered {
		return errors.New("update address not allowed when not registered")
	}
	return s.store.UpdateAddress(number, newAddress)
}

// Удаление (только в registered)
func (s *ParcelService) Remove(number int64) error {
	p, err := s.store.Get(number)
	if err != nil {
		return err
	}
	if p.Status != ParcelStatusRegistered {
		return errors.New("delete not allowed when not registered")
	}
	return s.store.Delete(number)
}
