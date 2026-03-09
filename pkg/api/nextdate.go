package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	frmt = "20060102"
)

// функция NextDate() высчитывает следующую дату по полученному правилу
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if dstart == "" {
		return "", fmt.Errorf("В date пусто")
	}

	date, err := time.Parse(frmt, dstart)
	if err != nil {
		return "", err
	}

	// разделяем repeat на правило и число
	sliceRules := strings.Split(repeat, " ")

	// проверка на ошибку символа
	if sliceRules[0] != "y" && sliceRules[0] != "d" {
		return "", fmt.Errorf("Недопустимый символ")
	}

	switch sliceRules[0] {
	case "d":
		// проверка на ошибку длины
		if len(sliceRules) == 1 {
			return "", fmt.Errorf("Не указан интервал в днях")
		}

		num, err := strconv.Atoi(sliceRules[1])
		if err != nil {
			return "", err
		}

		// проверка на ошибку интервала для "d"
		if num >= 400 || num < 0 {
			return "", fmt.Errorf("num должен быть в диапазоне от 1 до 400")
		}

		// добавляем дни
		for {
			date = date.AddDate(0, 0, num)
			if afterNow(date, now) {
				break
			}
		}
	case "y":
		// проверка на ошибку длины
		if len(sliceRules) > 1 {
			return "", fmt.Errorf(`Параметр "y" не требует дополнительных уточнений`)
		}

		// добавляем год
		for {
			date = date.AddDate(1, 0, 0)
			if afterNow(date, now) {
				break
			}
		}
	}

	return date.Format(frmt), nil
}

func afterNow(date, now time.Time) bool {
	return date.Format(frmt) > now.Format(frmt)
}
