package giftmessage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/validate"
)

const (
	approvedStatus = "approved"

	defaultListLimit  = 10
	maxListLimit      = 50
	defaultAdminLimit = 20
	maxAdminLimit     = 100
)

type TransactionFinder interface {
	GetByID(ctx context.Context, id int64) (*TransactionSnapshot, error)
}
type AuditLogger interface {
	LogAction(ctx context.Context, userID int64, action string, details map[string]any) error
}

const (
	auditMessageCreated = "giftmessage.created"
	auditMessageRemoved = "giftmessage.removed"
)

type Media struct {
	DeclaredMime string
	Size         int64
	Reader       io.Reader
}

type Service struct {
	repo    TxAwareRepository
	txns    TransactionFinder
	storage Storage
	audit   AuditLogger
	ttl     time.Duration
}

func NewService(repo TxAwareRepository, txns TransactionFinder, storage Storage, audit AuditLogger, ttl time.Duration) *Service {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return &Service{repo: repo, txns: txns, storage: storage, audit: audit, ttl: ttl}
}

func (s *Service) recordAudit(ctx context.Context, userID int64, action string, details map[string]any) {
	if s.audit == nil || userID == 0 {
		return
	}
	if err := s.audit.LogAction(ctx, userID, action, details); err != nil {
		slog.Error("giftmessage.service audit failed", "action", action, "user_id", userID, "error", err)
	}
}

func (s *Service) Create(ctx context.Context, txID, requesterUserID int64, in CreateInput, media *Media) (*GiftMessage, error) {
	if requesterUserID == 0 {
		return nil, apperror.Unauthorized("autenticação obrigatória")
	}
	if err := validate.Struct(in); err != nil {
		return nil, err
	}

	tx, err := s.txns.GetByID(ctx, txID)
	if err != nil {
		return nil, err
	}
	if tx.UserID != requesterUserID {
		return nil, apperror.Forbidden("transação não pertence ao usuário")
	}
	if tx.Status != approvedStatus {
		return nil, apperror.Conflict("aguarde a aprovação do pagamento para deixar uma mensagem")
	}

	if existing, err := s.repo.GetByTransactionID(ctx, txID); err == nil && existing != nil {
		return nil, apperror.Conflict("já existe uma mensagem para essa transação")
	} else if err != nil {
		var ae *apperror.AppError
		if !errors.As(err, &ae) || ae.Code != http.StatusNotFound {
			return nil, err
		}
	}

	row := CreateRow{
		GiftTransactionID: tx.ID,
		GiftID:            tx.GiftID,
		UserID:            requesterUserID,
		AuthorName:        in.AuthorName,
		Content:           in.Content,
	}

	var uploadedKey string
	if media != nil {
		if s.storage == nil {
			return nil, apperror.ServiceUnavailable("Mensagens com mídia indisponíveis neste ambiente.")
		}

		mime, peeked, err := resolveMediaMIME(media.DeclaredMime, media.Reader)
		if err != nil {
			return nil, err
		}
		spec, err := ValidateMedia(mime, media.Size)
		if err != nil {
			return nil, err
		}

		key, err := buildObjectKey(tx.ID, spec.ext)
		if err != nil {
			return nil, apperror.Internal("falha ao gerar chave de mídia", err)
		}
		if err := s.storage.Upload(ctx, key, mime, peeked, media.Size); err != nil {
			slog.Error("giftmessage.service create: storage upload failed",
				"key", key, "error", err)
			return nil, apperror.ServiceUnavailable("Não foi possível enviar sua mídia agora. Tente novamente.")
		}

		uploadedKey = key
		size := media.Size
		row.MediaObjectKey = &key
		row.MediaKind = &spec.kind
		row.MediaSizeBytes = &size
		row.MediaMimeType = &mime
	}

	created, err := s.repo.Create(ctx, row)
	if err != nil {
		if uploadedKey != "" {
			if delErr := s.storage.Delete(context.Background(), uploadedKey); delErr != nil {
				slog.Error("giftmessage.service create: orphan delete failed",
					"key", uploadedKey, "error", delErr)
			}
		}
		return nil, err
	}

	mediaKindLog := ""
	if created.MediaKind != nil {
		mediaKindLog = *created.MediaKind
	}
	s.recordAudit(ctx, requesterUserID, auditMessageCreated, map[string]any{
		"message_id": created.ID,
		"tx_id":      created.GiftTransactionID,
		"gift_id":    created.GiftID,
		"media_kind": mediaKindLog,
	})

	slog.Info("giftmessage.service create: done",
		"message_id", created.ID,
		"tx_id", created.GiftTransactionID,
		"gift_id", created.GiftID,
		"user_id", requesterUserID,
		"media_kind", mediaKindLog,
	)

	return created, nil
}

