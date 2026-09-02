package domain

import "errors"

var ErrNotFound = errors.New("record not found")
var ErrInvalidAmount = errors.New("invalid amount")
var ErrAmountTooHigh = errors.New("amount too high")
var ErrInsufficientStock = errors.New("insufficient stock")