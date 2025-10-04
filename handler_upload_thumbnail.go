package main

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20 // 10 MB

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse multipart form", err)
		return
	}

	thumbnailFile, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get thumbnail from form", err)
		return
	}
	defer thumbnailFile.Close()

	contentType := header.Header.Get("Content-Type")

	imageData, err := io.ReadAll(thumbnailFile)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't read the image data", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Video not found", err)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get video", err)
		}
		return
	}

	// if the authenticated user is not the owner of the video, return a 403
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "You can't upload a thumbnail for this video", nil)
		return
	}

	// saving the image to the database. yes really! but don't worry, we'll change this later
	encodedImage := base64.StdEncoding.EncodeToString(imageData)

	/*
	*
	* creating a data URL for the image
	* for what the date URL check https://developer.mozilla.org/en-US/docs/Web/URI/Reference/Schemes/data
	*
	 */
	thumbnailURL := fmt.Sprintf("data:%s;base64,%s", contentType, encodedImage)

	// thumbnailURL := fmt.Sprintf("http://localhost:%s/api/thumbnails/%s", cfg.port, videoID)
	video.ThumbnailURL = &thumbnailURL

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video with thumbnail URL", err)
	}

	respondWithJSON(w, http.StatusOK, video)
}
