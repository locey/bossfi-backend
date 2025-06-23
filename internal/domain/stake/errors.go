package stake

import "errors"

// 质押相关错误
var (
	ErrStakeNotFound           = errors.New("stake not found")
	ErrInvalidStakeAmount      = errors.New("invalid stake amount")
	ErrInsufficientBalance     = errors.New("insufficient balance to stake")
	ErrCannotUnstake           = errors.New("cannot unstake")
	ErrCannotCompleteUnstake   = errors.New("cannot complete unstake")
	ErrUnstakeDelayNotMet      = errors.New("unstake delay period not met")
	ErrStakeAlreadyCompleted   = errors.New("stake already completed")
	ErrMinStakeAmountNotMet    = errors.New("minimum stake amount not met")
	ErrRewardCalculationFailed = errors.New("reward calculation failed")
	ErrInvalidStakeStatus      = errors.New("invalid stake status")
)
