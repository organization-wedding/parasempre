package payment

import (
	"context"

	"github.com/ferjunior7/parasempre/backend/internal/giftmessage"
)

type MessageTxFinder struct {
	repo Repository
}

func NewMessageTxFinder(repo Repository) *MessageTxFinder {
	return &MessageTxFinder{repo: repo}
}

func (a *MessageTxFinder) GetByID(ctx context.Context, id int64) (*giftmessage.TransactionSnapshot, error) {
	tx, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &giftmessage.TransactionSnapshot{
		ID:     tx.ID,
		GiftID: tx.GiftID,
		UserID: tx.UserID,
		Status: tx.Status,
	}, nil
}
