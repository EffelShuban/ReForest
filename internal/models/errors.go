package models

import (
	"errors"
)

var (
	ErrNotFound           = errors.New("resource not found")
	ErrAlreadyExists      = errors.New("resource already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInternal           = errors.New("internal system error")
	ErrInvalidInput       = errors.New("invalid input provided")
	ErrInsufficientBalance = errors.New("insufficient balance")
)