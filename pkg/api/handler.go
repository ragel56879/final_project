package api

import (
	"bytes"
	"encoding/json"
	"final_project/pkg/db"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type ErrResponse struct {
	Error string `json:"error"`
}

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

var (
	errResp ErrResponse
)

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
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// получаем следующую дату
	result, err := NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(result))
}

// AddTaskHandler() добавляет задачу в ДБ
func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	// читаем тело
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		errResp.Error = err.Error()
		errJson(w, errResp)
		return
	}
	// десериализируем в структуру `task`
	task := &db.Task{}
	if err = json.Unmarshal(buf.Bytes(), task); err != nil {
		errResp.Error = err.Error()
		errJson(w, errResp)
		return
	}

	if task.Title == "" {
		errResp.Error = "Не указан заголовок задачи"
		errJson(w, errResp)
		return
	}

	// проверяем дату на корректность
	if err = checkDate(task); err != nil {
		errResp.Error = err.Error()
		errJson(w, errResp)
		return
	}

	// добавляем в ДБ
	id, err := db.AddTask(task)
	if err != nil {
		errResp.Error = err.Error()
		errJson(w, errResp)
		return
	}
	n := int(id)
	task.ID = strconv.Itoa(n)

	writeJson(w, &task)
}

// TasksHandler() обрабатывает параметр `search` в строке запроса
func TasksHandler(w http.ResponseWriter, r *http.Request) {
	// получаем параметр `search`
	search := r.FormValue("search")

	// получаем список задач
	tasks, err := db.Tasks(search, 50) // `50` максимальное количество записей
	if err != nil {
		errResp.Error = err.Error()
		errJson(w, errResp)
		return
	}
	if tasks == nil {
		tasks = make([]*db.Task, 0)
	}

	writeJson(w, &TasksResp{
		Tasks: tasks,
	})
}

// TaskByIDHandler() получает, изменяет или удаляет задачу по `id`
func TaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// получаем задачу по `id`
	case http.MethodGet:
		id, err := getId(r)
		if err != nil {
			errResp.Error = err.Error()
			errJson(w, errResp)
			return
		}
		res, err := db.GetTask(id)
		if err != nil {
			errResp.Error = err.Error()
			errJson(w, errResp)
			return
		}
		writeJson(w, res)

	// изменяем задачу по `id`
	case http.MethodPut:
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			errResp.Error = err.Error()
			errJson(w, errResp)
			return
		}
		task := &db.Task{}
		if err = json.Unmarshal(buf.Bytes(), task); err != nil {
			errResp.Error = err.Error()
			errJson(w, errResp)
			return
		}

		if task.Title == "" {
			errResp.Error = "Не указан заголовок задачи"
			errJson(w, errResp)
			return
		}

		if err = checkDate(task); err != nil {
			errResp.Error = err.Error()
			errJson(w, errResp)
			return
		}
		err = db.UpdateTask(task)
		if err != nil {
			errResp.Error = err.Error()
			errJson(w, errResp)
			return
		}
		writeJson(w, db.Task{})

	// удаляем задачу по `id`
	case http.MethodDelete:
		id, err := getId(r)
		if err != nil {
			errResp.Error = err.Error()
			errJson(w, errResp)
			return
		}

		_, err = db.GetTask(id)
		if err != nil {
			errResp.Error = err.Error()
			errJson(w, errResp)
			return
		}

		err = db.DeleteTask(id)
		if err != nil {
			errResp.Error = err.Error()
			errJson(w, errResp)
			return
		}
		writeJson(w, db.Task{})
	}
}

// TaskDoneHandler() делает задачу выполненной
func TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	id, err := getId(r)
	if err != nil {
		errResp.Error = err.Error()
		errJson(w, errResp)
		return
	}

	// получаем данные из ДБ по id
	res, err := db.GetTask(id)
	if err != nil {
		log.Print("err: ", err)
		errResp.Error = err.Error()
		errJson(w, errResp)
		return
	}

	// если нет правила - удаляем задачу
	if res.Repeat == "" {
		err = db.DeleteTask(id)
		if err != nil {
			errResp.Error = err.Error()
			errJson(w, errResp)
			return
		}
		writeJson(w, db.Task{})
		return
	}

	// если правило указано - высчитываем следующую дату
	date, err := NextDate(time.Now(), res.Date, res.Repeat)
	if err != nil {
		errResp.Error = err.Error()
		errJson(w, errResp)
		return
	}

	// полученную дату обновляем в ДБ
	err = db.UpdateDate(date, id)
	if err != nil {
		errResp.Error = err.Error()
		errJson(w, errResp)
		return
	}
	writeJson(w, db.Task{})
}

// getId() получает параметр `id` из строки запроса
func getId(r *http.Request) (string, error) {
	id := r.FormValue("id")
	if id == "" {
		return "", fmt.Errorf("Не указан идентификатор")
	}
	return id, nil
}
