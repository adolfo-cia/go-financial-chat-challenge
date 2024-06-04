package user

import (
	"context"
	"net/http"

	"github.com/tomiok/webh"
)

type userService interface {
	Create(ctx context.Context, req createUserReq) (User, error)
	Login(ctx context.Context, req loginReq) (loginRes, error)
}

type handler struct {
	service userService
}

func NewHandler(service userService) *handler {
	return &handler{service: service}
}

func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	var req createUserReq
	_, err := webh.DJson(r.Body, &req)
	if err != nil {
		return webh.ErrHTTP{Code: 400, Message: err.Error()}
	}

	res, err := h.service.Create(r.Context(), req)
	if err != nil {
		return webh.ErrHTTP{Code: 500, Message: err.Error()}
	}

	err = webh.EJson(w, res)
	if err != nil {
		return webh.ErrHTTP{Code: 400, Message: err.Error()}
	}

	return nil
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) error {
	var req loginReq
	_, err := webh.DJson(r.Body, &req)
	if err != nil {
		return webh.ErrHTTP{Code: 400, Message: err.Error()}
	}

	res, err := h.service.Login(r.Context(), req)
	if err != nil {
		return webh.ErrHTTP{Code: 500, Message: err.Error()}
	}

	err = webh.EJson(w, res)
	if err != nil {
		return webh.ErrHTTP{Code: 400, Message: err.Error()}
	}

	return nil
}
