package main

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/epaitoo/ephermalbridge/internal/data"
	"github.com/epaitoo/ephermalbridge/internal/upload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const maxUploadSize = 10 << 20 // 10MB

func (app *application) uploadFilesHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "file(s) too large, max 10MB")
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		app.errorResponse(w, r, http.StatusBadRequest, "no files provided, use form key 'files'")
		return
	}

	// Build FileInput slice from multipart headers
	var fileInputs []upload.FileInput
	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			app.logger.Error("failed to open uploaded file", "filename", fh.Filename, "error", err)
			continue
		}

		contentType := fh.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		fileInputs = append(fileInputs, upload.FileInput{
			Reader:      f.(io.ReadSeeker),
			Filename:    fh.Filename,
			ContentType: contentType,
			Size:        fh.Size,
		})
	}

	if len(fileInputs) == 0 {
		app.errorResponse(w, r, http.StatusBadRequest, "failed to open uploaded files")
		return
	}

	// Respond immediately — uploads happen in the background
	app.writeJSON(w, http.StatusAccepted, envelope{
		"message": "upload started",
		"count":   len(fileInputs),
	}, nil)

	go app.coordinator.ProcessUploads(fileInputs)
}

func (app *application) getAllFilesHandler(w http.ResponseWriter, r *http.Request) {
	files, err := app.models.Files.GetAllFiles()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	var response []data.FileResponse
	for _, f := range *files {
		response = append(response, f.ToResponse())
	}

	app.writeJSON(w, http.StatusOK, envelope{"files": response}, nil)
}

func (app *application) getFileHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid file ID")
		return
	}

	file, err := app.models.Files.Get(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	response := file.ToResponse()
	app.writeJSON(w, http.StatusOK, envelope{"file": response}, nil)
}

func (app *application) deleteFileHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid file ID")
		return
	}

	err = app.models.Files.Delete(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"message": "file deleted"}, nil)
}

const presignedURLExpiry = 10 * time.Minute

func (app *application) downloadFileHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid file ID")
		return
	}

	file, err := app.models.Files.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if file.ExpiresAt != nil && file.ExpiresAt.Before(time.Now()) {
		app.errorResponse(w, r, http.StatusGone, "file has expired")
		return
	}

	downloadURL, err := app.storage.GenerateDownloadURL(r.Context(), file.Bucket, file.ObjectKey, presignedURLExpiry)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if file.DownloadedAt == nil {
		expiresAt := time.Now().Add(24 * time.Hour)
		err = app.models.Files.MarkDownloaded(id, expiresAt)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	app.writeJSON(w, http.StatusOK, envelope{
		"download_url":       downloadURL,
		"expires_in_seconds": int(presignedURLExpiry.Seconds()),
	}, nil)
}

func (app *application) deleteExpiredFilesHandler(w http.ResponseWriter, r *http.Request) {
	app.coordinator.ProcessDeleteExpiredFiles()

	app.writeJSON(w, http.StatusOK, envelope{"message": "expired files cleanup completed"}, nil)
}
