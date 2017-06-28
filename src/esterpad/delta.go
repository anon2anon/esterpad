/*
Esterpad online collaborative editor
Copyright (C) 2017 Anon2Anon

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package esterpad

import (
	"container/list"
	. "esterpad_utils"
	"fmt"
)

var deltaLogger = LogInit("delta")

func DeltaAddInsert(ops *list.List, text []rune, meta *PMeta, textRO bool) {
	if back := ops.Back(); back != nil {
		switch op := back.Value.(type) {
		case *POpInsert:
			if *op.Meta == *meta {
				if op.TextRO {
					t := make([]rune, len(op.Text)+len(text))
					copy(t, op.Text)
					copy(t[len(op.Text):], text)
					op.Text = t
					op.TextRO = false
				} else {
					op.Text = append(op.Text, text...)
				}
				return
			}
		}
	}
	ops.PushBack(&POpInsert{text, meta, textRO})
}

func DeltaAddDelete(ops *list.List, len uint32) {
	if back := ops.Back(); back != nil {
		switch op := back.Value.(type) {
		case *POpDelete:
			op.Len = op.Len + len
			return
		}
	}
	ops.PushBack(&POpDelete{len})
}

func DeltaAddRetain(ops *list.List, len uint32, meta *PMeta) {
	if back := ops.Back(); back != nil {
		switch op := back.Value.(type) {
		case *POpRetain:
			if *op.Meta == *meta {
				op.Len = op.Len + len
				return
			}
		}
	}
	ops.PushBack(&POpRetain{len, meta})
}

func DeltaValidateFromClient(ops []*Op, canWriteWash bool, userId uint32) *list.List {
	listOps := list.New()
	for _, op := range ops {
		switch op := op.Op.(type) {
		case *Op_Insert:
			text := []rune(op.Insert.Text)
			if len(text) > 0 {
				meta := op.Insert.Meta
				pmeta := PMeta{}
				if meta != nil {
					pmeta.Changemask = meta.Changemask&31 | 32
					if meta.Changemask&1 != 0 {
						pmeta.Bold = meta.Bold
					}
					if meta.Changemask&2 != 0 {
						pmeta.Italic = meta.Italic
					}
					if meta.Changemask&4 != 0 {
						pmeta.Underline = meta.Underline
					}
					if meta.Changemask&8 != 0 {
						pmeta.Strike = meta.Strike
					}
					if meta.Changemask&16 != 0 {
						pmeta.FontSize = meta.FontSize
					}
					if meta.Changemask&32 != 0 && canWriteWash {
						if meta.UserId != 0 {
							pmeta.User = CacherGetUser(meta.UserId)
						}
					} else {
						pmeta.User = CacherGetUser(userId)
					}
				} else {
					pmeta.Changemask = 32
					pmeta.User = CacherGetUser(userId)
				}
				DeltaAddInsert(listOps, text, &pmeta, false)
			}
		case *Op_Delete:
			if op.Delete.Len > 0 {
				DeltaAddDelete(listOps, op.Delete.Len)
			}
		case *Op_Retain:
			if op.Retain.Len > 0 {
				meta := op.Retain.Meta
				pmeta := PMeta{}
				if meta != nil {
					pmeta.Changemask = meta.Changemask & 31
					if meta.Changemask&1 != 0 {
						pmeta.Bold = meta.Bold
					}
					if meta.Changemask&2 != 0 {
						pmeta.Italic = meta.Italic
					}
					if meta.Changemask&4 != 0 {
						pmeta.Underline = meta.Underline
					}
					if meta.Changemask&8 != 0 {
						pmeta.Strike = meta.Strike
					}
					if meta.Changemask&16 != 0 {
						pmeta.FontSize = meta.FontSize
					}
					if meta.Changemask&32 != 0 && (canWriteWash || meta.UserId == userId) {
						pmeta.Changemask |= 32
						pmeta.User = CacherGetUser(meta.UserId)
					}
				}
				DeltaAddRetain(listOps, op.Retain.Len, &pmeta)
			}
		}
	}
	return listOps
}

func DeltaMetaAppend(what *PMeta, to *PMeta) *PMeta {
	meta := *to
	if what.Changemask&1 != 0 {
		meta.Bold = what.Bold
	}
	if what.Changemask&2 != 0 {
		meta.Italic = what.Italic
	}
	if what.Changemask&4 != 0 {
		meta.Underline = what.Underline
	}
	if what.Changemask&8 != 0 {
		meta.Strike = what.Strike
	}
	if what.Changemask&16 != 0 {
		meta.FontSize = what.FontSize
	}
	if what.Changemask&32 != 0 {
		meta.User = what.User
	}
	meta.Changemask |= what.Changemask
	return &meta
}

func DeltaMetaComplement(what *PMeta, to *PMeta) *PMeta {
	meta := *to
	if what.Changemask&1 != 0 {
		meta.Bold = false
	}
	if what.Changemask&2 != 0 {
		meta.Italic = false
	}
	if what.Changemask&4 != 0 {
		meta.Underline = false
	}
	if what.Changemask&8 != 0 {
		meta.Strike = false
	}
	if what.Changemask&16 != 0 {
		meta.FontSize = 0
	}
	if what.Changemask&32 != 0 {
		meta.User = nil
	}
	meta.Changemask &= what.Changemask ^ 63
	return &meta
}

func DeltaMetaInvert(what *PMeta, to *PMeta) *PMeta {
	meta := *to
	if what.Changemask&1 == 0 {
		meta.Bold = false
	}
	if what.Changemask&2 == 0 {
		meta.Italic = false
	}
	if what.Changemask&4 == 0 {
		meta.Underline = false
	}
	if what.Changemask&8 == 0 {
		meta.Strike = false
	}
	if what.Changemask&16 == 0 {
		meta.FontSize = 0
	}
	if what.Changemask&32 == 0 {
		meta.User = nil
	}
	meta.Changemask = what.Changemask
	return &meta
}

func DeltaInvert(delta *list.List, text *list.List) *list.List {
	ret := list.New()
	ai := delta.Front()
	bi := text.Front()
	at := 0
	bt := 0
	ac := interface{}(nil)
	bc := interface{}(nil)
	am := (*PMeta)(nil)
	bm := (*PMeta)(nil)
	nexta := func() {
		at = -1
		if ai != nil {
			switch op := ai.Value.(type) {
			case *POpInsert:
				ac = op.Text
				at = 0
				am = op.Meta
			case *POpDelete:
				ac = op.Len
				at = 1
				am = nil
			case *POpRetain:
				ac = op.Len
				at = 2
				am = op.Meta
			}
			ai = ai.Next()
		}
	}
	nextb := func() {
		bt = -1
		if bi != nil {
			switch op := bi.Value.(type) {
			case *POpInsert:
				bc = op.Text
				bt = 0
				bm = op.Meta
			}
			bi = bi.Next()
		}
	}
	nexta()
	nextb()
	for at != -1 && (at == 0 || bt != -1) {
		if at == 0 {
			DeltaAddDelete(ret, uint32(len(ac.([]rune))))
			nexta()
		} else {
			aq := ac.(uint32)
			bq := uint32(len(bc.([]rune)))
			minq := bq
			if aq < bq {
				minq = aq
			}
			if at == 1 {
				DeltaAddInsert(ret, bc.([]rune)[:minq], bm, minq < bq)
			} else if at == 2 {
				DeltaAddRetain(ret, minq, DeltaMetaInvert(am, bm))
			}
			if aq == minq {
				nexta()
			} else {
				ac = aq - minq
			}
			if bq == minq {
				nextb()
			} else {
				bc = bc.([]rune)[minq:]
			}
		}
	}
	if at == -1 && bt == -1 {
		return ret
	}
	return nil
}

func DeltaComposeOld(what *list.List, to *list.List) *list.List {
	ret := list.New()
	ai := what.Front()
	bi := to.Front()
	at := 0
	bt := 0
	ac := interface{}(nil)
	bc := interface{}(nil)
	am := (*PMeta)(nil)
	bm := (*PMeta)(nil)
	nexta := func() {
		at = -1
		if ai != nil {
			switch op := ai.Value.(type) {
			case *POpInsert:
				ac = op.Text
				at = 0
				am = op.Meta
			case *POpDelete:
				ac = op.Len
				at = 1
				am = nil
			case *POpRetain:
				ac = op.Len
				at = 2
				am = op.Meta
			}
			ai = ai.Next()
		}
	}
	nextb := func() {
		bt = -1
		if bi != nil {
			switch op := bi.Value.(type) {
			case *POpInsert:
				bc = op.Text
				bt = 0
				bm = op.Meta
			case *POpDelete:
				bc = op.Len
				bt = 1
				bm = nil
			case *POpRetain:
				bc = op.Len
				bt = 2
				bm = op.Meta
			}
			bi = bi.Next()
		}
	}
	nexta()
	nextb()
	for (at != -1 || bt == 1) && (at == 0 || bt != -1) {
		if bt == 1 {
			DeltaAddDelete(ret, bc.(uint32))
			nextb()
		} else if at == 0 {
			DeltaAddInsert(ret, ac.([]rune), am, false)
			nexta()
		} else {
			aq := ac.(uint32)
			bq := uint32(0)
			if bt == 0 {
				bq = uint32(len(bc.([]rune)))
			} else {
				bq = bc.(uint32)
			}
			minq := bq
			if aq < bq {
				minq = aq
			}
			if at == 2 && bt == 2 {
				DeltaAddRetain(ret, minq, DeltaMetaAppend(am, bm))
			} else if at == 2 && bt == 0 {
				DeltaAddInsert(ret, bc.([]rune)[:minq], DeltaMetaAppend(am, bm), minq < bq)
			} else if at == 1 && bt == 2 {
				DeltaAddDelete(ret, minq)
			}
			if aq == minq {
				nexta()
			} else {
				ac = aq - minq
			}
			if bq == minq {
				nextb()
			} else {
				if bt == 0 {
					bc = bc.([]rune)[minq:]
				} else {
					bc = bq - minq
				}
			}
		}
	}
	if at == -1 && bt == -1 {
		return ret
	}
	return nil
}

func DeltaCompose(what *list.List, to *list.List, canWriteWash bool, canEdit bool, user *User) [2]*list.List {
	an := list.New()
	bn := list.New()
	ai := what.Front()
	bi := to.Front()
	at := 0
	bt := 0
	ac := interface{}(nil)
	bc := interface{}(nil)
	am := (*PMeta)(nil)
	bm := (*PMeta)(nil)
	nexta := func() {
		at = -1
		if ai != nil {
			switch op := ai.Value.(type) {
			case *POpInsert:
				ac = op.Text
				at = 0
				am = op.Meta
			case *POpDelete:
				ac = op.Len
				at = 1
				am = nil
			case *POpRetain:
				ac = op.Len
				at = 2
				am = op.Meta
			}
			ai = ai.Next()
		}
	}
	nextb := func() {
		bt = -1
		if bi != nil {
			switch op := bi.Value.(type) {
			case *POpInsert:
				bc = op.Text
				bt = 0
				bm = op.Meta
			case *POpDelete:
				bc = op.Len
				bt = 1
				bm = nil
			case *POpRetain:
				bc = op.Len
				bt = 2
				bm = op.Meta
			}
			bi = bi.Next()
		}
	}
	nexta()
	nextb()
	for (at != -1 || bt == 1) && (at == 0 || bt != -1) {
		if bt == 1 {
			DeltaAddDelete(bn, bc.(uint32))
			nextb()
		} else if at == 0 {
			DeltaAddInsert(an, ac.([]rune), am, false)
			DeltaAddInsert(bn, ac.([]rune), am, false)
			nexta()
		} else {
			aq := ac.(uint32)
			bq := uint32(0)
			if bt == 0 {
				bq = uint32(len(bc.([]rune)))
			} else {
				bq = bc.(uint32)
			}
			minq := bq
			if aq < bq {
				minq = aq
			}
			if at == 2 && bt == 2 {
				if canWriteWash || canEdit && bm.User != nil || bm.User == user {
					DeltaAddRetain(an, minq, am)
					DeltaAddRetain(bn, minq, DeltaMetaAppend(am, bm))
				} else {
					DeltaAddRetain(an, minq, &PMeta{})
					DeltaAddRetain(bn, minq, bm)
				}
			} else if at == 2 && bt == 0 {
				if canWriteWash || canEdit && bm.User != nil || bm.User == user {
					DeltaAddRetain(an, minq, am)
					DeltaAddInsert(bn, bc.([]rune)[:minq], DeltaMetaAppend(am, bm), minq < bq)
				} else {
					DeltaAddRetain(an, minq, &PMeta{})
					DeltaAddInsert(bn, bc.([]rune)[:minq], bm, minq < bq)
				}
			} else if at == 1 && bt == 2 {
				if canWriteWash || canEdit && bm.User != nil || bm.User == user {
					DeltaAddDelete(an, minq)
					DeltaAddDelete(bn, minq)
				} else {
					DeltaAddRetain(an, minq, &PMeta{})
					DeltaAddRetain(bn, minq, bm)
				}
			} else if at == 1 && bt == 0 {
				if canWriteWash || canEdit && bm.User != nil || bm.User == user {
					DeltaAddDelete(an, minq)
				} else {
					DeltaAddRetain(an, minq, &PMeta{})
					DeltaAddInsert(bn, bc.([]rune)[:minq], bm, minq < bq)
				}
			}
			if aq == minq {
				nexta()
			} else {
				ac = aq - minq
			}
			if bq == minq {
				nextb()
			} else {
				if bt == 0 {
					bc = bc.([]rune)[minq:]
				} else {
					bc = bq - minq
				}
			}
		}
	}
	if at == -1 && bt == -1 {
		return [2]*list.List{an, bn}
	}
	return [2]*list.List{nil, nil}
}

func DeltaTransform(a *list.List, b *list.List) *list.List {
	an := list.New()
	//bn := list.New()
	ai := a.Front()
	bi := b.Front()
	at := 0
	bt := 0
	ac := interface{}(nil)
	bc := interface{}(nil)
	am := (*PMeta)(nil)
	bm := (*PMeta)(nil)
	nexta := func() {
		at = -1
		if ai != nil {
			switch op := ai.Value.(type) {
			case *POpInsert:
				ac = op.Text
				at = 0
				am = op.Meta
			case *POpDelete:
				ac = op.Len
				at = 1
				am = nil
			case *POpRetain:
				ac = op.Len
				at = 2
				am = op.Meta
			}
			ai = ai.Next()
		}
	}
	nextb := func() {
		bt = -1
		if bi != nil {
			switch op := bi.Value.(type) {
			case *POpInsert:
				bc = op.Text
				bt = 0
				bm = op.Meta
			case *POpDelete:
				bc = op.Len
				bt = 1
				bm = nil
			case *POpRetain:
				bc = op.Len
				bt = 2
				bm = op.Meta
			}
			bi = bi.Next()
		}
	}
	nexta()
	nextb()
	for (at != -1 || bt == 0) && (at == 0 || bt != -1) {
		if at == 0 {
			DeltaAddInsert(an, ac.([]rune), am, false)
			//DeltaAddRetain(bn, uint32(len(ac.([]rune))), &PMeta{})
			nexta()
		} else if bt == 0 {
			DeltaAddRetain(an, uint32(len(bc.([]rune))), &PMeta{})
			//DeltaAddInsert(bn, bc.([]rune), bm, false)
			nextb()
		} else {
			aq := ac.(uint32)
			bq := bc.(uint32)
			minq := bq
			if aq < bq {
				minq = aq
			}
			if at == 2 && bt == 2 {
				DeltaAddRetain(an, minq, am)
				//	DeltaAddRetain(bn, minq, DeltaMetaComplement(am, bm))
			} else if bt == 2 {
				DeltaAddDelete(an, minq)
			} /* else if at == 2 {
				DeltaAddDelete(bn, minq)
			}*/
			if aq == minq {
				nexta()
			} else {
				ac = aq - minq
			}
			if bq == minq {
				nextb()
			} else {
				bc = bq - minq
			}
		}
	}
	if at == -1 && bt == -1 {
		return an
	}
	return nil
}

