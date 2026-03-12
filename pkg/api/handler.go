package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"final_project/pkg/db"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type errJSON struct {
	Code int    `json:"-"`
	Err  string `json:"error"`
}

func (e errJSON) Error() string {
	return e.Err
}

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// NextDayHandler() высчитывает следующую дату
func NextDayHandler(w http.ResponseWriter, r *http.Request) {
	// получаем параметры из строки поиска
	nowStr := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	var err error

	if nowStr == "" {
		now = time.Now()
	} else {
		now, err = time.Parse(frmt, nowStr)
		if err != nil {
			errWriteJSON(w, errJSON{
				Code: http.StatusBadRequest,
				Err:  "Неверный формат now",
			})
			return
		}
	}

	// получаем следующую дату
	result, err := NextDate(now, date, repeat)
	if err != nil {
		errWriteJSON(w, errJSON{
			Code: http.StatusUnprocessableEntity,
			Err:  fmt.Sprintf("Ошибка вычисления следующей даты: %v", err),
		})
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

// AddTaskHandler() добавляет задачу в ДБ
func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	// читаем тело
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		errWriteJSON(w, errJSON{
			Code: http.StatusInternalServerError,
			Err:  fmt.Sprintf("Ошибка чтения тела запроса: %v", err),
		})
		return
	}
	// десериализируем в структуру `task`
	task := &db.Task{}
	if err = json.Unmarshal(buf.Bytes(), task); err != nil {
		errWriteJSON(w, errJSON{
			Code: http.StatusBadRequest,
			Err:  fmt.Sprintf("Ошибка десериализации: %v", err),
		})
		return
	}

	if task.Title == "" {
		errWriteJSON(w, errJSON{
			Code: http.StatusBadRequest,
			Err:  "Не указан заголовок задачи",
		})
		return
	}

	// проверяем дату на корректность
	if err = checkDate(task); err != nil {
		errWriteJSON(w, errJSON{
			Code: http.StatusBadRequest,
			Err:  fmt.Sprintf("Ошибка даты на корректность: %v", err),
		})
		return
	}

	// добавляем в ДБ
	id, err := db.AddTask(task)
	if err != nil {
		errWriteJSON(w, errJSON{
			Code: http.StatusInternalServerError,
			Err:  fmt.Sprintf("Ошибка добавления задачи в ДБ: %v", err),
		})
		return
	}
	n := int(id)
	task.ID = strconv.Itoa(n)

	writeJSON(w, http.StatusOK, &task)
}

// TasksHandler() обрабатывает параметр `search` в строке запроса
func TasksHandler(w http.ResponseWriter, r *http.Request) {
	// получаем параметр `search`
	search := r.FormValue("search")

	// получаем список задач
	tasks, err := db.Tasks(search, 50) // `50` максимальное количество записей
	if err != nil {
		errWriteJSON(w, errJSON{
			Code: http.StatusInternalServerError,
			Err:  fmt.Sprintf("Ошибка получения задач из ДБ: %v", err),
		})
		return
	}
	if tasks == nil {
		tasks = make([]*db.Task, 0)
	}

	writeJSON(w, http.StatusOK, &TasksResp{Tasks: tasks})
}

