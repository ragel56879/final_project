package db

import (
	"database/sql"
	"fmt"
	"time"
)

// функция AddTask() добавляет задачу в таблицу и возвращает id добавленной задачи
func AddTask(task *Task) (int64, error) {
	var id int64

	// делаем INSERT запрос в базу
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)`
	res, err := DB.Exec(
		query,
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
	)

	// получаем id добавленной задачи, если нет ошибки
	if err == nil {
		id, err = res.LastInsertId()
	}

	return id, err
}

// функция Tasks() делает запрос в ДБ и записывает все задачи в слайс структуры []*Task
func Tasks(search string, limit int) ([]*Task, error) {
	// слайс интерфейсов для хранения аргументов args для запроса query
	var args []interface{}
	var query string

	// формируем SELECT запросы получение задач с фильтром по дате, заголовку или без фильтра
	if date, err := time.Parse("02.01.2006", search); err == nil {
		dateStr := date.Format("20060102")
		query = "SELECT * FROM scheduler WHERE date = :date LIMIT :limit"
		args = append(args, sql.Named("date", dateStr), sql.Named("limit", limit))
	} else if search != "" {
		search = "%" + search + "%"
		query = "SELECT * FROM scheduler WHERE title LIKE :title OR comment LIKE :comment ORDER BY date LIMIT :limit"
		args = append(args, sql.Named("title", search), sql.Named("comment", search), sql.Named("limit", limit))
	} else {
		query = "SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT :limit"
		args = append(args, sql.Named("limit", limit))
	}

	// делаем запрос в ДБ
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// читаем и записываем ответы ДБ в слайс структур []*Task
	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		err = rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// фукция GetTask() получает задачу по id
func GetTask(id string) (*Task, error) {
	// делаем SELECT запрос в ДБ и записываем ответ в структуру Task{}
	row := DB.QueryRow("SELECT * FROM scheduler WHERE id = :id", sql.Named("id", id))
	task := &Task{}
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Задача не найдена")
	} else if err != nil {
		return nil, err
	}

	return task, nil
}

// функция UpdateTask() редактирует задачи
func UpdateTask(task *Task) error {
	// делаем UPDATE запрос в ДБ
	query := `UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id`
	res, err := DB.Exec(
		query,
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
		sql.Named("id", task.ID),
	)
	if err != nil {
		return err
	}

	// получаем кол-во измененных строк
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("Неверный id")
	}
	return nil
}

// функция DeleteTask() удаляет задачу по id
func DeleteTask(id string) error {
	// делаем DELETE запрос в ДБ
	_, err := DB.Exec("DELETE FROM scheduler WHERE id = :id", sql.Named("id", id))
	if err != nil {
		return err
	}

	return nil
}

// фукнкция UpdateDate() изменяет дату задачи по id
func UpdateDate(date, id string) error {
	// делаем UPDATE запрос в ДБ
	_, err := DB.Exec(
		"UPDATE scheduler SET date = :date WHERE id = :id",
		sql.Named("date", date),
		sql.Named("id", id),
	)
	if err != nil {
		return err
	}

	return nil
}
