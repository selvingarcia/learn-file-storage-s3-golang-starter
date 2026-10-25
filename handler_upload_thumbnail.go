package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
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
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}

	var thumbnails thumbnail

	b, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to read the file", err)
		return
	}

	thumbnails.data = b
	thumbnails.mediaType = header.Header.Get("Content-Type")

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			respondWithError(w, http.StatusNotFound, err.Error(), err)
			return
		default:
			respondWithError(w,http.StatusInternalServerError, err.Error(), err)
			return
		}
	}

	if video.CreateVideoParams.UserID != userID{
		respondWithError(w, http.StatusForbidden, "Not Authorized", nil)
		return
	}


	// thumbnailUrl := fmt.Sprintf("http://localhost:%s/api/thumbnails/%s", cfg.port, videoID)

	//TODO: change later
	// video.ThumbnailURL = &thumbnailUrl
	str := base64.StdEncoding.EncodeToString(thumbnails.data)

	videoFormat := fmt.Sprintf("data:%s;base64,%s", thumbnails.mediaType, str)

	video.ThumbnailURL = &videoFormat
	video.UpdatedAt = time.Now()

	if err:= cfg.db.UpdateVideo(video); err!=nil{
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}


	respondWithJSON(w, http.StatusOK, video)
}
