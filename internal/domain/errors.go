package domain

import "errors"

var (
	ErrInvalidUsername                   = errors.New("invalid username")
	ErrInvalidEmail                      = errors.New("invalid email")
	ErrInvalidPasswordHash               = errors.New("invalid password hash")
	ErrInvalidRole                       = errors.New("invalid role")
	ErrInvalidTransactionAmount          = errors.New("invalid transaction amount")
	ErrInvalidTransactionType            = errors.New("invalid transaction type")
	ErrInvalidTransactionStatus          = errors.New("invalid transaction status")
	ErrInvalidTransactionUsers           = errors.New("invalid transaction users")
	ErrSameSenderReceiver                = errors.New("sender and receiver cannot be the same")
	ErrInvalidTransactionStateTransition = errors.New("invalid transaction state transition")
	ErrInvalidBalanceUserID              = errors.New("invalid balance user id")
	ErrInvalidBalanceAmount              = errors.New("invalid balance amount")
	ErrInsufficientBalance               = errors.New("insufficient balance")
	ErrBalanceNotFound                   = errors.New("balance not found")
	ErrUserNotFound                      = errors.New("user not found")
	ErrEmailAlreadyExists                = errors.New("email already exists")
	ErrUsernameAlreadyExists             = errors.New("username already exists")
	ErrInvalidCredentials                = errors.New("invalid credentials")
	ErrInvalidPassword                   = errors.New("invalid password")
	ErrUnauthorized                      = errors.New("unauthorized")
)
