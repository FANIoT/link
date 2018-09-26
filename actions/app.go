package actions

import (
	"net/http"
	"strconv"
	"time"

	"github.com/I1820/link/core"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/ssl"
	"github.com/gobuffalo/envy"
	"github.com/unrolled/secure"

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
		// Automatically redirect to SSL
		app.Use(ssl.ForceSSL(secure.Options{
			SSLRedirect:     ENV == "production",
			SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
		}))

		// set the request content type to JSON (until new version of buffalo)
		app.Use(middleware.SetContentType("application/json"))
		app.Use(func(next buffalo.Handler) buffalo.Handler {
			return func(c buffalo.Context) error {
				defer func() {
					c.Response().Header().Set("Content-Type", "application/json")
				}()

				return next(c)
			}
		})

		if ENV == "development" {
			app.Use(middleware.ParameterLogger)
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

		prometheus.NewGoCollector()
		prometheus.MustRegister(rds)

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
				}()

				return next(c)
			}
		})

		// Routes
		app.GET("/about", AboutHandler)
		// mqtt service (authorization module)
		mqtt := app.Group("/mqtt")
		{
			mqtt.POST("/auth", func(c buffalo.Context) error {
				return c.Render(http.StatusOK, r.JSON(true))
			})
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
