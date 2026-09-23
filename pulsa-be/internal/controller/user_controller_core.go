package controller

import "pulsa2/internal/service"

type UserController struct {
	svc  *service.UserService
	base string
}

func NewUserController(svc *service.UserService, base string) *UserController {
	return &UserController{svc: svc, base: base}
}
