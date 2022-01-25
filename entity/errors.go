package entity

import (
	"errors"
)

var (
	ErrObjectNotFound  = errors.New("object not found")
	ErrMessageNotFound = errors.New("message not found")
)
