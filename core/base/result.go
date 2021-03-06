package base

import (
	"fmt"
)

type BlockType int32

const (
	// sentinel internal error
	BlockTypeUnknown BlockType = iota
	BlockTypeFlow
	BlockTypeCircuitBreaking
)

type TokenResultStatus int32

const (
	ResultStatusPass TokenResultStatus = iota
	ResultStatusBlocked
	ResultStatusShouldWait
)

type TokenResult struct {
	status TokenResultStatus

	blockErr *BlockError
	waitMs   uint64
}

func (r *TokenResult) IsPass() bool {
	return r.status == ResultStatusPass
}

func (r *TokenResult) IsBlocked() bool {
	return r.status == ResultStatusBlocked
}

func (r *TokenResult) Status() TokenResultStatus {
	return r.status
}

func (r *TokenResult) BlockError() *BlockError {
	return r.blockErr
}

func (r *TokenResult) WaitMs() uint64 {
	return r.waitMs
}

func (r *TokenResult) String() string {
	var blockMsg string
	if r.blockErr == nil {
		blockMsg = "none"
	} else {
		blockMsg = r.blockErr.Error()
	}
	return fmt.Sprintf("TokenResult{status=%d, blockErr=%s, waitMs=%d}", r.status, blockMsg, r.waitMs)
}

func NewTokenResultPass() *TokenResult {
	return &TokenResult{status: ResultStatusPass, waitMs: 0}
}

func NewTokenResultBlocked(blockType BlockType, blockMsg string) *TokenResult {
	return &TokenResult{
		status:   ResultStatusBlocked,
		blockErr: NewBlockError(blockType, blockMsg),
		waitMs:   0,
	}
}

func NewTokenResultShouldWait(waitMs uint64) *TokenResult {
	return &TokenResult{status: ResultStatusShouldWait, waitMs: waitMs}
}
