package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to parse form file", err)
		return
	}

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "invalid media type", err)
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

	extension := strings.Split(mediaType, "/")[1]
	thumbnailId, err := generateRandomFilename(extension)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to generate thumbnail ID", err)
		return
	}
	newThumbnailPath := filepath.Join(cfg.assetsRoot, thumbnailId)

	fileHandle, err := os.Create(newThumbnailPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to create thumbnail file", err)
		return
	}
	defer fileHandle.Close()

	_, err = io.Copy(fileHandle, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to copy file", err)
		return
	}

	thumbnailUrl := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, thumbnailId)

	video.ThumbnailURL = &thumbnailUrl

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error when update thumbnail url", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}

func generateRandomFilename(mediaType string) (string, error) {
	_, err := rand.Read(make([]byte, 32))
	if err != nil {
		return "", err
	}
	thumbnailId := base64.RawURLEncoding.EncodeToString(make([]byte, 32))
	return thumbnailId + "." + mediaType, nil
}
