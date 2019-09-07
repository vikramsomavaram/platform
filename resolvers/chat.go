/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"context"
	"github.com/tribehq/platform/models"
)

type chatResolver struct{ *Resolver }

func (r *chatResolver) CreatedBy(ctx context.Context, obj *models.Chat) (string, error) {
	return obj.CreatedBy.Hex(), nil
}

type chatMessageResolver struct{ *Resolver }

func (chatMessageResolver) Type(ctx context.Context, obj *models.ChatMessage) (*models.ChatMessageType, error) {
	return &obj.Type, nil
}
