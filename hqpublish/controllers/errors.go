package controllers

import (
	"errors"
)

var (
	ERROR_REQUEST_PARAM    = errors.New("invalid parameter")
	ERROR_KLINE_BEGIN_TIME = errors.New("invalid kline begin time")
)
