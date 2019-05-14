package main

import (
	"flag"
	"fmt"
	"github.com/766b/chi-prometheus"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/mfamador/go-payments-api/pkg/admin"
	"github.com/mfamador/go-payments-api/pkg/health"
	"github.com/mfamador/go-payments-api/pkg/logger"
	"github.com/mfamador/go-payments-api/pkg/payments"
	"github.com/mfamador/go-payments-api/pkg/util"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/middleware/stdlib"
	"github.com/ulule/limiter/drivers/store/memory"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	listen             *string
	limit              *string
	compress           *bool
	metrics            *bool
	repoDriver         *string
	repoUri            *string
	repoMigrations     *string
	repoSchemaPayments *string
	enableCors         *bool
	timeout            *int
	adminRoutes        *bool
	profiling          *bool
	apiVersion         *string
	externalUrl        *string
	maxResults         *int
)

func init() {
	listen = flag.String("listen", ":8080", "the http interface to listen at")
	limit = flag.String("limit", "", "rate limit (eg. 5-S for 5 reqs/second)")
	compress = flag.Bool("compress", false, "gzip responses")
	metrics = flag.Bool("metrics", false, "expose prometheus metrics")
	enableCors = flag.Bool("cors", false, "enable cors")
	timeout = flag.Int("timeout", 60, "request timeout")
	repoDriver = flag.String("repo", "sqlite3", "type of persistence repository to use, eg. sqlite3, postgres")
	repoUri = flag.String("repo-uri", "", "repo specific connection string")
	repoMigrations = flag.String("repo-migrations", "./schema", "path to database migrations")
	repoSchemaPayments = flag.String("repo-schema-payments", "payments", "the table or schema where we store payments")
	adminRoutes = flag.Bool("admin", false, "enable admin endpoints")
	profiling = flag.Bool("profiling", false, "enable profiling")
	apiVersion = flag.String("api-version", "v1", "api version to expose our services at")
	externalUrl = flag.String("external-url", "http://localhost:8080", "url to access our microservice from the outside")
	maxResults = flag.Int("max-results", 20, "Maximum number of results when listing items (eg. payments)")
}

func main() {
	flag.Parse()

	baseUrl := fmt.Sprintf("%s/%s", *externalUrl, *apiVersion)

	paymentsRepo, err := util.NewRepo(util.RepoConfig{
		Driver:     *repoDriver,
		Uri:        *repoUri,
		Migrations: *repoMigrations,
		Schema:     *repoSchemaPayments,
	})

	if err != nil {
		log.Fatal(errors.Wrap(err, "Could not create repo"))
	}

	err = paymentsRepo.Init()
	if err != nil {
		log.Fatal(errors.Wrap(err, "Could not init repo"))
	}

	defer paymentsRepo.Close()
	if err := paymentsRepo.Check(); err != nil {
		log.Fatal(errors.Wrap(err, "Could connect to the repo"))
	}

	router := chi.NewRouter()

	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.Timeout(time.Duration(*timeout)*time.Second),
		middleware.RedirectSlashes,
		middleware.Recoverer,
		middleware.RequestID,
		middleware.RealIP,
		middleware.AllowContentType("application/json", "text/plain"),
		middleware.NoCache,
	)

	if *metrics {
		router.Use(chiprometheus.NewMiddleware("payments"))
	}

	router.Use(logger.NewHttpLogger())

	if *compress {
		router.Use(middleware.DefaultCompress)
	}

	if *limit != "" {
		rate, err := limiter.NewRateFromFormatted(*limit)
		if err != nil {
			log.Fatal(errors.Wrap(err, "Error setting rate limit"))
		}
		store := memory.NewStore()
		router.Use(stdlib.NewMiddleware(limiter.New(store, rate)).Handler)
	}

	if *enableCors {
		cors := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		})
		router.Use(cors.Handler)
	}

	if *metrics {
		router.Mount("/metrics", prometheus.Handler())
	}

	if *profiling {
		router.Mount("/profiling", middleware.Profiler())
	}

	router.Mount("/health", health.New(paymentsRepo).Routes())

	if *adminRoutes {
		router.Route("/admin", func(adminRouter chi.Router) {
			adminRouter.Mount("/", admin.New(paymentsRepo).Routes())
		})
	}

	router.Route("/v1", func(v1Router chi.Router) {

		// payments api
		v1Router.Mount("/", payments.New(paymentsRepo, baseUrl, *maxResults).Routes())

		// more endpoints here...
	})

	if err := chi.Walk(router, func(method string,
		route string,
		handler http.Handler,
		middlewares ...func(http.Handler) http.Handler) error {
		route = strings.Replace(route, "/*/", "/", -1)
		logger.Info("Mounted route", &RouteInfo{Method: method, Path: route})
		return nil
	}); err != nil {
		log.Printf(err.Error())
	}

	logger.Info("Started server", &ServerInfo{
		ExternalUrl: *externalUrl,
		ApiVersion:  *apiVersion,
		Interface:   *listen,
	})
	log.Fatal(http.ListenAndServe(*listen, router))
}

type RouteInfo struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type ServerInfo struct {
	Interface   string `json:"interface"`
	ExternalUrl string `json:"externalUrl"`
	ApiVersion  string `json:"apiVersion"`
}
