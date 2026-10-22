package newsletter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

const maxBodyBytes = 4 << 10

// Handler exposes the Service over HTTP.
type Handler struct {
	svc     *Service
	siteURL string
	log     *slog.Logger
}

// NewHandler returns a Handler that redirects browser flows to pages under siteURL.
func NewHandler(svc *Service, siteURL string, log *slog.Logger) *Handler {
	return &Handler{svc: svc, siteURL: siteURL, log: log}
}

// Register adds the newsletter routes to mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/subscribe", h.subscribe)
	mux.HandleFunc("GET /v1/confirm", h.confirm)
	mux.HandleFunc("GET /v1/unsubscribe", h.unsubscribe)
}

type subscribeRequest struct {
	Email   string `json:"email"`
	List    string `json:"list"`
	Website string `json:"website"` // honeypot: real users leave it empty
}

func (h *Handler) subscribe(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	var req subscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if req.Website != "" {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	err := h.svc.Subscribe(r.Context(), req.Email, req.List)
	switch {
	case errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrUnknownList):
		writeError(w, http.StatusBadRequest, err.Error())
	case err != nil:
		h.log.ErrorContext(r.Context(), "subscribe failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *Handler) confirm(w http.ResponseWriter, r *http.Request) {
	h.finish(w, r, h.svc.Confirm(r.Context(), r.URL.Query().Get("token")), "/newsletter/confirmed/")
}

func (h *Handler) unsubscribe(w http.ResponseWriter, r *http.Request) {
	h.finish(w, r, h.svc.Unsubscribe(r.Context(), r.URL.Query().Get("token")), "/newsletter/unsubscribed/")
}

func (h *Handler) finish(w http.ResponseWriter, r *http.Request, err error, okPath string) {
	path := okPath
	switch {
	case errors.Is(err, ErrInvalidToken):
		path = "/newsletter/invalid/"
	case err != nil:
		h.log.ErrorContext(r.Context(), "request failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, h.siteURL+path, http.StatusSeeOther)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
