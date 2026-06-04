package giftmessage

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/httputil"
	"github.com/ferjunior7/parasempre/backend/internal/middleware"
)

const (
	maxMessageBodyBytes = 55 << 20
	multipartInMemory   = 1 << 20
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("autenticação obrigatória"))
		return
	}

	txID, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("id de transação inválido", err))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxMessageBodyBytes)

	if err := r.ParseMultipartForm(multipartInMemory); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httputil.WriteError(w, r, apperror.Validation("arquivo excede o tamanho máximo permitido"))
			return
		}
		httputil.WriteError(w, r, apperror.Validation("formato multipart inválido"))
		return
	}

	input := CreateInput{
		AuthorName: r.FormValue("author_name"),
		Content:    r.FormValue("content"),
	}

	var media *Media
	file, header, err := r.FormFile("media")
	switch err {
	case nil:
		defer file.Close()
		if header.Size <= 0 {
			httputil.WriteError(w, r, apperror.Validation("arquivo de mídia vazio"))
			return
		}
		media = &Media{
			DeclaredMime: header.Header.Get("Content-Type"),
			Size:         header.Size,
			Reader:       file,
		}
	case http.ErrMissingFile:
	default:
		httputil.WriteError(w, r, apperror.Validation("falha ao ler arquivo de mídia"))
		return
	}

	msg, err := h.svc.Create(r.Context(), txID, userID, input, media)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("falha ao criar mensagem", err))
		return
	}

	pub, _ := h.svc.signSingle(r.Context(), *msg)
	httputil.WriteJSON(w, http.StatusCreated, pub)
}

func (h *Handler) HandleGetMine(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("autenticação obrigatória"))
		return
	}
	txID, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("id de transação inválido", err))
		return
	}
	msg, err := h.svc.GetMine(r.Context(), txID, userID)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	if msg == nil {
		httputil.WriteError(w, r, apperror.NotFound("mensagem não encontrada"))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, msg)
}

func (h *Handler) HandleListByGift(w http.ResponseWriter, r *http.Request) {
	giftID, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("id de presente inválido", err))
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	resp, err := h.svc.ListByGift(r.Context(), giftID, page, limit)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleAdminList(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	resp, err := h.svc.ListAll(r.Context(), page, limit)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleAdminDelete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("autenticação obrigatória"))
		return
	}
	id, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("id de mensagem inválido", err))
		return
	}
	if err := h.svc.Remove(r.Context(), id, userID); err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
