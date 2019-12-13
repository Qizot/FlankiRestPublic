package app

import (
	"FlankiRest/config"
	"FlankiRest/errors"
	"FlankiRest/logger"
	u "FlankiRest/utils"
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type AuthTokenResponse struct {
	Expires  uint    `json:"expires_in"`
	ClientID string `json:"client_id"`
	UserID   uint `json:"user_id"`
}



func Oauth2Authentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		notAuth := []string{"/user/create", "/user/login", "players/[0-9]+", "/players/ranking", "/players/ranking/[0-9]+", "/images/[0-9]+", "/remember_password", "/reset_password" } //List of endpoints that doesn't require auth
		requestPath := r.URL.Path //current request path

		//check if request does not need authentication, serve the request if it doesn't need it
		for _, pattern := range notAuth {

			if match, _ := regexp.MatchString(config.API_PREFIX + pattern, requestPath); match {
				next.ServeHTTP(w, r)
				return
			}
		}

		authcfg := config.GetAuthServerConfig()
		logEntry := logger.GetGlobalLogger().WithField("prefix", "[OAUTH MIDDLEWARE]")

		url := authcfg.Domain + ":" + authcfg.Port + "/authorize"
		request, err := http.NewRequest("POST", url, bytes.NewBuffer(nil))
		request.Header = r.Header
		client := &http.Client{}
		resp, err := client.Do(request)
		if err != nil {
			logEntry.Error("Couldn't connect with authorization server")
			u.ApiErrorResponse(w, errors.New("Couldn't reach authorization server", 500))
			return
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)




		if resp.StatusCode == 500 {
			u.ApiErrorResponse(w, errors.New("Authorization server internal error", 500))
			logEntry.Error("Authorization server: ", strings.TrimSpace(string(body)))
			return
		} else if resp.StatusCode != 200 {
			u.ApiErrorResponse(w, errors.InvalidToken)
			logEntry.Warn("Unauthorized user: ", strings.TrimSuffix(string(body),"\n"))
			return
		}

		tokenResp := &AuthTokenResponse{}
		err = json.NewDecoder(bytes.NewReader(body)).Decode(tokenResp)
		if err != nil {
			logEntry.Error(err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), "user", tokenResp.UserID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})

}