func (s *Service) GetMine(ctx context.Context, txID, requesterUserID int64) (*PublicMessage, error) {
	if requesterUserID == 0 {
		return nil, apperror.Unauthorized("autenticação obrigatória")
	}
	tx, err := s.txns.GetByID(ctx, txID)
	if err != nil {
		return nil, err
	}
	if tx.UserID != requesterUserID {
		return nil, apperror.Forbidden("transação não pertence ao usuário")
	}

	msg, err := s.repo.GetByTransactionID(ctx, txID)
	if err != nil {
		var ae *apperror.AppError
		if errors.As(err, &ae) && ae.Code == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	pub, _ := s.signSingle(ctx, *msg)
	return &pub, nil
}

func (s *Service) signSingle(ctx context.Context, m GiftMessage) (PublicMessage, error) {
	if m.MediaObjectKey == nil || s.storage == nil {
		return toPublic(m, ""), nil
	}
	urls, err := s.storage.SignURLs(ctx, []string{*m.MediaObjectKey}, s.ttl)
	if err != nil {
		slog.Warn("giftmessage.service sign_single: failed", "key", *m.MediaObjectKey, "error", err)
		return toPublic(m, ""), err
	}
	return toPublic(m, urls[*m.MediaObjectKey]), nil
}

func (s *Service) ListByGift(ctx context.Context, giftID int64, page, limit int) (*Paged[PublicMessage], error) {
	page, limit = normalizePaging(page, limit, defaultListLimit, maxListLimit)
	rows, total, err := s.repo.ListByGift(ctx, giftID, limit, (page-1)*limit)
	if err != nil {
		return nil, apperror.WrapIfNotApp("falha ao listar recados", err)
	}

	urls, err := s.signMediaURLs(ctx, rows)
	if err != nil {
		slog.Warn("giftmessage.service list_by_gift: sign urls failed", "error", err)
	}

	data := make([]PublicMessage, len(rows))
	for i, m := range rows {
		data[i] = toPublic(m, urlFor(m, urls))
	}
	return &Paged[PublicMessage]{Data: data, Page: page, Limit: limit, Total: total}, nil
}

func (s *Service) ListAll(ctx context.Context, page, limit int) (*Paged[AdminMessage], error) {
	page, limit = normalizePaging(page, limit, defaultAdminLimit, maxAdminLimit)
	rows, total, err := s.repo.ListAll(ctx, limit, (page-1)*limit)
	if err != nil {
		return nil, apperror.WrapIfNotApp("falha ao listar recados (admin)", err)
	}

	urls, err := s.signMediaURLs(ctx, rows)
	if err != nil {
		slog.Warn("giftmessage.service list_all: sign urls failed", "error", err)
	}

	data := make([]AdminMessage, len(rows))
	for i, m := range rows {
		pub := toPublic(m, urlFor(m, urls))
		data[i] = AdminMessage{
			PublicMessage:     pub,
			UserID:            m.UserID,
			GiftTransactionID: m.GiftTransactionID,
		}
	}
	return &Paged[AdminMessage]{Data: data, Page: page, Limit: limit, Total: total}, nil
}

func (s *Service) Remove(ctx context.Context, id, byUserID int64) error {
	if byUserID == 0 {
		return apperror.Unauthorized("autenticação obrigatória")
	}
	if err := s.repo.SoftDelete(ctx, id, byUserID); err != nil {
		return err
	}
	s.recordAudit(ctx, byUserID, auditMessageRemoved, map[string]any{
		"message_id": id,
	})
	slog.Info("giftmessage.service remove: done", "message_id", id, "by_user_id", byUserID)
	return nil
}

func (s *Service) signMediaURLs(ctx context.Context, rows []GiftMessage) (map[string]string, error) {
	if s.storage == nil {
		return nil, nil
	}
	keys := make([]string, 0, len(rows))
	for _, m := range rows {
		if m.MediaObjectKey != nil {
			keys = append(keys, *m.MediaObjectKey)
		}
	}
	if len(keys) == 0 {
		return nil, nil
	}
	return s.storage.SignURLs(ctx, keys, s.ttl)
}

func urlFor(m GiftMessage, urls map[string]string) string {
	if m.MediaObjectKey == nil || urls == nil {
		return ""
	}
	return urls[*m.MediaObjectKey]
}

func toPublic(m GiftMessage, signedURL string) PublicMessage {
	out := PublicMessage{
		ID:         m.ID,
		GiftID:     m.GiftID,
		AuthorName: m.AuthorName,
		Content:    m.Content,
		MediaKind:  m.MediaKind,
		CreatedAt:  m.CreatedAt,
	}
	if signedURL != "" {
		u := signedURL
		out.MediaURL = &u
	}
	return out
}

func normalizePaging(page, limit, defaultLimit, maxLimit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return page, limit
}

func buildObjectKey(txID int64, ext string) (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("messages/%d/%s%s", txID, hex.EncodeToString(buf), ext), nil
}
