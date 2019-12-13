package controllers

import (
	"FlankiRest/services"
	u "FlankiRest/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type ImageController struct {
	logger *logrus.Logger
}

func NewImageController(logg *logrus.Logger) *ImageController {
	return &ImageController{logg}
}

func (controller *ImageController) GetImageById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imgID, _ := strconv.Atoi(vars["id"])
	img, contentType, err := services.GetImageService().GetUserImageById(imgID)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}

	w.Header().Set("Content-Type", contentType)
	_, err = w.Write(img)
	if err != nil {
		controller.logger.Error("Encountered error while sending image: ", err.Error())
	}
	return
}

func (controller *ImageController) GetOwnerImage(w http.ResponseWriter, r *http.Request) {
	id, err := u.GetUserIdFromContext(r.Context())
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}

	img, contentType, err := services.GetImageService().GetUserImageById(int(id))
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, err = w.Write(img)
	if err != nil {
		controller.logger.Error("Encountered error while sending image: ", err.Error())
	}
	return
}

func (controller *ImageController) UploadImage(w http.ResponseWriter, r *http.Request) {
	id, err := u.GetUserIdFromContext(r.Context())
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	contentType := r.Header.Get("Content-Type")
	err = services.GetImageService().UploadImageById(int(id), contentType, r.Body)
	if err != nil {
		u.ApiErrorResponse(w, err)
		return
	}
	u.SimpleRespond(w, u.TextMessage("Image has been uploaded"))
	return
}
