/*
 * Copyright (c) 2019. Pandranki Global Private Limited
 */

package resolvers

import (
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/gofrs/uuid"
	cloudstorage "github.com/tribehq/platform/lib/storage"
	"github.com/tribehq/platform/models"
	"io"
	"net/http"
	"path"
)

//SingleUpload returns a file upload
func (r *mutationResolver) SingleUpload(ctx context.Context, file graphql.Upload) (*models.File, error) {
	return uploadToGCS(file)
}

//SingleUploadWithPayload gives a file upload with payload
func (r *mutationResolver) SingleUploadWithPayload(ctx context.Context, req models.UploadFile) (*models.File, error) {
	return uploadToGCS(req.File)
}

//MultipleUpload gives multiple file uploads
func (r *mutationResolver) MultipleUpload(ctx context.Context, files []*graphql.Upload) ([]*models.File, error) {
	if len(files) == 0 {
		return nil, errors.New("empty file upload list")
	}
	var resp []*models.File
	for i := range files {

		uploadedFile, err := uploadToGCS(*files[i])
		if err != nil {
			return nil, err
		}

		resp = append(resp, uploadedFile)
	}
	return resp, nil
}

//MultipleUploadWithPayload gives multiple files with payload
func (r *mutationResolver) MultipleUploadWithPayload(ctx context.Context, req []*models.UploadFile) ([]*models.File, error) {
	if len(req) == 0 {
		return nil, errors.New("empty file upload list")
	}
	var resp []*models.File
	for i := range req {

		uploadedFile, err := uploadToGCS(req[i].File)
		if err != nil {
			return nil, err
		}

		resp = append(resp, uploadedFile)
	}
	return resp, nil
}

//GetFileContentType retrieves the content type of files
func GetFileContentType(out io.Reader) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

func uploadToGCS(file graphql.Upload) (*models.File, error) {
	if cloudstorage.StorageBucket == nil {
		return nil, errors.New("internal server error, storage bucket is missing")
	}

	// random filename, retaining existing extension.
	name := uuid.Must(uuid.NewV4()).String() + path.Ext(file.Filename)

	w := cloudstorage.StorageBucket.Object(name).NewWriter(context.Background())

	// Warning: storage.AllUsers gives public read access to anyone.
	w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	contentType, err := GetFileContentType(file.File)
	if err != nil {
		return nil, err //invalid content type
	}

	w.ContentType = contentType
	// Entries are immutable, be aggressive about caching (1 day).
	w.CacheControl = "public, max-age=86400"

	if _, err := io.Copy(w, file.File); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	const publicURL = "https://storage.googleapis.com/%s/%s"

	uploadedFile := &models.File{
		ID:   name,
		Name: file.Filename,
		URL:  fmt.Sprintf(publicURL, cloudstorage.StorageBucketName, name),
	}

	return uploadedFile, nil
}
