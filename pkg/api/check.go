package api

import (
	"final_project/pkg/db"
	"time"
)

// checkDate() проверяет дату на корректность
func checkDate(task *db.Task) error {
	now := time.Now()

	if task.Date == "" {
		task.Date = now.Format(frmt)
	}

	t, err := time.Parse(frmt, task.Date)
	if err != nil {
		return err
	}

	// если `now` > `Date` и в `Repeat` ничего, то `Date` = `now`
	// иначе высчитываем следующую дату для `Date`
	if afterNow(now, t) {
		if len(task.Repeat) == 0 {
			task.Date = now.Format(frmt)
		} else {
			task.Date, err = NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
