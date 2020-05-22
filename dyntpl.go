package dyntpl

import (
	"bytes"
	"io"
	"sync"

	"github.com/koykov/cbytealg"
	"github.com/koykov/fastconv"
	"github.com/koykov/inspector"
)

type Tpl struct {
	Id   string
	tree *Tree
}

var (
	mux         sync.Mutex
	tplRegistry = map[string]*Tpl{}
)

func RegisterTpl(id string, tree *Tree) {
	tpl := Tpl{
		Id:   id,
		tree: tree,
	}
	mux.Lock()
	tplRegistry[id] = &tpl
	mux.Unlock()
}

func Render(id string, ctx *Ctx) ([]byte, error) {
	buf := bytes.Buffer{}
	err := RenderTo(&buf, id, ctx)
	return buf.Bytes(), err
}

func RenderTo(w io.Writer, id string, ctx *Ctx) (err error) {
	mux.Lock()
	tpl, ok := tplRegistry[id]
	mux.Unlock()
	if !ok {
		err = ErrTplNotFound
		return
	}
	for _, node := range tpl.tree.nodes {
		err = tpl.renderNode(w, node, ctx)
		if err != nil {
			if err == ErrInterrupt {
				err = nil
			}
			return
		}
	}

	return
}

func (t *Tpl) renderNode(w io.Writer, node Node, ctx *Ctx) (err error) {
	switch node.typ {
	case TypeRaw:
		_, err = w.Write(node.raw)
	case TypeTpl:
		var raw interface{}
		if node.rawStatic {
			raw = node.raw
		} else {
			raw = ctx.getSsc(node.raw, node.rawSsc)
		}
		if ctx.Err != nil {
			err = ctx.Err
			return
		}
		if len(node.mod) > 0 {
			for _, mod := range node.mod {
				ctx.modA = ctx.modA[:0]
				if len(mod.arg) > 0 {
					for _, arg := range mod.arg {
						if arg.static {
							ctx.modA = append(ctx.modA, &arg.val)
						} else {
							val := ctx.getSsc(arg.val, arg.ssc)
							ctx.modA = append(ctx.modA, val)
						}
					}
				}
				ctx.buf = raw
				ctx.Err = (*mod.fn)(ctx, &ctx.buf, ctx.buf, ctx.modA)
				if ctx.Err != nil {
					break
				}
				raw = ctx.buf
			}
		}
		if ctx.Err != nil {
			return
		}
		if raw == nil || raw == "" {
			err = ErrEmptyArg
			return
		}
		ctx.bbuf, err = cbytealg.AnyToBytes(ctx.bbuf, raw)
		if err == nil {
			if len(node.prefix) > 0 {
				_, _ = w.Write(node.prefix)
			}
			_, err = w.Write(ctx.bbuf)
			if len(node.suffix) > 0 {
				_, _ = w.Write(node.suffix)
			}
		}
	case TypeCtx:
		if node.ctxSrcStatic {
			ctx.SetBytes(fastconv.B2S(node.ctxVar), node.ctxSrc)
		} else {
			ins, err := inspector.GetInspector(fastconv.B2S(node.ctxIns))
			if err != nil {
				return err
			}
			ctx.Set(fastconv.B2S(node.ctxVar), ctx.getSsc(node.ctxSrc, node.ctxSrcSsc), ins)
		}
	case TypeCond:
		sl := node.condLStatic
		sr := node.condRStatic
		if sl && sr {
			err = ErrSenselessCond
			return
		}
		var r bool
		if sr {
			// Right side is static.
			r = ctx.cmp(node.condL, node.condOp, node.condR)
		} else if sl {
			// Left side is static.
			r = ctx.cmp(node.condR, node.condOp.Swap(), node.condL)
		} else {
			// Both sides isn't static.
			ctx.getSsc(node.condR, node.condRSsc)
			if ctx.Err == nil {
				ctx.bbuf, err = cbytealg.AnyToBytes(ctx.bbuf[:0], ctx.buf)
				if err != nil {
					return
				}
				r = ctx.cmp(node.condL, node.condOp, ctx.bbuf)
			}
		}
		if ctx.Err != nil {
			err = ctx.Err
			return
		}
		if r {
			if len(node.child) > 0 {
				err = t.renderNode(w, node.child[0], ctx)
			}
		} else {
			if len(node.child) > 1 {
				err = t.renderNode(w, node.child[0], ctx)
			}
		}
	case TypeCondTrue, TypeCondFalse, TypeCase, TypeDefault:
		for _, ch := range node.child {
			err = t.renderNode(w, ch, ctx)
			if err != nil {
				return
			}
		}
	case TypeLoopCount:
		ctx.cloop(node, t, w)
		if ctx.Err != nil {
			err = ctx.Err
			return
		}
	case TypeLoopRange:
		ctx.rloop(node.loopSrc, node, t, w)
		if ctx.Err != nil {
			err = ctx.Err
			return
		}
	case TypeBreak:
		err = ErrBreakLoop
	case TypeContinue:
		err = ErrContLoop
	case TypeSwitch:
		r := false
		if len(node.switchArg) > 0 {
			// Classic switch case.
			for _, ch := range node.child {
				if ch.typ == TypeCase {
					if ch.caseLStatic {
						r = ctx.cmp(node.switchArg, OpEq, ch.caseL)
					} else {
						ctx.getSsc(ch.caseL, ch.caseLSsc)
						if ctx.Err == nil {
							ctx.bbuf, err = cbytealg.AnyToBytes(ctx.bbuf[:0], ctx.buf)
							if err != nil {
								return
							}
							r = ctx.cmp(node.switchArg, OpEq, ctx.bbuf)
						}
					}
				}
				if r {
					err = t.renderNode(w, ch, ctx)
					break
				}
			}
		} else {
			// Switch without condition case.
			for _, ch := range node.child {
				if ch.typ == TypeCase {
					sl := ch.caseLStatic
					sr := ch.caseRStatic
					if sl && sr {
						err = ErrSenselessCond
						return
					}
					if sr {
						// Right side is static.
						r = ctx.cmp(ch.caseL, ch.caseOp, ch.caseR)
					} else if sl {
						// Left side is static.
						r = ctx.cmp(ch.caseR, ch.caseOp.Swap(), ch.caseL)
					} else {
						// Both sides isn't static.
						ctx.getSsc(ch.caseR, ch.caseRSsc)
						if ctx.Err == nil {
							ctx.bbuf, err = cbytealg.AnyToBytes(ctx.bbuf[:0], ctx.buf)
							if err != nil {
								return
							}
							r = ctx.cmp(ch.caseL, ch.caseOp, ctx.bbuf)
						}
					}
					if ctx.Err != nil {
						err = ctx.Err
						return
					}
					if r {
						err = t.renderNode(w, ch, ctx)
						break
					}
				}
			}
		}
		if !r {
			for _, ch := range node.child {
				if ch.typ == TypeDefault {
					err = t.renderNode(w, ch, ctx)
					break
				}
			}
		}
	case TypeExit:
		err = ErrInterrupt
	default:
		err = ErrUnknownCtl
	}
	return
}
