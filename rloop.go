package dyntpl

import (
	"io"

	"github.com/koykov/fastconv"
	"github.com/koykov/inspector"
)

const (
	rlFree  = uint(0)
	rlInuse = uint(1)
)

type RangeLoop struct {
	cntr int
	stat uint
	node Node
	tpl  *Tpl
	ctx  *Ctx
	next *RangeLoop
	w    io.Writer
}

func NewRangeLoop(node Node, tpl *Tpl, ctx *Ctx, w io.Writer) *RangeLoop {
	rl := RangeLoop{
		node: node,
		tpl:  tpl,
		ctx:  ctx,
		w:    w,
	}
	return &rl
}

func (rl *RangeLoop) RequireKey() bool {
	return len(rl.node.loopKey) > 0
}

func (rl *RangeLoop) SetKey(val interface{}, ins inspector.Inspector) {
	rl.ctx.Set(fastconv.B2S(rl.node.loopKey), val, ins)
}

func (rl *RangeLoop) SetVal(val interface{}, ins inspector.Inspector) {
	rl.ctx.Set(fastconv.B2S(rl.node.loopVal), val, ins)
}

func (rl *RangeLoop) Iterate() inspector.LoopCtl {
	if rl.cntr > 0 && len(rl.node.loopSep) > 0 {
		_, _ = rl.w.Write(rl.node.loopSep)
	}
	rl.cntr++
	var err error
	for _, ch := range rl.node.child {
		err = rl.tpl.renderNode(rl.w, ch, rl.ctx)
		if err == ErrBreakLoop {
			return inspector.LoopCtlBrk
		}
		if err == ErrContLoop {
			return inspector.LoopCtlCnt
		}
	}
	return inspector.LoopCtlNone
}

func (rl *RangeLoop) Reset() {
	crl := rl
	for crl != nil {
		crl.stat = rlFree
		crl.cntr = 0
		crl.ctx = nil
		crl.tpl = nil
		crl.w = nil
		crl = crl.next
	}
}
