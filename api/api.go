package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/shaardie/mhealth/storage"
)

type Config struct {
	Port int `yaml:"port"`
}

type Server struct {
	cfg    Config
	router *http.ServeMux
	db     storage.DB
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

func (s Server) Run() {
	http.ListenAndServe(fmt.Sprintf(":%v", s.cfg.Port), s.router)
}

func Init(cfg Config, db storage.DB) (*Server, error) {
	s := &Server{
		cfg:    cfg,
		router: http.NewServeMux(),
		db:     db,
	}
	s.router.HandleFunc("/health", s.getHealth())
	return s, nil
}
