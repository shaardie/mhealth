package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/shaardie/mhealth/storage"
)

type Config struct {
	Port int `yaml:"port"`
}

type Server struct {
	cfg                    Config
	router                 *http.ServeMux
	db                     storage.DB
	metricNumberOfRuns     prometheus.GaugeVec
	metricNumberOfFailures prometheus.GaugeVec
}

func (s *Server) getHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var number int
		err := s.db.GetFailedChecks.QueryRowContext(r.Context()).Scan(&number)
		if err != nil {
			log.Printf("Failed to connect to database, %v", err)
			http.Error(w, "NOT OK", http.StatusInternalServerError)
			return
		}
		if number > 0 {
			http.Error(w, "NOT OK", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "OK")
	}
}

func (s *Server) getMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			checkType        string
			name             string
			numberOfRuns     int
			numberOfFailures int
		)
		rows, err := s.db.GetChecks.QueryContext(r.Context())
		if err != nil {
			log.Printf("Failed to connect to database, %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&checkType, &name, &numberOfRuns, &numberOfFailures)
			if err != nil {
				log.Printf("Failed to scan row to database, %v", err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
			s.metricNumberOfRuns.WithLabelValues(checkType, name).Set(float64(numberOfRuns))
			s.metricNumberOfFailures.WithLabelValues(checkType, name).Set(float64(numberOfFailures))
		}
		promhttp.Handler().ServeHTTP(w, r)
	}
}

func (s Server) Run() {
	http.ListenAndServe(fmt.Sprintf(":%v", s.cfg.Port), s.router)
}

func Init(cfg Config, db storage.DB) (*Server, error) {
	s := &Server{
		cfg:    cfg,
		router: http.NewServeMux(),
		db:     db,
		metricNumberOfRuns: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "number_of_runs",
				Help: "How often a check has run",
			},
			[]string{"type", "name"},
		),
		metricNumberOfFailures: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "number_of_failures",
				Help: "Number of consecutive failures of a check",
			},
			[]string{"type", "name"},
		),
	}

	prometheus.MustRegister(s.metricNumberOfRuns)
	prometheus.MustRegister(s.metricNumberOfFailures)

	s.router.HandleFunc("/health", s.getHealth())
	s.router.HandleFunc("/metrics", s.getMetrics())
	return s, nil
}
