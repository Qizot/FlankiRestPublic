package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

// Handler wrapper intentionally for http.FileServer to serve single page web app
// if file was not found it means that we need to redirect to index.html
// and thats what we do here
func IndexWhenNotFound(homeDirectory string, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(homeDirectory + r.URL.Path); os.IsNotExist(err) {
			http.ServeFile(w, r, homeDirectory + "/index.html")
		} else {
			h.ServeHTTP(w,r)
		}
	}
}


type SSLConfig struct {
	Cert    string
	Privkey string
}

// Temporary solution, I mount frontend container to specific folder but
// better way to go would be env variables with paths to ssl certs
func LoadSSLConfig() (*SSLConfig, error) {
	ssl := &SSLConfig{}
	ssl.Cert = "/ssl_certs/cert.pem"
	ssl.Privkey = "/ssl_certs/privkey.pem"
	return ssl, nil
}

func main() {
	// this line is is only needed to be predefined when running frontend on localhost, otherwise remote server should
	// set env variable FRONTEND_BUILD_PATH which will point to folder with website statics and index.html
	buildFolder := `C:\Users\Jakub\go\src\FlankiRest\FrontendServer\frontend\flaneczki\build`

	if val, ok := os.LookupEnv("FRONTEND_BUILD_PATH"); ok {
		buildFolder = val
	}

	router := mux.NewRouter()

	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(buildFolder + "/static")))
	router.PathPrefix("/static").Handler(staticHandler)

	fs := http.FileServer(http.Dir(buildFolder))
	router.PathPrefix("/").Handler(IndexWhenNotFound(buildFolder, fs))

	log.Println("Starting frontend server")
	if val, ok := os.LookupEnv("ENABLE_SSL"); ok && val == "true" {
		log.Println("Using SSL")
		sslConfig, _ := LoadSSLConfig()

		// https server
		log.Fatal(http.ListenAndServeTLS(":443", sslConfig.Cert, sslConfig.Privkey, router))
	} else {
		log.Fatal(http.ListenAndServe(":80", router))
	}

}
