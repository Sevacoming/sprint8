package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// добавляем посылку
	res, err := s.db.Exec(
		`INSERT INTO parcel (client, status, address, created_at)
         VALUES (?, ?, ?, ?)`,
		p.Client,
		p.Status,
		p.Address,
		p.CreatedAt,
	)
	if err != nil {
		return 0, err
	}

	// возвращаем идентификатор
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// читаем одну строку по number
	var p Parcel
	err := s.db.QueryRow(
		`SELECT number, client, status, address, created_at
         FROM parcel
         WHERE number = ?`,
		number,
	).Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		// ревьюер просил явно обработать ошибку Scan
		return Parcel{}, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// читаем все строки по client
	rows, err := s.db.Query(
		`SELECT number, client, status, address, created_at
         FROM parcel
         WHERE client = ?`,
		client,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Parcel
	for rows.Next() {
		var p Parcel
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, p)
	}

	// ревьюер отдельно писал, что rows.Err() нужно явно обработать
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// обновляем статус
	_, err := s.db.Exec(
		`UPDATE parcel
         SET status = ?
         WHERE number = ?`,
		status,
		number,
	)
	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// менять адрес можно только если статус registered
	var status string
	err := s.db.QueryRow(
		`SELECT status
         FROM parcel
         WHERE number = ?`,
		number,
	).Scan(&status)
	if err != nil {
		return err
	}

	if status != ParcelStatusRegistered {
		return fmt.Errorf("cannot change address for parcel %d with status %q", number, status)
	}

	_, err = s.db.Exec(
		`UPDATE parcel
         SET address = ?
         WHERE number = ?`,
		address,
		number,
	)
	return err
}

func (s ParcelStore) Delete(number int) error {
	// удалять можно только если статус registered
	var status string
	err := s.db.QueryRow(
		`SELECT status
         FROM parcel
         WHERE number = ?`,
		number,
	).Scan(&status)
	if err != nil {
		return err
	}

	if status != ParcelStatusRegistered {
		return fmt.Errorf("cannot delete parcel %d with status %q", number, status)
	}

	_, err = s.db.Exec(
		`DELETE FROM parcel
         WHERE number = ?`,
		number,
	)
	return err
}
