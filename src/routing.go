package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	v5 "github.com/retailcrm/api-client-go/v5"
)

func connectHandler(c *gin.Context) {
	res := struct {
		Conn   Connection
		Locale map[string]interface{}
		Year   int
	}{
		c.MustGet("account").(Connection),
		getLocale(),
		time.Now().Year(),
	}

	c.HTML(http.StatusOK, "home", &res)
}

func botSettingsHandler(c *gin.Context) {
	jm := map[string]string{}

	if err := c.ShouldBindJSON(&jm); err != nil {
		c.Error(err)
		return
	}

	conn := getConnection(jm["client_id"])
	conn.Lang = jm["lang"]
	conn.Currency = jm["currency"]
	conn.Timezone = jm["timezone"]

	err := conn.saveConnection()
	if err != nil {
		c.Error(err)
		return
	}

	wm.setWorker(conn)

	c.JSON(http.StatusOK, gin.H{"msg": getLocalizedMessage("successful")})
}

func settingsHandler(c *gin.Context) {
	uid := c.Param("uid")
	p := getConnection(uid)
	if p.ID == 0 {
		c.Redirect(http.StatusFound, "/")
		return
	}

	timezones, e := getTimezones()
	if e != nil {
		logger.Debugf("Getting timezones error: %s", e.Error())
	}

	res := struct {
		Conn         *Connection
		Locale       map[string]interface{}
		Year         int
		LangCode     []string
		CurrencyCode map[string]string
		Timezones    []string
	}{
		p,
		getLocale(),
		time.Now().Year(),
		[]string{"en", "ru", "es"},
		currency,
		timezones,
	}

	c.HTML(200, "form", res)
}

func saveHandler(c *gin.Context) {
	conn := c.MustGet("connection").(Connection)

	_, err, code := getAPIClient(conn.APIURL, conn.APIKEY)
	if err != nil {
		if code == http.StatusInternalServerError {
			c.Error(err)
		} else {
			c.JSON(code, gin.H{"error": err.Error()})
		}
		return
	}

	err = conn.saveConnection()
	if err != nil {
		c.Error(err)
		return
	}

	wm.setWorker(&conn)

	c.JSON(http.StatusOK, gin.H{"msg": getLocalizedMessage("successful")})
}

func createHandler(c *gin.Context) {
	conn := c.MustGet("connection").(Connection)

	cl := getConnectionByURL(conn.APIURL)
	if cl.ID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": getLocalizedMessage("connection_already_created")})
		return
	}

	client, err, code := getAPIClient(conn.APIURL, conn.APIKEY)
	if err != nil {
		if code == http.StatusInternalServerError {
			c.Error(err)
		} else {
			c.JSON(code, gin.H{"error": err.Error()})
		}
		return
	}

	conn.ClientID = GenerateToken()

	data, status, e := client.IntegrationModuleEdit(getIntegrationModule(conn.ClientID))
	if e != nil {
		c.Error(e)
		return
	}

	if status >= http.StatusBadRequest {
		c.JSON(http.StatusBadRequest, gin.H{"error": getLocalizedMessage("error_activity_mg")})
		logger.Error(conn.APIURL, status, e, data)
		return
	}

	conn.MGURL = data.Info.MgBotInfo.EndpointUrl
	conn.MGToken = data.Info.MgBotInfo.Token
	conn.Active = true
	conn.Lang = "ru"
	conn.Currency = currency["Российский рубль"]

	bj, _ := json.Marshal(botCommands)
	conn.Commands.RawMessage = bj

	code, err = SetBotCommand(conn.MGURL, conn.MGToken)
	if err != nil {
		c.JSON(code, gin.H{"error": getLocalizedMessage("error_activity_mg")})
		logger.Error(conn.APIURL, code, err)
		return
	}

	err = conn.createConnection()
	if err != nil {
		c.Error(err)
		return
	}

	wm.setWorker(&conn)

	c.JSON(
		http.StatusCreated,
		gin.H{
			"url":     "/settings/" + conn.ClientID,
			"message": getLocalizedMessage("successful"),
		},
	)
}

func activityHandler(c *gin.Context) {
	var (
		activity  v5.Activity
		systemUrl = c.PostForm("systemUrl")
		clientId  = c.PostForm("clientId")
	)

	conn := getConnection(clientId)
	if conn.ID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{
				"success": false,
				"error":   "Wrong data",
			},
		)
		return
	}

	err := json.Unmarshal([]byte(c.PostForm("activity")), &activity)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{
				"success": false,
				"error":   "Wrong data",
			},
		)
		return
	}

	conn.Active = activity.Active && !activity.Freeze

	if systemUrl != "" {
		conn.APIURL = systemUrl
	}
	conn.NormalizeApiUrl()

	if err := conn.setConnectionActivity(); err != nil {
		c.Error(err)
		return
	}

	if !conn.Active {
		wm.stopWorker(conn)
	} else {
		wm.setWorker(conn)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func getIntegrationModule(clientId string) v5.IntegrationModule {
	return v5.IntegrationModule{
		Code:            config.BotInfo.Code,
		IntegrationCode: config.BotInfo.Code,
		Active:          true,
		Name:            config.BotInfo.Name,
		ClientID:        clientId,
		Logo: fmt.Sprintf(
			"https://%s%s",
			config.HTTPServer.Host,
			config.BotInfo.LogoPath,
		),
		BaseURL: fmt.Sprintf(
			"https://%s",
			config.HTTPServer.Host,
		),
		AccountURL: fmt.Sprintf(
			"https://%s/settings/%s",
			config.HTTPServer.Host,
			clientId,
		),
		Actions: map[string]string{"activity": "/actions/activity"},
		Integrations: &v5.Integrations{
			MgBot: &v5.MgBot{},
		},
	}
}
