package main

import (
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

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to parse form file", err)
		return
	}

	mediaType := header.Header.Get("Content-Type")
	println("media type", mediaType)

	byteFile, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to read the file", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "cannot get video metadata", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "not the video's owner", nil)
		return
	}

	thumbnail := thumbnail{
		data:      byteFile,
		mediaType: mediaType,
	}
	videoThumbnails[videoID] = thumbnail

	thumbnailUrl := fmt.Sprintf("http://localhost:%s/api/thumbnails/%s", cfg.port, videoID)

	video.ThumbnailURL = &thumbnailUrl

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		delete(videoThumbnails, videoID)
		respondWithError(w, http.StatusInternalServerError, "error when update thumbnail url", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
