package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

const (
	schemaTbl = `CREATE TABLE scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date CHAR(8) NOT NULL DEFAULT '',
		title VARCHAR(128) NOT NULL DEFAULT '',
		comment TEXT,
		repeat VARCHAR(128) NOT NULL DEFAULT ''
		);`
	schemaInd = `CREATE INDEX date_ind ON scheduler (date);`
)

type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date,omitempty"`
	Title   string `json:"title,omitempty"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

var DB *sql.DB

func (t Task) String() string {
	query := fmt.Sprintf(
		"id: %s, date: %s, title: %s, comment: %s, repeat: %s",
		t.ID,
		t.Date,
		t.Title,
		t.Comment,
		t.Repeat,
	)
	return query
}

func Init(dbFile string) error {
	// проверка на существование дб
	_, err := os.Stat(dbFile)

	var install bool
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			install = true
			log.Print("ДБ не существует - создаем")
		} else {
			return err
		}
	} else {
		log.Print("ДБ существует")
	}

	// открываем бд или создаем если не существует
	DB, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	// если бд пустая - добавляем таблицу и индекс
	if install {
		log.Print("ДБ пустая - добавляем таблицу")
		_, err := DB.Exec(schemaTbl)
		if err != nil {
			return err
		}

		_, err = DB.Exec(schemaInd)
		if err != nil {
			return err
		}
	}

	return nil
}
