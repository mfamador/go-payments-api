package admin

import (
	"github.com/go-chi/chi"
	. "github.com/mfamador/go-payments-api/pkg/util"
	"net/http"
)

type AdminService struct {
	repo Repo
}

func New(repo Repo) *AdminService {
	return &AdminService{repo: repo}
}

func (s *AdminService) Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Route("/repo", func(r chi.Router) {
		r.Delete("/", s.DeleteRepo)
		r.Get("/", s.GetRepo)
	})
	return router
}

func (s *AdminService) DeleteRepo(w http.ResponseWriter, r *http.Request) {
	err := s.repo.DeleteAll()
	if err != nil {
		HandleHttpError(w, r, http.StatusInternalServerError, err)
		return
	}
	RenderNoContent(w, r)
}

func (s *AdminService) GetRepo(w http.ResponseWriter, r *http.Request) {
	info, err := s.repo.Info()
	if err != nil {
		HandleHttpError(w, r, http.StatusInternalServerError, err)
		return
	}
	RenderJSON(w, r, http.StatusOK, info)
}
