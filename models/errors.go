package models

import "errors"

var (
	ErrMissingPermission = errors.New("missing permissions")
	ErrNotOngoing        = errors.New("vote is not ongoing")
	ErrNoAbstain         = errors.New("vote has no abstain rights")
	ErrReservedEndpoint  = errors.New("reserved endpoint")
)
