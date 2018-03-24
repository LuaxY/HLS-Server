package middleware

import (
    "net/http"

    "HLS-Server/src/errors"

    "github.com/sirupsen/logrus"
)

func PanicRecover(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {

                var message interface{}

                fields := logrus.Fields{
                    "remote_addr": r.RemoteAddr,
                    "method":      r.Method,
                    "request_uri": r.RequestURI,
                    "referer":     r.Referer(),
                    "user_agent":  r.UserAgent(),
                }

                switch err.(type) {
                case errors.Error:
                    errInfo := err.(errors.Error)
                    message = errInfo.Error
                    for k, v := range errInfo.Fields {
                        fields[k] = v
                    }
                default:
                    message = err
                }

                log.WithFields(fields).Error(message)

                w.WriteHeader(http.StatusInternalServerError)
                w.Write([]byte("Error"))
            }
        }()

        next.ServeHTTP(w, r)
    })
}
