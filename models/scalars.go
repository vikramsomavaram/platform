/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package models

import (
	"errors"
	"github.com/99designs/gqlgen/graphql"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"strconv"
	"time"
)

// Lets redefine the base ID type to use object id from mongodb library
func MarshalID(id primitive.ObjectID) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, err := io.WriteString(w, strconv.Quote(id.Hex()))
		if err != nil {
			log.Error(err)
		}
	})
}

// And the same for the unmarshal
func UnmarshalID(v interface{}) (primitive.ObjectID, error) {
	str, ok := v.(string)
	if !ok {
		return primitive.NilObjectID, primitive.ErrInvalidHex
	}
	return primitive.ObjectIDFromHex(str)
}

func MarshalTimestamp(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, err := io.WriteString(w, strconv.Quote(t.Format(time.RFC3339)))
		if err != nil {
			log.Error(err)
		}
	})
}

func UnmarshalTimestamp(v interface{}) (time.Time, error) {
	if tmpStr, ok := v.(string); ok {
		timeStmp, err := time.Parse(time.RFC3339, tmpStr)
		if err != nil {
			log.Error(err)
		}
		return timeStmp, err
	}
	return time.Time{}, errors.New("invalid timestamp")
}
