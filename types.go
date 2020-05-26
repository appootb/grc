package grc

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync/atomic"
)

type UpdateEvent func()

type AtomicUpdate interface {
	Store(v string)
}

type EmbedString struct {
	v atomic.Value
}

func newEmbedString(v string) *EmbedString {
	es := &EmbedString{}
	es.v.Store(v)
	return es
}

func (t *EmbedString) String() string {
	v := t.v.Load()
	if v == nil {
		return ""
	}
	return v.(string)
}

type String struct {
	EmbedString
}

func (t *String) Store(v string) {
	if t.String() == v {
		return
	}
	t.v.Store(v)
	CallbackMgr.EvtChan() <- t
}

func (t *String) Changed(evt UpdateEvent) {
	CallbackMgr.RegChan() <- &RegisterCallback{
		Val: t,
		Evt: evt,
	}
}

type EmbedBool struct {
	v int32
}

func (t *EmbedBool) String() string {
	return strconv.FormatBool(t.Bool())
}

func (t *EmbedBool) Bool() bool {
	return atomic.LoadInt32(&t.v) == 1
}

type Bool struct {
	EmbedBool
}

func (t *Bool) Store(v string) {
	b, _ := strconv.ParseBool(v)
	if t.Bool() == b {
		return
	}
	if b {
		atomic.StoreInt32(&t.v, 1)
	} else {
		atomic.StoreInt32(&t.v, 0)
	}
	CallbackMgr.EvtChan() <- t
}

func (t *Bool) Changed(evt UpdateEvent) {
	CallbackMgr.RegChan() <- &RegisterCallback{
		Val: t,
		Evt: evt,
	}
}

type EmbedInt struct {
	v int64
}

func (t *EmbedInt) String() string {
	return strconv.FormatInt(t.Int64(), 10)
}

func (t *EmbedInt) Int() int {
	return int(t.Int64())
}

func (t *EmbedInt) Int8() int8 {
	return int8(t.Int64())
}

func (t *EmbedInt) Int16() int16 {
	return int16(t.Int64())
}

func (t *EmbedInt) Int32() int32 {
	return int32(t.Int64())
}

func (t *EmbedInt) Int64() int64 {
	return atomic.LoadInt64(&t.v)
}

type Int struct {
	EmbedInt
}

func (t *Int) Store(v string) {
	iv, _ := strconv.ParseInt(v, 10, 64)
	if t.Int64() == iv {
		return
	}
	atomic.StoreInt64(&t.v, iv)
	CallbackMgr.EvtChan() <- t
}

func (t *Int) Changed(evt UpdateEvent) {
	CallbackMgr.RegChan() <- &RegisterCallback{
		Val: t,
		Evt: evt,
	}
}

type EmbedUint struct {
	v uint64
}

func (t *EmbedUint) String() string {
	return strconv.FormatUint(t.Uint64(), 10)
}

func (t *EmbedUint) Uint() uint {
	return uint(t.Uint64())
}

func (t *EmbedUint) Uint8() uint8 {
	return uint8(t.Uint64())
}

func (t *EmbedUint) Uint16() uint16 {
	return uint16(t.Uint64())
}

func (t *EmbedUint) Uint32() uint32 {
	return uint32(t.Uint64())
}

func (t *EmbedUint) Uint64() uint64 {
	return atomic.LoadUint64(&t.v)
}

type Uint struct {
	EmbedUint
}

func (t *Uint) Store(v string) {
	uv, _ := strconv.ParseUint(v, 10, 64)
	if t.Uint64() == uv {
		return
	}
	atomic.StoreUint64(&t.v, uv)
	CallbackMgr.EvtChan() <- t
}

func (t *Uint) Changed(evt UpdateEvent) {
	CallbackMgr.RegChan() <- &RegisterCallback{
		Val: t,
		Evt: evt,
	}
}

type EmbedFloat struct {
	v atomic.Value
}

func newEmbedFloat(f float64) *EmbedFloat {
	ef := &EmbedFloat{}
	ef.v.Store(f)
	return ef
}

func (t *EmbedFloat) String() string {
	return strconv.FormatFloat(t.Float64(), 'f', 6, 64)
}

func (t *EmbedFloat) Float32() float32 {
	return float32(t.Float64())
}

func (t *EmbedFloat) Float64() float64 {
	v := t.v.Load()
	if v == nil {
		return 0.0
	}
	return v.(float64)
}

type Float struct {
	EmbedFloat
}

func (t *Float) Store(v string) {
	fv, _ := strconv.ParseFloat(v, 64)
	if big.NewFloat(t.Float64()).Cmp(big.NewFloat(fv)) == 0 {
		return
	}
	t.v.Store(fv)
	CallbackMgr.EvtChan() <- t
}

func (t *Float) Changed(evt UpdateEvent) {
	CallbackMgr.RegChan() <- &RegisterCallback{
		Val: t,
		Evt: evt,
	}
}

type EmbedSlice struct {
	v atomic.Value
	r bool
}

func newEmbedSlice(sv []string, recursion bool) *EmbedSlice {
	es := &EmbedSlice{
		r: recursion,
	}
	es.v.Store(sv)
	return es
}

func (t *EmbedSlice) load() []string {
	sv := t.v.Load()
	if sv == nil {
		return []string{}
	}
	return sv.([]string)
}

func (t *EmbedSlice) Len() int {
	return len(t.load())
}

func (t *EmbedSlice) CanSlice() bool {
	return t.r
}

func (t *EmbedSlice) String() string {
	if t.r {
		return strings.Join(t.load(), ";")
	}
	return strings.Join(t.load(), ",")
}

func (t *EmbedSlice) Strings() []*EmbedString {
	sv := t.load()
	es := make([]*EmbedString, 0, len(sv))
	for _, v := range sv {
		es = append(es, newEmbedString(v))
	}
	return es
}

