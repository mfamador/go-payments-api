package payments

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	. "github.com/mfamador/go-payments-api/pkg/util"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
)

var (
	maxResults          int
	paymentsLinkPattern string
	paymentLinkPattern  string
)

func init() {
	paymentsLinkPattern = "/payments?from=%v&to=%v"
	paymentLinkPattern = "/payment/%v"
}

type PaymentsService struct {
	HttpService
	repo       Repo
	maxResults int
}

func New(repo Repo, baseUrl string, maxResults int) *PaymentsService {
	return &PaymentsService{
		HttpService: HttpService{
			BaseUrl: baseUrl,
		},
		repo:       repo,
		maxResults: maxResults,
	}
}

func (s *PaymentsService) Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Get("/payments", s.List)
	router.Get("/payments/{id}", s.Fetch)
	router.Post("/payments", s.Create)
	router.Put("/payments/{id}", s.Update)
	router.Delete("/payments/{id}", s.Delete)
	return router
}

func (s *PaymentsService) List(w http.ResponseWriter, r *http.Request) {
	from := IntFromStringOrDefault(r.URL.Query().Get("from"), 0)
	to := IntFromStringOrDefault(r.URL.Query().Get("to"), s.maxResults)

	limit := to - from

	if limit <= 0 {
		HandleHttpError(w, r, http.StatusBadRequest, fmt.Errorf("Invalid from (%v) or to (%v) query params", from, to))
		return
	}

	if limit > s.maxResults {
		limit = s.maxResults
	}

	repoItems, err := s.repo.List(from, limit)
	if err != nil {
		HandleHttpError(w, r, http.StatusInternalServerError, err)
		return
	}

	payments, err := NewPaymentsFromRepoItems(repoItems)
	if err != nil {
		HandleHttpError(w, r, http.StatusInternalServerError, err)
		return
	}

	links := make(Links)
	links["self"] = s.UrlFor(fmt.Sprintf(paymentsLinkPattern, from, to))
	links["next"] = s.UrlFor(fmt.Sprintf(paymentsLinkPattern, to, to+limit))

	if from >= limit {
		links["prev"] = s.UrlFor(fmt.Sprintf(paymentsLinkPattern, from-limit, from))
	}

	RenderJSON(w, r, http.StatusOK, &PaymentsResponse{
		Data:  payments,
		Links: links,
	})

}

func (s *PaymentsService) Fetch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	found, err := s.repo.Fetch(&RepoItem{Id: id})
	if err != nil {
		// Look for not found errors
		if s.repo.IsNotFound(err) {
			HandleHttpError(w, r, http.StatusNotFound, err)
		} else {
			HandleHttpError(w, r, http.StatusInternalServerError, err)
		}
		return
	}

	p, err := NewPaymentFromRepoItem(found)
	if err != nil {
		HandleHttpError(w, r, http.StatusInternalServerError, err)
		return
	}

	links := make(Links)
	links["self"] = s.UrlFor(fmt.Sprintf(paymentLinkPattern, id))

	RenderJSON(w, r, http.StatusOK, &PaymentResponse{
		Data:  p,
		Links: links,
	})
}

func (s *PaymentsService) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	versionQP := strings.TrimSpace(r.URL.Query().Get("version"))
	version, err := strconv.Atoi(versionQP)
	if err != nil {
		HandleHttpError(w, r, http.StatusBadRequest, err)
		return
	}

	_, err = s.repo.Fetch(&RepoItem{Id: id})
	if err != nil {
		// Look for not found errors
		if s.repo.IsNotFound(err) {
			HandleHttpError(w, r, http.StatusNotFound, err)
		} else {
			HandleHttpError(w, r, http.StatusInternalServerError, err)
		}
		return
	}

	err = s.repo.Delete(&RepoItem{Id: id, Version: version})
	if err != nil {
		// Look for not found errors again
		errorCode := http.StatusInternalServerError
		if s.repo.IsNotFound(err) || s.repo.IsConflict(err) {
			// The item was deleted in between, from a different goroutine
			// or was updated and the version increased. We treat both
			// cases as a concurren modification, that we
			// translate into a 409 Conflict
			errorCode = http.StatusConflict
		}
		HandleHttpError(w, r, errorCode, err)
		return
	}

	RenderNoContent(w, r)
}

