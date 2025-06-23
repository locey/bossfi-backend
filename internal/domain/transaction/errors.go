package transaction

import "errors"

var (
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrInvalidTransactionHash = errors.New("invalid transaction hash")
	ErrTransactionFailed      = errors.New("transaction failed")
	ErrInvalidAmount          = errors.New("invalid amount")
	ErrInvalidAddress         = errors.New("invalid address")
	ErrInsufficientFee        = errors.New("insufficient fee")
	ErrTransactionCancelled   = errors.New("transaction cancelled")
	ErrDuplicateTransaction   = errors.New("duplicate transaction")
)
