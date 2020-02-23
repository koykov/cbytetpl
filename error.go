package cbytetpl

import "errors"

var (
	ErrUnexpectedEOF = errors.New("unexpected end of file: control structure couldn't be closed")
	ErrBadCtl        = errors.New("bad control structure caught")
)
