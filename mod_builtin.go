package dyntpl

import "math"

const (
	round = iota
	roundPrec
	ceil
	ceilPrec
	floor
	floorPrec
)

func modDefault(_ *Ctx, buf *interface{}, val interface{}, args []interface{}) (err error) {
	if len(args) == 0 {
		err = ErrModNoArgs
		return
	}
	if i, ok := ConvInt(val); ok {
		if i == 0 {
			*buf = args[0]
			return
		}
	}
	if u, ok := ConvUint(val); ok {
		if u == 0 {
			*buf = args[0]
			return
		}
	}
	if f, ok := ConvFloat(val); ok {
		if f == 0 {
			*buf = args[0]
			return
		}
	}
	if b, ok := ConvBytes(val); ok {
		if len(b) == 0 {
			*buf = args[0]
			return
		}
	}
	if s, ok := ConvStr(val); ok {
		if len(s) == 0 {
			*buf = args[0]
			return
		}
	}
	if b, ok := ConvBool(val); ok {
		if !b {
			*buf = args[0]
			return
		}
	}
	return nil
}

func modIfThen(_ *Ctx, buf *interface{}, val interface{}, args []interface{}) (err error) {
	if len(args) == 0 {
		err = ErrModNoArgs
		return
	}
	if b, ok := ConvBool(val); ok {
		if b {
			*buf = args[0]
		}
	}
	return
}

func modIfThenElse(_ *Ctx, buf *interface{}, val interface{}, args []interface{}) (err error) {
	if len(args) < 2 {
		err = ErrModPoorArgs
		return
	}
	if b, ok := ConvBool(val); ok {
		if b {
			*buf = args[0]
		} else {
			*buf = args[1]
		}
	}
	return
}

func modRound(ctx *Ctx, buf *interface{}, val interface{}, args []interface{}) (err error) {
	if f, ok := ConvFloat(val); ok {
		ctx.fbuf = roundHelper(f, round, args)
		*buf = &ctx.fbuf
	}
	return
}

func modRoundPrec(ctx *Ctx, buf *interface{}, val interface{}, args []interface{}) (err error) {
	if f, ok := ConvFloat(val); ok {
		ctx.fbuf = roundHelper(f, roundPrec, args)
		*buf = &ctx.fbuf
	}
	return
}

func modCeil(ctx *Ctx, buf *interface{}, val interface{}, args []interface{}) (err error) {
	if f, ok := ConvFloat(val); ok {
		ctx.fbuf = roundHelper(f, ceil, args)
		*buf = &ctx.fbuf
	}
	return
}

func modCeilPrec(ctx *Ctx, buf *interface{}, val interface{}, args []interface{}) (err error) {
	if f, ok := ConvFloat(val); ok {
		ctx.fbuf = roundHelper(f, ceilPrec, args)
		*buf = &ctx.fbuf
	}
	return
}

func modFloor(ctx *Ctx, buf *interface{}, val interface{}, args []interface{}) (err error) {
	if f, ok := ConvFloat(val); ok {
		ctx.fbuf = roundHelper(f, floor, args)
		*buf = &ctx.fbuf
	}
	return
}

func modFloorPrec(ctx *Ctx, buf *interface{}, val interface{}, args []interface{}) (err error) {
	if f, ok := ConvFloat(val); ok {
		ctx.fbuf = roundHelper(f, floorPrec, args)
		*buf = &ctx.fbuf
	}
	return
}

func roundHelper(f float64, mode int, args []interface{}) float64 {
	var (
		prec int64
		ok   bool
	)
	if len(args) > 0 {
		if prec, ok = if2int(args[0]); !ok {
			return f
		}
	}
	switch mode {
	case round:
		return math.Round(f)
	case roundPrec:
		p := math.Pow10(int(prec))
		return float64(int(f*p)) / p
	case ceil:
		return math.Ceil(f)
	case ceilPrec:
		p := math.Pow10(int(prec))
		x := p * f
		return math.Ceil(x) / p
	case floor:
		return math.Floor(f)
	case floorPrec:
		p := math.Pow10(int(prec))
		x := p * f
		return math.Floor(x) / p
	}
	return f
}