// TaskByIDHandler() получает, изменяет или удаляет задачу по `id`
func TaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// получаем задачу по `id`
	case http.MethodGet:
		id, err := getId(r)
		if err != nil {
			errWriteJSON(w, errJSON{
				Code: http.StatusBadRequest,
				Err:  fmt.Sprintf("Ошибка получения id из строки запроса: %v", err),
			})
			return
		}
		res, err := db.GetTask(id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				errWriteJSON(w, errJSON{
					Code: http.StatusNotFound,
					Err:  fmt.Sprintf("Задача не найдена: %v", err),
				})
			} else {
				errWriteJSON(w, errJSON{
					Code: http.StatusInternalServerError,
					Err:  fmt.Sprintf("Ошибка получения задачи из ДБ: %v", err),
				})
			}
			return
		}
		writeJSON(w, http.StatusOK, res)

	// изменяем задачу по `id`
	case http.MethodPut:
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			errWriteJSON(w, errJSON{
				Code: http.StatusInternalServerError,
				Err:  fmt.Sprintf("Ошибка чтения тела запроса: %v", err),
			})
			return
		}
		task := &db.Task{}
		if err = json.Unmarshal(buf.Bytes(), task); err != nil {
			errWriteJSON(w, errJSON{
				Code: http.StatusBadRequest,
				Err:  fmt.Sprintf("Ошибка десериализации: %v", err),
			})
			return
		}

		if task.Title == "" {
			errWriteJSON(w, errJSON{
				Code: http.StatusBadRequest,
				Err:  "Не указан заголовок задачи",
			})
			return
		}

		if err = checkDate(task); err != nil {
			errWriteJSON(w, errJSON{
				Code: http.StatusBadRequest,
				Err:  fmt.Sprintf("Ошибка даты на корректность: %v", err),
			})
			return
		}
		err = db.UpdateTask(task)
		if err != nil {
			errWriteJSON(w, errJSON{
				Code: http.StatusInternalServerError,
				Err:  fmt.Sprintf("Ошибка изменения задачи в ДБ: %v", err),
			})
			return
		}
		writeJSON(w, http.StatusOK, db.Task{})

	// удаляем задачу по `id`
	case http.MethodDelete:
		id, err := getId(r)
		if err != nil {
			errWriteJSON(w, errJSON{
				Code: http.StatusBadRequest,
				Err:  fmt.Sprintf("Ошибка получения id из строки запроса: %v", err),
			})
			return
		}

		_, err = db.GetTask(id)
		if err != nil {
			errWriteJSON(w, errJSON{
				Code: http.StatusNotFound,
				Err:  fmt.Sprintf("Ошибка получения задачи из ДБ: %v", err),
			})
			return
		}

		err = db.DeleteTask(id)
		if err != nil {
			errWriteJSON(w, errJSON{
				Code: http.StatusInternalServerError,
				Err:  fmt.Sprintf("Ошибка удаления задачи из ДБ: %v", err),
			})
			return
		}
		writeJSON(w, http.StatusOK, db.Task{})
	}
}

// TaskDoneHandler() делает задачу выполненной
func TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getId(r)
	if err != nil {
		errWriteJSON(w, errJSON{
			Code: http.StatusBadRequest,
			Err:  fmt.Sprintf("Ошибка получения id из строки запроса: %v", err),
		})
		return
	}

	// получаем данные из ДБ по id
	res, err := db.GetTask(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errWriteJSON(w, errJSON{
				Code: http.StatusNotFound,
				Err:  fmt.Sprintf("Задача не найдена: %v", err),
			})
		} else {
			errWriteJSON(w, errJSON{
				Code: http.StatusInternalServerError,
				Err:  fmt.Sprintf("Ошибка получения задачи из ДБ: %v", err),
			})
		}
		return
	}

	// если нет правила - удаляем задачу
	if res.Repeat == "" {
		err = db.DeleteTask(id)
		if err != nil {
			errWriteJSON(w, errJSON{
				Code: http.StatusInternalServerError,
				Err:  fmt.Sprintf("Ошибка удаления задачи из ДБ: %v", err),
			})
			return
		}
		writeJSON(w, http.StatusOK, db.Task{})
		return
	}

	// если правило указано - высчитываем следующую дату
	date, err := NextDate(time.Now(), res.Date, res.Repeat)
	if err != nil {
		errWriteJSON(w, errJSON{
			Code: http.StatusUnprocessableEntity,
			Err:  fmt.Sprintf("Ошибка вычисления следующей даты: %v", err),
		})
		return
	}

	// полученную дату обновляем в ДБ
	err = db.UpdateDate(date, id)
	if err != nil {
		errWriteJSON(w, errJSON{
			Code: http.StatusInternalServerError,
			Err:  fmt.Sprintf("Ошибка изменения задачи в ДБ: %v", err),
		})
		return
	}
	writeJSON(w, http.StatusOK, db.Task{})
}

// getId() получает параметр `id` из строки запроса
func getId(r *http.Request) (string, error) {
	id := r.FormValue("id")
	if id == "" {
		return "", fmt.Errorf("Не указан идентификатор")
	}
	return id, nil
}
