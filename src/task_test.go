package main

import (
	"testing"
	"time"

	v1 "github.com/retailcrm/mg-bot-api-client-go/v1"
)

const layout = "2006-01-02T15:04:05Z"

var et time.Time

func createTask(lang string, command string) Task {
	return TaskInit(
		&Connection{
			Timezone: "Europe/Moscow",
			Lang:     lang,
		},
		&v1.Message{
			Time: string(time.Now().Format(layout)),
		},
		getLang(lang),
		command,
	)
}

func TestSearchWhenEn(t *testing.T) {
	tasks := []string{
		"at 00:00",
		"in 2 days",
		"in 15 minutes",
		"on 2019-01-01",
		"on tomorrow",
		"on 1 Jun 15:00",
		"on Monday",
	}

	for _, when := range tasks {
		task := createTask("en", "/task \"Task\" "+when)
		task.searchWhen()

		if et.Format("2006-01-02 15:04") == task.When {
			t.Error("Error parsing `when`")
		}
	}
}

func TestSearchWhenRu(t *testing.T) {
	tasks := []string{
		"в 00:00",
		"через 2 дня",
		"через 15 минут",
		"на 2019-01-01",
		"на завтра",
		"на 1 Июн 15:00",
		"на Понедельник",
	}

	for _, when := range tasks {
		task := createTask("ru", "/task \"Task\" "+when)
		task.searchWhen()

		if et.Format("2006-01-02 15:04") == task.When {
			t.Error("Error parsing `when`")
		}
	}
}

func TestSearchWhat(t *testing.T) {
	task := createTask("en", "/task \"Task\" at 00:00")
	task.searchWhat()

	if "" == task.What {
		t.Error("Error parsing `what`")
	}
}

func TestSearchWhatIncorrect(t *testing.T) {
	task := createTask("en", "/task Task at 00:00")
	task.searchWhat()

	if "" != task.What {
		t.Error("Error parsing `what`")
	}
}
