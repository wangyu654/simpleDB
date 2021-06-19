package bptree

import "errors"

var HasExistedKeyError = errors.New("hasExistedKey")
var NotFoundKey = errors.New("notFoundKey")
var InvalidDBFormat = errors.New("invalid db format")
var TooLargeVal = errors.New("too large val")
