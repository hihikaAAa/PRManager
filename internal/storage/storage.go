package storage

import (
	"context"
	"fmt"
	"time"

	"database/sql"
)

func New(dsn string) (*sql.DB,error){
	const op = "internal.storage.New"

	db, err := sql.Open("postgres",dsn)
	if err != nil{
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(),5 * time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err!= nil{
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return db, nil
}