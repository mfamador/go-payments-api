package health

import (
	"github.com/go-chi/chi"
	. "github.com/mfamador/go-payments-api/pkg/util"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type HealthService struct {
	repo Repo
}

func New(repo Repo) *HealthService {
	return &HealthService{repo: repo}
}

func (s *HealthService) Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/", s.Get)
	return router
}

func (s *HealthService) Get(w http.ResponseWriter, r *http.Request) {
	log.Info("Health check")

	statusCode := http.StatusOK
	statusMsg := "up"
	if s.repo.Check() != nil {
		statusCode = http.StatusServiceUnavailable
		statusMsg = "down"
	}
	RenderJSON(w, r, statusCode, &Health{
		Status: statusMsg,
	})
}
