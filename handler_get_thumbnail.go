package main

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerThumbnailGet(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid video ID", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Thumbnail not found", nil)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get video", err)
		}
		return
	}

	/*
	* extracting the required data from the data schema
	* https://developer.mozilla.org/en-US/docs/Web/URI/Reference/Schemes/data
	 */
	const prefix = "data:"
	comma := strings.IndexByte(*video.ThumbnailURL, ',')
	if comma < 0 || !strings.HasPrefix(*video.ThumbnailURL, prefix) {
		respondWithError(w, http.StatusInternalServerError, "Invalid thumbnail data", nil)
		return
	}

	meta := (*video.ThumbnailURL)[len(prefix):comma]
	mimeEnd := strings.IndexByte(meta, ';')

	contentType := meta[:mimeEnd]

	w.Header().Set("Content-Type", contentType)

	data := (*video.ThumbnailURL)[comma+1:]

	decodedImage, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode image", err)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(decodedImage)))

	_, err = w.Write(decodedImage)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error writing response", err)
		return
	}
}
