package actions

import (
	"strconv"
	"time"

	"github.com/FANIoT/link/core"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	contenttype "github.com/gobuffalo/mw-contenttype"
	paramlogger "github.com/gobuffalo/mw-paramlogger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gobuffalo/x/sessions"
	"github.com/rs/cors"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App
var coreApp *core.Application

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:          ENV,
			SessionStore: sessions.Null{},
			PreWares: []buffalo.PreWare{
				cors.Default().Handler,
			},
			SessionName: "_link_session",
		})

		// If no content type is sent by the client
		// the application/json will be set, otherwise the client's
		// content type will be used.
		app.Use(contenttype.Add("application/json"))

		if ENV == "development" {
			app.Use(paramlogger.ParameterLogger)
		}

		// core application provides a simple way for parse and store
		// incoming data
		coreApp = core.New()
		coreApp.Run()

		// prometheus collectors
		rds := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "link",
				Name:      "request_duration_seconds",
				Help:      "A histogram of latencies for requests.",
			},
			[]string{"path", "method", "code"},
		)

		rc := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "link",
				Name:      "request_counter",
				Help:      "How many HTTP requests processed",
			},
			[]string{"path", "method", "code"},
		)

		prometheus.NewGoCollector()
		prometheus.MustRegister(rds)
		prometheus.MustRegister(rc)

		app.Use(func(next buffalo.Handler) buffalo.Handler {
			return func(c buffalo.Context) error {
				now := time.Now()

				defer func() {
					ws := c.Response().(*buffalo.Response)
					req := c.Request()

					rds.With(prometheus.Labels{
						"path":   req.URL.String(),
						"code":   strconv.Itoa(ws.Status),
						"method": req.Method,
					}).Observe(time.Since(now).Seconds())

					rc.With(prometheus.Labels{
						"path":   req.URL.String(),
						"code":   strconv.Itoa(ws.Status),
						"method": req.Method,
					}).Inc()
				}()

				return next(c)
			}
		})

		// Routes
		app.GET("/about", AboutHandler)
		// mqtt service (authorization module)
		mqtt := app.Group("/mqtt")
		{
			vmq := VernemqAuthPlugin{}
			mqtt.POST("/auth/publish", vmq.OnPublish)
			mqtt.POST("/auth/subscribe", vmq.OnSubscribe)
		}
		// http service
		http := app.Group("/http")
		{
			http.Use(HTTPAuthorize)
			http.POST("/push/{thing_id}", HTTPHandler)
		}
		// ttn integration module
		ttn := app.Group("/ttn")
		{
			ttn.Use(TTNAuthorize)
			ttn.POST("/{project_id}", TTNHandler)
		}
		app.GET("/metrics", buffalo.WrapHandler(promhttp.Handler()))
	}

	return app
}