func (t *EmbedSlice) Bools() []*EmbedBool {
	sv := t.load()
	eb := make([]*EmbedBool, 0, len(sv))
	for _, v := range sv {
		bv := 0
		if b, _ := strconv.ParseBool(v); b {
			bv = 1
		}
		eb = append(eb, &EmbedBool{v: int32(bv)})
	}
	return eb
}

func (t *EmbedSlice) Ints() []*EmbedInt {
	sv := t.load()
	ei := make([]*EmbedInt, 0, len(sv))
	for _, v := range sv {
		i, _ := strconv.ParseInt(v, 10, 64)
		ei = append(ei, &EmbedInt{v: i})
	}
	return ei
}

func (t *EmbedSlice) Uints() []*EmbedUint {
	sv := t.load()
	eu := make([]*EmbedUint, 0, len(sv))
	for _, v := range sv {
		u, _ := strconv.ParseUint(v, 10, 64)
		eu = append(eu, &EmbedUint{v: u})
	}
	return eu
}

func (t *EmbedSlice) Floats() []*EmbedFloat {
	sv := t.load()
	ef := make([]*EmbedFloat, 0, len(sv))
	for _, v := range sv {
		f, _ := strconv.ParseFloat(v, 64)
		ef = append(ef, newEmbedFloat(f))
	}
	return ef
}

func (t *EmbedSlice) Slices(i int) *EmbedSlice {
	if !t.r {
		panic("grc: only support two level map/slice")
	}
	sv := t.load()
	if i+1 > len(sv) {
		panic("grc: index out of range")
	}
	return newEmbedSlice(strings.Split(sv[i], ","), false)
}

type Slice struct {
	EmbedSlice
}

func (t *Slice) Store(v string) {
	if t.String() == v {
		return
	}
	sv := strings.Split(v, ";")
	t.v.Store(sv)
	CallbackMgr.EvtChan() <- t
}

func (t *Slice) Changed(evt UpdateEvent) {
	CallbackMgr.RegChan() <- &RegisterCallback{
		Val: t,
		Evt: evt,
	}
}

type EmbedMap struct {
	v atomic.Value
	r bool
}

func newEmbedMap(mv map[string]string, recursion bool) *EmbedMap {
	em := &EmbedMap{
		r: recursion,
	}
	em.v.Store(mv)
	return em
}

func (t *EmbedMap) load() map[string]string {
	mv := t.v.Load()
	if mv == nil {
		return map[string]string{}
	} else {
		return mv.(map[string]string)
	}
}

func (t *EmbedMap) String() string {
	mv := t.load()
	s := make([]string, 0, len(mv))
	for k, v := range mv {
		if v == "" {
			s = append(s, k)
		} else {
			s = append(s, fmt.Sprintf("%s:%s", k, v))
		}
	}
	if t.r {
		return strings.Join(s, ";")
	}
	return strings.Join(s, ",")
}

func (t *EmbedMap) Keys() *EmbedSlice {
	mv := t.load()
	keys := make([]string, 0, len(mv))
	for k := range mv {
		keys = append(keys, k)
	}
	return newEmbedSlice(keys, false)
}

func (t *EmbedMap) StringVal(key string) *EmbedString {
	mv := t.load()
	return newEmbedString(mv[key])
}

func (t *EmbedMap) BoolVal(key string) *EmbedBool {
	mv := t.load()
	if v, ok := mv[key]; ok {
		if b, _ := strconv.ParseBool(v); b {
			return &EmbedBool{v: 1}
		}
	}
	return &EmbedBool{}
}

func (t *EmbedMap) IntVal(key string) *EmbedInt {
	mv := t.load()
	if v, ok := mv[key]; ok {
		i, _ := strconv.ParseInt(v, 10, 64)
		return &EmbedInt{v: i}
	}
	return &EmbedInt{}
}

func (t *EmbedMap) UintVal(key string) *EmbedUint {
	mv := t.load()
	if v, ok := mv[key]; ok {
		u, _ := strconv.ParseUint(v, 10, 64)
		return &EmbedUint{v: u}
	}
	return &EmbedUint{}
}

func (t *EmbedMap) FloatVal(key string) *EmbedFloat {
	mv := t.load()
	if v, ok := mv[key]; ok {
		f, _ := strconv.ParseFloat(v, 64)
		return newEmbedFloat(f)
	}
	return &EmbedFloat{}
}

func (t *EmbedMap) SliceVal(key string) *EmbedSlice {
	if !t.r {
		panic("grc: only support two level map/slice")
	}
	mv := t.load()
	return newEmbedSlice(strings.Split(mv[key], ","), false)
}

func (t *EmbedMap) parse(v, sep string) map[string]string {
	vv := strings.Split(v, sep)
	mv := make(map[string]string, len(vv))
	for _, v := range vv {
		parts := strings.SplitN(v, ":", 2)
		if len(parts) == 1 {
			mv[parts[0]] = ""
		} else {
			mv[parts[0]] = parts[1]
		}
	}
	return mv
}

func (t *EmbedMap) MapVal(key string) *EmbedMap {
	if !t.r {
		panic("grc: only support two level map/slice")
	}
	mv := t.load()
	m := t.parse(mv[key], ",")
	return newEmbedMap(m, false)
}

type Map struct {
	EmbedMap
}

func (t *Map) Store(v string) {
	if t.String() == v {
		return
	}
	mv := t.parse(v, ";")
	t.v.Store(mv)
	CallbackMgr.EvtChan() <- t
}

func (t *Map) Changed(evt UpdateEvent) {
	CallbackMgr.RegChan() <- &RegisterCallback{
		Val: t,
		Evt: evt,
	}
}
