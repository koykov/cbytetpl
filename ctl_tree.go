package cbytetpl

import (
	"bytes"
)

type Type int
type Op int

const (
	TypeRaw       Type = 0
	TypeTpl       Type = 1
	TypeCond      Type = 2
	TypeCondTrue  Type = 3
	TypeCondFalse Type = 4
	TypeLoopRange Type = 5
	TypeLoopCount Type = 6
	TypeBreak     Type = 7
	TypeContinue  Type = 8
	TypeCtx       Type = 9
	TypeSwitch    Type = 10
	TypeCase      Type = 11
	TypeDefault   Type = 12
	TypeDiv       Type = 13
	TypeExit      Type = 99

	// Must be in sync with inspector.Op type.
	OpUnk Op = 0
	OpEq  Op = 1
	OpNq  Op = 2
	OpGt  Op = 3
	OpGtq Op = 4
	OpLt  Op = 5
	OpLtq Op = 6
	OpInc Op = 7
	OpDec Op = 8
)

func (typ Type) String() string {
	switch typ {
	case TypeRaw:
		return "raw"
	case TypeTpl:
		return "tpl"
	case TypeCond:
		return "cond"
	case TypeCondTrue:
		return "true"
	case TypeCondFalse:
		return "false"
	case TypeLoopRange:
		return "rloop"
	case TypeLoopCount:
		return "cloop"
	case TypeBreak:
		return "break"
	case TypeContinue:
		return "cont"
	case TypeCtx:
		return "ctx"
	case TypeSwitch:
		return "switch"
	case TypeCase:
		return "case"
	case TypeDefault:
		return "def"
	case TypeDiv:
		return "div"
	case TypeExit:
		return "exit"
	default:
		return "unk"
	}
}

func (o Op) String() string {
	switch o {
	case OpEq:
		return "=="
	case OpNq:
		return "!="
	case OpGt:
		return ">"
	case OpGtq:
		return ">="
	case OpLt:
		return "<"
	case OpLtq:
		return "<="
	case OpInc:
		return "++"
	case OpDec:
		return "--"
	default:
		return "unk"
	}
}

func (o Op) Swap() Op {
	switch o {
	case OpGt:
		return OpLt
	case OpGtq:
		return OpLtq
	case OpLt:
		return OpGt
	case OpLtq:
		return OpGtq
	default:
		return o
	}
}

type Tree struct {
	nodes []Node
}

type Node struct {
	typ    Type
	raw    []byte
	prefix []byte
	suffix []byte

	ctxVar       []byte
	ctxSrc       []byte
	ctxSrcStatic bool
	ctxIns       []byte

	condL       []byte
	condR       []byte
	condStaticL bool
	condStaticR bool
	condOp      Op

	loopKey       []byte
	loopVal       []byte
	loopSrc       []byte
	loopCnt       []byte
	loopCntInit   []byte
	loopCntStatic bool
	loopCntOp     Op
	loopCondOp    Op
	loopLim       []byte
	loopLimStatic bool
	loopSep       []byte

	switchArg []byte

	caseL       []byte
	caseR       []byte
	caseStaticL bool
	caseStaticR bool
	caseOp      Op

	mod []mod

	child []Node
}

func (t *Tree) humanReadable() []byte {
	if len(t.nodes) == 0 {
		return nil
	}
	var buf bytes.Buffer
	t.hrHelper(&buf, t.nodes, []byte("\t"), 0)
	return buf.Bytes()
}