func (s *PaymentsService) Create(w http.ResponseWriter, r *http.Request) {

	p, err := decodePayment(r)
	if err != nil {
		// Something wrong with the JSON
		// Translate this into a 400 Bad Request and finish
		// the request
		HandleHttpError(w, r, http.StatusBadRequest, err)
		return
	}

	log.Info("payment: %v", p)

	err = p.Validate()
	if err != nil {
		HandleHttpError(w, r, http.StatusBadRequest, err)
		return
	}

	repoItem, err := p.ToRepoItem()
	if err != nil {
		// check for conflicts
		// or internal errors
		HandleHttpError(w, r, http.StatusInternalServerError, err)
		return
	}

	createdItem, err := s.repo.Create(repoItem)
	if err != nil {
		if s.repo.IsConflict(err) {
			// We have a conflict, so return the appropiate
			// status code
			HandleHttpError(w, r, http.StatusConflict, err)
		} else {
			HandleHttpError(w, r, http.StatusInternalServerError, err)
		}

		return
	}

	p, err = NewPaymentFromRepoItem(createdItem)
	if err != nil {
		HandleHttpError(w, r, http.StatusInternalServerError, err)
		return
	}

	links := make(Links)
	links["self"] = s.UrlFor(fmt.Sprintf(paymentLinkPattern, p.Id))

	RenderJSON(w, r, http.StatusCreated, &PaymentResponse{
		Data:  p,
		Links: links,
	})
}

func (s *PaymentsService) Update(w http.ResponseWriter, r *http.Request) {

	p, err := decodePayment(r)
	if err != nil {
		// Something is wrong with the JSON
		// Translate this into a 400 Bad Request and finish
		// the request
		HandleHttpError(w, r, http.StatusBadRequest, err)
		return
	}

	err = p.Validate()
	if err != nil {
		HandleHttpError(w, r, http.StatusBadRequest, err)
		return
	}

	id := chi.URLParam(r, "id")

	if p.Id != "" && id != p.Id {
		HandleHttpError(w, r, http.StatusBadRequest, err)
		return
	}

	_, err = s.repo.Fetch(&RepoItem{Id: id})
	if err != nil {
		// Look for not found errors
		if s.repo.IsNotFound(err) {
			HandleHttpError(w, r, http.StatusNotFound, err)
		} else {
			HandleHttpError(w, r, http.StatusInternalServerError, err)
		}
		return
	}

	repoItem, err := p.ToRepoItem()
	if err != nil {
		HandleHttpError(w, r, http.StatusInternalServerError, err)
		return
	}

	updatedItem, err := s.repo.Update(repoItem)
	if err != nil {
		if s.repo.IsConflict(err) {
			// We have a conflict, so return the appropiate
			// status code
			HandleHttpError(w, r, http.StatusConflict, err)
		} else {
			HandleHttpError(w, r, http.StatusInternalServerError, err)
		}
		return
	}

	p, err = NewPaymentFromRepoItem(updatedItem)
	if err != nil {
		HandleHttpError(w, r, http.StatusBadRequest, err)
		return
	}

	links := make(Links)
	links["self"] = s.UrlFor(fmt.Sprintf(paymentLinkPattern, id))

	RenderJSON(w, r, http.StatusOK, &PaymentResponse{
		Data:  p,
		Links: links,
	})
}

func decodePayment(r *http.Request) (*Payment, error) {
	decoder := json.NewDecoder(r.Body)
	var pr PaymentRequest
	err := decoder.Decode(&pr)
	return pr.Payment, err
}
