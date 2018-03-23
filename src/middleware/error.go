package middleware

import (
    "net/http"

    "HLS-Server/src/errors"
)

func PanicRecover(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                switch err.(type) {
                case errors.Error:
                    errInfo := err.(errors.Error)
                    log.WithFields(errInfo.Fields).Error(errInfo.Error)
                default:
                    log.Error(err)
                }

                w.WriteHeader(http.StatusInternalServerError)
                w.Write([]byte("Error"))
            }
        }()

        next.ServeHTTP(w, r)
    })
}
