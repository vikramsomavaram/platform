/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Translation represents a translation.
type Translation struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt    *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy    string             `json:"createdBy" bson:"createdBy"`
	Key          string             `json:"key" bson:"key"`
	Translations map[string]string  `json:"translations" bson:"translations"`
}

// Language represents a i18n language.
type Language struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	CreatedBy string             `json:"createdBy" bson:"createdBy"`
	Name      string             `json:"name"`
	Code      string             `json:"code"`
}
