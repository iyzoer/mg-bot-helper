package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/getsentry/raven-go"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
)

func init() {
	parser.AddCommand("run",
		"Run bot",
		"Run bot.",
		&RunCommand{},
	)
}

var (
	sentry *raven.Client
	wm     = NewWorkersManager()
)

// RunCommand struct
type RunCommand struct{}

// Execute command
func (x *RunCommand) Execute(args []string) error {
	config = LoadConfig(options.Config)
	orm = NewDb(config)
	logger = newLogger()

	go start()

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	for sig := range c {
		switch sig {
		case os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM:
			orm.DB.Close()
			return nil
		default:
		}
	}

	return nil
}

func start() {
	router := setup()
	startWS()
	router.Run(config.HTTPServer.Listen)
}

func setup() *gin.Engine {
	loadTranslateFile()
	updateCommands()
	setValidation()

	if config.Debug == false {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	if config.Debug {
		r.Use(gin.Logger())
	}

	r.HTMLRender = createHTMLRender()

	r.Static("/static", "./static")

	r.Use(func(c *gin.Context) {
		setLocale(c.GetHeader("Accept-Language"))
	})

	errorHandlers := []ErrorHandlerFunc{
		PanicLogger(),
		ErrorResponseHandler(),
	}

	sentry, _ = raven.New(config.SentryDSN)
	if sentry != nil {
		errorHandlers = append(errorHandlers, ErrorCaptureHandler(sentry, true))
	}

	r.Use(ErrorHandler(errorHandlers...))

	r.GET("/", checkAccountForRequest(), connectHandler)
	r.Any("/settings/:uid", settingsHandler)
	r.POST("/save/", checkConnectionForRequest(), saveHandler)
	r.POST("/create/", checkConnectionForRequest(), createHandler)
	r.POST("/bot-settings/", botSettingsHandler)
	r.POST("/actions/activity", activityHandler)

	return r
}

func createHTMLRender() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("home", "templates/layout.html", "templates/home.html")
	r.AddFromFiles("form", "templates/layout.html", "templates/form.html")
	return r
}

func checkAccountForRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		ra := rx.ReplaceAllString(c.Query("account"), ``)
		p := Connection{
			APIURL: ra,
		}

		c.Set("account", p)
	}
}

func checkConnectionForRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		var conn Connection

		if err := c.ShouldBindJSON(&conn); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": getLocalizedMessage("incorrect_url_key")})
			return
		}
		conn.NormalizeApiUrl()

		c.Set("connection", conn)
	}
}

func startWS() {
	res := getActiveConnection()
	if len(res) > 0 {
		for _, v := range res {
			wm.setWorker(v)
		}
	}
}
