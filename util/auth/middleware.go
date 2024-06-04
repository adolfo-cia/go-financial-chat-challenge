package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/tomiok/webh"
)

const (
	authorizationHeaderKey     = "authorization"
	authorizationTypeBearer    = "bearer"
	AuthorizationPayloadCtxKey = "authPayload"
)

func Middleware(tokenMaker *PasetoMaker) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorizationHeader := r.Header.Get(authorizationHeaderKey)
			if len(authorizationHeader) == 0 {
				webh.ResponseErr(http.StatusUnauthorized, w, "missing authentication header", nil)
				return
			}

			fields := strings.Fields(authorizationHeader)
			if len(fields) < 2 {
				webh.ResponseErr(http.StatusUnauthorized, w, "invalid authorization header format", nil)
				return
			}

			authType := strings.ToLower(fields[0])
			if authType != authorizationTypeBearer {
				webh.ResponseErr(http.StatusUnauthorized, w, "unsupported authorization type: "+authType, nil)
				return
			}

			accessToken := fields[1]

			payload, err := tokenMaker.VerifyToken(accessToken)
			if err != nil {
				webh.ResponseErr(http.StatusUnauthorized, w, err.Error(), nil)
				return
			}

			ctx := context.WithValue(r.Context(), AuthorizationPayloadCtxKey, payload)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
