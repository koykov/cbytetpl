package cbytetpl

import (
	"bytes"
	"io"
	"sync"
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
		err = tpl.renderNode(w, &node, ctx)
		if err != nil {
			return
		}
	}

	return
}

func (t *Tpl) renderNode(w io.Writer, node *Node, ctx *Ctx) (err error) {
	switch node.typ {
	case TypeRaw:
		_, err = w.Write(node.raw)
	case TypeTpl:
		raw := ctx.get(node.raw)
		if ctx.Err != nil {
			err = ctx.Err
			return
		}
		if raw == nil || raw == "" {
			err = ErrEmptyArg
			return
		}
		for _, bcFn := range byteConvFnRegistry {
			ctx.bbuf = ctx.bbuf[:0]
			ctx.bbuf, err = bcFn(ctx.bbuf, raw)
			if err == nil && len(ctx.bbuf) > 0 {
				_, err = w.Write(ctx.bbuf)
				break
			}
		}
	case TypeCond:
		sl := node.condStaticL
		sr := node.condStaticR
		if sl && sr {
			err = ErrSenselessCond
			return
		}
		var r bool
		if sr {
			r = ctx.cmp(node.condL, node.condOp, node.condR)
		}
		if ctx.Err != nil {
			err = ctx.Err
			return
		}
		if sl {
			r = ctx.cmp(node.condR, node.condOp.Swap(), node.condL)
		}
		if ctx.Err != nil {
			err = ctx.Err
			return
		}
		if r {
			if len(node.child) > 0 {
				err = t.renderNode(w, &node.child[0], ctx)
			}
		} else {
			if len(node.child) > 1 {
				err = t.renderNode(w, &node.child[0], ctx)
			}
		}
	case TypeCondTrue, TypeCondFalse:
		for _, ch := range node.child {
			err = t.renderNode(w, &ch, ctx)
			if err != nil {
				return
			}
		}
	case TypeLoopCount:
		// todo implement me
	case TypeLoopRange:
		ctx.tpl = t
		ctx.node = *node
		ctx.w = w
		ctx.loop(node.loopSrc)
	default:
		err = ErrUnknownCtl
	}
	return
}

func init() {
	RegisterByteConvFn(byteConvBytes)
	RegisterByteConvFn(byteConvStr)
	RegisterByteConvFn(byteConvBool)
	RegisterByteConvFn(byteConvInt)
	RegisterByteConvFn(byteConvUint)
	RegisterByteConvFn(byteConvFloat)
}