func DeltaToProtobuf(opsList *list.List) []*Op {
	ops := make([]*Op, opsList.Len())
	count := 0
	for op := opsList.Front(); op != nil; op = op.Next() {
		switch op := op.Value.(type) {
		case *POpInsert:
			meta := OpMeta{op.Meta.Changemask, op.Meta.Bold, op.Meta.Italic, op.Meta.Underline, op.Meta.Strike, op.Meta.FontSize, 0}
			if op.Meta.User != nil {
				meta.UserId = op.Meta.User.Id
			}
			ops[count] = &Op{&Op_Insert{&OpInsert{string(op.Text), &meta}}}
		case *POpDelete:
			ops[count] = &Op{&Op_Delete{&OpDelete{op.Len}}}
		case *POpRetain:
			meta := OpMeta{op.Meta.Changemask, op.Meta.Bold, op.Meta.Italic, op.Meta.Underline, op.Meta.Strike, op.Meta.FontSize, 0}
			if op.Meta.User != nil {
				meta.UserId = op.Meta.User.Id
			}
			ops[count] = &Op{&Op_Retain{&OpRetain{op.Len, &meta}}}
		}
		count++
	}
	return ops
}

func DeltaToString(opsList *list.List) string {
	ops := make([]*Op, opsList.Len())
	count := 0
	for op := opsList.Front(); op != nil; op = op.Next() {
		switch op := op.Value.(type) {
		case *POpInsert:
			meta := OpMeta{op.Meta.Changemask, op.Meta.Bold, op.Meta.Italic, op.Meta.Underline, op.Meta.Strike, op.Meta.FontSize, 0}
			if op.Meta.User != nil {
				meta.UserId = op.Meta.User.Id
			}
			ops[count] = &Op{&Op_Insert{&OpInsert{string(op.Text), &meta}}}
		case *POpDelete:
			ops[count] = &Op{&Op_Delete{&OpDelete{op.Len}}}
		case *POpRetain:
			meta := OpMeta{op.Meta.Changemask, op.Meta.Bold, op.Meta.Italic, op.Meta.Underline, op.Meta.Strike, op.Meta.FontSize, 0}
			if op.Meta.User != nil {
				meta.UserId = op.Meta.User.Id
			}
			ops[count] = &Op{&Op_Retain{&OpRetain{op.Len, &meta}}}
		}
		count++
	}
	return fmt.Sprintf("%v", ops)
}