func (t *Tree) hrHelper(buf *bytes.Buffer, nodes []Node, indent []byte, depth int) {
	for _, node := range nodes {
		buf.Write(bytes.Repeat(indent, depth))
		buf.WriteString(node.typ.String())
		if node.typ != TypeExit && node.typ != TypeBreak && node.typ != TypeContinue {
			buf.WriteByte(':')
			buf.WriteByte(' ')
			buf.Write(node.raw)
		}
		if len(node.prefix) > 0 {
			buf.WriteString(" pfx ")
			buf.Write(node.prefix)
		}
		if len(node.suffix) > 0 {
			buf.WriteString(" sfx ")
			buf.Write(node.suffix)
		}

		if len(node.ctxVar) > 0 && len(node.ctxSrc) > 0 {
			buf.WriteString("var ")
			buf.Write(node.ctxVar)
			buf.WriteString(" src ")
			buf.Write(node.ctxSrc)
			if len(node.ctxIns) > 0 {
				buf.WriteString(" ins ")
				buf.Write(node.ctxIns)
			}
		}

		if len(node.condL) > 0 {
			buf.WriteString("left ")
			buf.Write(node.condL)
		}
		if node.condOp != 0 {
			buf.WriteString(" op ")
			buf.WriteString(node.condOp.String())
		}
		if len(node.condR) > 0 {
			buf.WriteString(" right ")
			buf.Write(node.condR)
		}

		if len(node.loopKey) > 0 {
			buf.WriteString("key ")
			buf.Write(node.loopKey)
			buf.WriteByte(' ')
		}
		if len(node.loopVal) > 0 {
			buf.WriteString("val ")
			buf.Write(node.loopVal)
		}
		if len(node.loopSrc) > 0 {
			buf.WriteString(" src ")
			buf.Write(node.loopSrc)
		}
		if len(node.loopCnt) > 0 {
			buf.WriteString("cnt ")
			buf.Write(node.loopCnt)
		}
		if node.loopCondOp != 0 {
			buf.WriteString(" cond ")
			buf.WriteString(node.loopCondOp.String())
		}
		if len(node.loopLim) > 0 {
			buf.WriteString(" lim ")
			buf.Write(node.loopLim)
		}
		if node.loopCntOp != 0 {
			buf.WriteString(" op ")
			buf.WriteString(node.loopCntOp.String())
		}
		if len(node.loopSep) > 0 {
			buf.WriteString(" sep ")
			buf.Write(node.loopSep)
		}

		if len(node.switchArg) > 0 {
			buf.WriteString("arg ")
			buf.Write(node.switchArg)
		}
		if len(node.caseL) > 0 && node.caseOp != 0 && len(node.caseR) > 0 {
			buf.WriteString("left ")
			buf.Write(node.caseL)
			buf.WriteString(" op ")
			buf.WriteString(node.caseOp.String())
			buf.WriteString(" right ")
			buf.Write(node.caseR)
		} else if len(node.caseL) > 0 {
			buf.WriteString("val ")
			buf.Write(node.caseL)
		}

		if len(node.mod) > 0 {
			buf.WriteString(" mod")
			for i, mod := range node.mod {
				if i > 0 {
					buf.WriteByte(',')
				}
				buf.WriteByte(' ')
				buf.Write(mod.id)
				if len(mod.arg) > 0 {
					buf.WriteByte('(')
					for j, a := range mod.arg {
						if j > 0 {
							buf.WriteByte(',')
							buf.WriteByte(' ')
						}
						if a.static {
							buf.WriteByte('"')
							buf.Write(a.val)
							buf.WriteByte('"')
						} else {
							buf.Write(a.val)
						}
					}
					buf.WriteByte(')')
				}
			}
		}

		buf.WriteByte('\n')
		if len(node.child) > 0 {
			t.hrHelper(buf, node.child, indent, depth+1)
		}
	}
}

func addNode(nodes []Node, node Node) []Node {
	nodes = append(nodes, node)
	return nodes
}

func addRaw(nodes []Node, raw []byte) []Node {
	if len(raw) == 0 {
		return nodes
	}
	nodes = append(nodes, Node{typ: TypeRaw, raw: raw})
	return nodes
}

func splitNodes(nodes []Node) [][]Node {
	if len(nodes) == 0 {
		return nil
	}
	split := make([][]Node, 0)
	var o int
	for i, node := range nodes {
		if node.typ == TypeDiv {
			split = append(split, nodes[o:i])
			o = i + 1
		}
	}
	if o < len(nodes) {
		split = append(split, nodes[o:])
	}
	return split
}

func rollupSwitchNodes(nodes []Node) []Node {
	if len(nodes) == 0 {
		return nil
	}
	var (
		r     = make([]Node, 0)
		group = Node{typ: -1}
	)
	for _, node := range nodes {
		if node.typ != TypeCase && node.typ != TypeDefault && group.typ == -1 {
			continue
		}
		if node.typ == TypeCase || node.typ == TypeDefault {
			if group.typ != -1 {
				r = append(r, group)
			}
			group = node
			continue
		}
		group.child = append(group.child, node)
	}
	if len(group.child) > 0 {
		r = append(r, group)
	}
	return r
}
