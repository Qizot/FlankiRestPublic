package app

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type RequestTimer struct {
	logger *logrus.Logger
	measure bool
}

func NewRequestTimer(logger *logrus.Logger, measure bool) *RequestTimer {
	return &RequestTimer{logger: logger, measure: measure}
}

func (timer *RequestTimer) RequestTimeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !timer.measure {
			next.ServeHTTP(w,r)
			return
		}
		now := time.Now()
		next.ServeHTTP(w,r)
		elapsed := time.Since(now)
		timer.logger.Info("Request served in ", elapsed)
	})
}
