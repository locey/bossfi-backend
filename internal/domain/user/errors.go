package user

import "errors"

// 用户相关错误
var (
	ErrUserNotFound               = errors.New("user not found")
	ErrUserAlreadyExists          = errors.New("user already exists")
	ErrWalletAddressAlreadyExists = errors.New("wallet address already exists")
	ErrUsernameAlreadyExists      = errors.New("username already exists")
	ErrEmailAlreadyExists         = errors.New("email already exists")
	ErrInvalidWalletAddress       = errors.New("invalid wallet address")
	ErrInvalidEmail               = errors.New("invalid email format")
	ErrInvalidUsername            = errors.New("invalid username format")
	ErrInsufficientBalance        = errors.New("insufficient BOSS balance")
	ErrInsufficientStakedAmount   = errors.New("insufficient staked amount")
	ErrProfileIncomplete          = errors.New("user profile is incomplete")
	ErrUnauthorized               = errors.New("unauthorized access")
)
