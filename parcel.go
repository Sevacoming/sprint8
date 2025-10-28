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

func (s *ParcelStore) Create(clientID int64, address string) (int64, error) {
 const q = `INSERT INTO parcel (client, status, address, created_at)
            VALUES (?, ?, ?, ?)`
 created := time.Now().UTC().Format(time.RFC3339)
 res, err := s.db.Exec(q, clientID, ParcelStatusRegistered, address, created)
 if err != nil {
  return 0, err
 }
 return res.LastInsertId()
}

func (s *ParcelStore) ListByClient(clientID int64) ([]Parcel, error) {
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

func (s *ParcelStore) UpdateStatus(number int64, newStatus string) error {
 switch newStatus {
 case ParcelStatusRegistered, ParcelStatusSent, ParcelStatusDelivered:
 default:
  return errors.New("unknown status")
 }
 const q = `UPDATE parcel SET status = ? WHERE number = ?`
 res, err := s.db.Exec(q, newStatus, number)
 if err != nil {
  return err
 }
 aff, _ := res.RowsAffected()
 if aff == 0 {
  return ErrNotFound
 }
 return nil
}

func (s *ParcelStore) UpdateAddress(number int64, newAddress string) error {
 const q = `UPDATE parcel SET address = ? WHERE number = ? AND status = 'registered'`
 res, err := s.db.Exec(q, newAddress, number)
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

func (s *ParcelStore) Delete(number int64) error {
 const q = `DELETE FROM parcel WHERE number = ? AND status = 'registered'`
 res, err := s.db.Exec(q, number)
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

type ParcelService struct{ store *ParcelStore }

func NewParcelService(store *ParcelStore) *ParcelService { return &ParcelService{store: store} }

func (s *ParcelService) RegisterParcel(clientID int64, address string) (int64, error) {
 return s.store.Create(clientID, address)
}
func (s *ParcelService) GetClientParcels(clientID int64) ([]Parcel, error) {
 return s.store.ListByClient(clientID)
}
func (s *ParcelService) ChangeStatus(number int64, newStatus string) error {
 return s.store.UpdateStatus(number, newStatus)
}
func (s *ParcelService) ChangeAddress(number int64, newAddress string) error {
 return s.store.UpdateAddress(number, newAddress)
}
func (s *ParcelService) Remove(number int64) error { return s.store.Delete(number) }
