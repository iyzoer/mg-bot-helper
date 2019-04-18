package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	v5 "github.com/retailcrm/api-client-go/v5"
)

// GenerateToken function
func GenerateToken() string {
	c := atomic.AddUint32(&tokenCounter, 1)

	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d%d", time.Now().UnixNano(), c))))
}

func getAPIClient(url, key string) (*v5.Client, error, int) {
	client := v5.New(url, key)

	cr, _, e := client.APICredentials()
	if e != nil {
		return nil, e, http.StatusInternalServerError
	}

	if !cr.Success {
		return nil, errors.New(getLocalizedMessage("incorrect_url_key")), http.StatusBadRequest
	}

	if res := checkCredentials(cr.Credentials); len(res) != 0 {
		return nil,
			errors.New(localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "missing_credentials",
				TemplateData: map[string]interface{}{
					"Credentials": strings.Join(res, ", "),
				},
			})),
			http.StatusBadRequest
	}

	return client, nil, 0
}

func checkCredentials(credential []string) []string {
	rc := make([]string, len(botCredentials))
	copy(rc, botCredentials)

	for _, vc := range credential {
		for kn, vn := range rc {
			if vn == vc {
				if len(rc) == 1 {
					rc = rc[:0]
					break
				}
				rc = append(rc[:kn], rc[kn+1:]...)
			}
		}
	}

	return rc
}

func getCommandsHash() (hash string, err error) {
	res, err := json.Marshal(getBotCommands())

	h := sha1.New()
	h.Write(res)
	hash = fmt.Sprintf("%x", h.Sum(nil))

	return
}

func containsPiece(piece string, str string) (res bool) {
	var err error
	res, err = regexp.MatchString(piece, str)
	if err != nil {
		return
	}

	return
}

func searchSubstring(s string, t string) (res string) {
	re := regexp.MustCompile(t)
	strs := re.FindStringSubmatch(s)

	if len(strs) > 0 {
		res = strs[0]
	}

	return
}

func convertTime(timezone string, t time.Time) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return t.In(loc), err
	}

	return t.In(loc), nil
}

func getTimezones() ([]string, error) {
	var timezones []string

	data, err := ioutil.ReadFile("./timezones.json")
	if err != nil {
		return timezones, err
	}

	err = json.Unmarshal(data, &timezones)
	if err != nil {
		return timezones, err
	}

	return timezones, nil
}
