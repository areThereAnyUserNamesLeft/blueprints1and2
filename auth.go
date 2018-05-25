package main

import (
	"net/http"
)

type authHander struct {
	next http.Handler
}

func (h *authHander) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("auth")
	if err == http.ErrNoCookie {
		// no auth
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	if err != nil {
		panic(err.Error())
		return
	}
	// works!
	h.next.ServeHTTP(w, r)

}
func MustAuth(handler http.Handler) http.Handler {
	return &authHander{next: handler}
}
