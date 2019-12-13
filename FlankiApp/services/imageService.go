package services

import (
	"FlankiRest/config"
	"FlankiRest/errors"
	"FlankiRest/logger"
	"FlankiRest/utils"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

type ImageService struct {
	Cfg *config.ImageServerConfig
}

func (service *ImageService) GetUserImageById(id int) (img []byte, contentType string, err error) {

	url := service.Cfg.Domain + ":" + service.Cfg.Port + "/images/" + strconv.Itoa(id)
	resp, err := http.Get(url)
	if err != nil {
		err = errors.New("ImageService error: " + err.Error(), 500)
		return
	}

	defer resp.Body.Close()
	img, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		err = errors.New("ImageService error: " + err.Error(), 500)
		return
	}
	if resp.StatusCode == 404 {
		err = errors.New("Image not found", 404)
		return
	}
	if resp.StatusCode != 200 {
		err = errors.New("Got an error while trying to find image", resp.StatusCode)
		return
	}
	contentType = resp.Header.Get("Content-Type")

	return
}

func (service *ImageService) UploadImageById(id int, contentType string, img io.Reader) error {
	url := service.Cfg.Domain + ":" + service.Cfg.Port + "/upload/" + strconv.Itoa(id)
	resp, err := http.Post(url,contentType, img)

	if err != nil {
		return errors.New("Error while uploading image: " + err.Error(), 500)
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		errMsg := utils.DecodeErrMessage(body)
		return errors.New(errMsg.Message, resp.StatusCode)
	}
	return nil
}

var imgServiceInstance *ImageService
var imgServiceOnce 		sync.Once

func GetImageService() *ImageService {
	imgServiceOnce.Do(func(){
		cfg := config.GetImageServerConfig()
		if cfg.Domain == "" || cfg.Port == "" {
			logger.GetGlobalLogger().Warn("Lacking Image Server configuration environmental variables")
		}
		imgServiceInstance = &ImageService{cfg}
	})
	return imgServiceInstance
}