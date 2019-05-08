package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	v1 "github.com/retailcrm/mg-bot-api-client-go/v1"
)

var timesString = map[string]map[string]string{
	"ru": {
		"минут": "minutes",
		"часов": "hours",
		"часа":  "hours",
		"час":   "hours",
		"день":  "days",
		"дней":  "days",
		"дня":   "days",
	},
	"en": {
		"minutes": "minutes",
		"hours":   "hours",
		"days":    "days",
		"day":     "days",
	},
	"es": {
		"minutos": "minutes",
		"hora":    "hours",
		"horas":   "hours",
		"dia":     "days",
		"dias":    "days",
	},
}

var daysOfWeek = map[string]map[string]time.Weekday{
	"ru": {
		"Воскресенье": time.Sunday,
		"Понедельник": time.Monday,
		"Вторник":     time.Tuesday,
		"Среда":       time.Wednesday,
		"Среду":       time.Wednesday,
		"Четверг":     time.Thursday,
		"Пятница":     time.Friday,
		"Пятницу":     time.Friday,
		"Суббота":     time.Saturday,
		"Субботу":     time.Saturday,
	},
	"en": {
		"Sunday":    time.Sunday,
		"Monday":    time.Monday,
		"Tuesday":   time.Tuesday,
		"Wednesday": time.Wednesday,
		"Thursday":  time.Thursday,
		"Friday":    time.Friday,
		"Saturday":  time.Saturday,
	},
	"es": {
		"Lunes":     time.Sunday,
		"Martes":    time.Monday,
		"Miércoles": time.Tuesday,
		"Jueves":    time.Wednesday,
		"Viernes":   time.Thursday,
		"Sábado":    time.Friday,
		"Domingo":   time.Saturday,
	},
}

var months = map[string]map[string]time.Month{
	"ru": {
		"Янв": time.January,
		"Фев": time.February,
		"Мар": time.March,
		"Апр": time.April,
		"Май": time.May,
		"Июн": time.June,
		"Июл": time.July,
		"Авг": time.August,
		"Сен": time.September,
		"Окт": time.October,
		"Ноя": time.November,
		"Дек": time.December,
	},
	"en": {
		"Jan": time.January,
		"Feb": time.February,
		"Mar": time.March,
		"Apr": time.April,
		"May": time.May,
		"Jun": time.June,
		"Jul": time.July,
		"Aug": time.August,
		"Sep": time.September,
		"Oct": time.October,
		"Nov": time.November,
		"Dec": time.December,
	},
	"es": {
		"Ene":  time.January,
		"Feb":  time.February,
		"Mar":  time.March,
		"Abr":  time.April,
		"May":  time.May,
		"Jun":  time.June,
		"Jul":  time.July,
		"Ago":  time.August,
		"Sept": time.September,
		"Oct":  time.October,
		"Nov":  time.November,
		"Dic":  time.December,
	},
}

type Task struct {
	Connection *Connection
	Message    *v1.Message
	Localizer  *i18n.Localizer
	Offset     int
	Command    string
	What       string
	When       string
	Customer   int
}

func TaskInit(connection *Connection, message *v1.Message, localizer *i18n.Localizer, command string) Task {
	return Task{
		Connection: connection,
		Message:    message,
		Command:    command,
		Localizer:  localizer,
	}
}

func (t *Task) searchWhat() *Task {
	re := regexp.MustCompile(`".*?"`)
	res := re.FindStringSubmatch(t.Command)
	sc := re.FindStringSubmatchIndex(t.Command)

	if len(sc) > 1 {
		t.Offset = sc[1]
	}

	if len(res) > 0 {
		t.What = strings.Trim(res[0], "\"")
	}

	return t
}

func (t *Task) searchWhen() {
	var rt time.Time
	str := t.Command[t.Offset:]

	at := fmt.Sprintf("%s %s", t.getMessage("task_at_time"), "\\d{2}:\\d{2}")
	in := fmt.Sprintf("%s %s (%s)", t.getMessage("task_in_time"), "(\\d+)", t.getMessage("times"))
	on := fmt.Sprintf("%s %s", t.getMessage("task_on_time"), ".*?")

	tm, _ := time.Parse("2006-01-02T15:04:05Z", t.Message.Time)
	ct, _ := convertTime(t.Connection.Timezone, tm)

	switch true {
	case containsPiece(at, str):
		re := regexp.MustCompile(at)
		strs := re.FindStringSubmatchIndex(str)
		t.When = str[strs[0]+len(t.getMessage("task_at_time")):]
		when := strings.Split(strings.Trim(t.When, " "), ":")
		h, _ := strconv.Atoi(when[0])
		m, _ := strconv.Atoi(when[1])
		rt = time.Date(ct.Year(), ct.Month(), ct.Day(), h, m, 0, 0, ct.Location())
	case containsPiece(in, str):
		re := regexp.MustCompile(in)
		strs := re.FindStringSubmatchIndex(str)
		t.When = str[strs[0]+len(t.getMessage("task_in_time")):]
		when := strings.Split(strings.Trim(t.When, " "), " ")

		if len(when) > 1 {
			timeString := timesString[t.Connection.Lang][when[1]]
			value, _ := strconv.ParseInt(when[0], 10, 64)

			if timeString == "minutes" {
				rt = ct.Add(time.Minute * time.Duration(value))
			}

			if timeString == "hours" {
				rt = ct.Add(time.Hour * time.Duration(value))
			}

			if timeString == "days" {
				rt = ct.Add(24 * time.Hour * time.Duration(value))
			}
		}
	case containsPiece(on, str):
		t.When = strings.TrimSpace(str)[len(t.getMessage("task_on_time"))+1:]

		y := ct.Year()
		m := ct.Month()
		d := ct.Day()
		h := ct.Hour()
		min := ct.Minute()

		dayOrMonth := searchSubstring(str, "([A-Z]|[А-Я]){1}([a-z]|[а-я])+")
		clock := searchSubstring(str, "\\d{2}:\\d{2}")
		date := searchSubstring(str, "\\d{4}-\\d{2}-\\d{2}")
		tomorrow := searchSubstring(str, t.getMessage("tomorrow"))

		if date != "" {
			dt := strings.Split(strings.TrimSpace(date), "-")
			y, _ = strconv.Atoi(dt[0])
			month, _ := strconv.Atoi(dt[1])
			m = time.Month(month)
			d, _ = strconv.Atoi(dt[2])
		}

		if tomorrow != "" {
			d = d + 1
		}

		if clock != "" {
			c := strings.Split(strings.TrimSpace(clock), ":")
			h, _ = strconv.Atoi(c[0])
			min, _ = strconv.Atoi(c[1])
		}

		if dayOrMonth != "" {
			day, ok := daysOfWeek[t.Connection.Lang][dayOrMonth]
			if ok {
				diff := setDay(ct, day)
				d = int(diff) + d
			}

			month, ok := months[t.Connection.Lang][dayOrMonth]
			if ok {
				m = month

				dv := searchSubstring(str, "(\\d){1,2}")
				if dv != "" {
					v, _ := strconv.Atoi(dv)
					d = v
				}
			}
		}

		rt = time.Date(y, m, d, h, min, 0, 0, ct.Location())
	}

	t.When = rt.Format("2006-01-02 15:04")
}

func (t *Task) getMessage(message string) string {
	return t.Localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: message})
}

func setDay(ct time.Time, d time.Weekday) (diff time.Weekday) {
	if ct.Weekday() < d {
		diff = d - ct.Weekday()
	} else {
		diff = 7 - (ct.Weekday() - d)
	}

	return
}
