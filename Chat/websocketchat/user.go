package websocketchat

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type User struct {
	ID uint `json:"-"`
	Nickname string `json:"nickname"`
	Connected time.Time `json:"-"`
	Token string `json:"token"`
}

// interface validating user information
// validating token which will be used for nickname retrieving
type UserChecker interface {

	// checks user authorization with given access token and saves his id fetched from authorization server
	Authorize(*User) error

	// obtains user information e.g. nickname
	FetchUserInformation(*User) error
}

type FlankiChecker struct {
	authEndpoint string
	apiEndpoint  string
}

var (
	GetFlankiChecker  = FlankiChecker{}
	httpClient = &http.Client{}
)

func (checker *FlankiChecker) SetEndpoints(api string, auth string) {
	checker.apiEndpoint = api
	checker.authEndpoint = auth
}

type TokenResponse struct {
	Expires  uint    `json:"expires_in"`
	ClientID string `json:"client_id"`
	UserID   uint `json:"user_id"`
}

func (checker *FlankiChecker) Authorize(user *User) error {
	url := checker.authEndpoint + "/authorize"
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(nil))
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", user.Token)
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("unauthorized user")
	}

	tokenInfo := TokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(&tokenInfo)
	if err != nil {
		return err
	}
	user.ID = tokenInfo.UserID
	return nil
}

func (checker *FlankiChecker) FetchUserInformation(user *User) error {
	url := checker.apiEndpoint + "/user/me"

	request, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", user.Token)

	resp, err := httpClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}

	if nickname, ok := data["account"].(map[string]interface{})["nickname"].(string); !ok {
		return errors.New("couldn't get nickname")
	} else {
		user.Nickname = nickname
	}
	return nil
}

func init() {

	if production := os.Getenv("PRODUCTION"); production != "true" {
		return
	}


	localCertFile := "/ssl_certs/cert.pem"
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read in the cert file
	certs, err := ioutil.ReadFile(localCertFile)
	if err != nil {
		Logger().Fatal("Failed to append %q to RootCAs: %v", localCertFile, err)
	}

	// Append our cert to the system pool
	if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
		Logger().Info("No certs appended, using system certs only")
	}

	// Trust the augmented cert pool in our client
	config := &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            rootCAs,
	}
	tr := &http.Transport{TLSClientConfig: config}
	httpClient = &http.Client{Transport: tr}
}