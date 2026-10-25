package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	defer file.Close()

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
	videoType := filepath.Ext(header.Filename)

	destino , err := os.Create(filepath.Join(cfg.assetsRoot,fmt.Sprintf("%s%s",  videoIDString,videoType )))
	if err!=nil{
		respondWithError(w,http.StatusInternalServerError, err.Error(), err)
	}

	defer destino.Close()

	io.Copy(destino, file)
	if err!=nil{
		respondWithError(w,http.StatusInternalServerError, err.Error(), err)
	}

	thumbnailUrl := fmt.Sprintf("http://localhost:%s/assets/%s%s", cfg.port, videoID, videoType)
	
	video.ThumbnailURL = &thumbnailUrl
	video.UpdatedAt = time.Now()

	if err:= cfg.db.UpdateVideo(video); err!=nil{
		respondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}


	respondWithJSON(w, http.StatusOK, video)
}
