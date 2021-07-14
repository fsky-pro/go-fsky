/**
@copyright: fantasysky 2016
@website: https://www.fsky.pro
@brief: object reader
@author: fanky
@version: 1.0
@date: 2021-06-18
**/

// 将解释过的结构对象，缓存到内存中，使得每次 sql 查询时，不再反射对象

package fsmysql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

var _colTags = []string{"mysql", "db"}
var _colTypeTags = []string{"mysqltd", "dbtd"}

// -------------------------------------------------------------------
// object member
// -------------------------------------------------------------------
// 结构成员解构信息
type s_ObjMember struct {
	dbkey  string       // 在数据库中的关键字
	vtype  reflect.Type // 类型
	offset uintptr      // 字段偏移值
}

// pobj 必须是指针，并且不能为 nil
func (this *s_ObjMember) Value(pobj interface{}) interface{} {
	p := unsafe.Pointer(this.offset + reflect.ValueOf(pobj).Elem().UnsafeAddr())
	return reflect.NewAt(this.vtype, p).Elem().Interface()
}

// -------------------------------------------------------------------
// object members
// -------------------------------------------------------------------
// {成员名称: *s_ObjMember}
type s_ObjInfo map[string]*s_ObjMember

func newObjInfo(tobj reflect.Type) s_ObjInfo {
	members := map[string]*s_ObjMember{}
L:
	for i := 0; i < tobj.NumField(); i++ {
		tfield := tobj.Field(i)
		dbkey := ""
		for _, tag := range _colTags {
			dbkey = tfield.Tag.Get(tag)
			if dbkey == "-" {
				// “-” 为，排除标记，表示该成员不映射到数据库
				continue L
			}
			if dbkey != "" {
				break
			}
		}
		// 如果没有 mysql tag，则将成员名称作为数据库列名
		if dbkey == "" {
			dbkey = tfield.Name
		}
		members[tfield.Name] = &s_ObjMember{dbkey, tfield.Type, tfield.Offset}
	}
	return members
}

// 获取 obj 成员的值，vobj = reflect.ValueOf(obj)
func (this s_ObjInfo) dbkeyMapMembers(vobj reflect.Value, mnames []string) (map[string]interface{}, error) {
	vs := make(map[string]interface{})
	vobj = vobj.Elem()
	if len(mnames) == 0 {
		for _, info := range this {
			vs[info.dbkey] = reflect.NewAt(info.vtype, unsafe.Pointer(vobj.UnsafeAddr()+info.offset)).Elem().Interface()
		}
		return vs, nil
	}
	for _, mname := range mnames {
		mname = strings.TrimSpace(mname)
		info, ok := this[mname]
		if !ok {
			return map[string]interface{}{}, fmt.Errorf("%q is not the member name of %v", mname, vobj.Type())
		}
		vs[info.dbkey] = reflect.NewAt(info.vtype, unsafe.Pointer(vobj.UnsafeAddr()+info.offset)).Elem().Interface()
	}
	return vs, nil
}

// 获取 obj 成员的指针，vobj = reflect.ValueOf(obj)
func (this s_ObjInfo) dbkeyMapPMembers(vobj reflect.Value, mnames []string) (map[string]interface{}, error) {
	ps := make(map[string]interface{})
	vobj = vobj.Elem()
	if len(mnames) == 0 {
		for _, info := range this {
			ps[info.dbkey] = reflect.NewAt(info.vtype, unsafe.Pointer(vobj.UnsafeAddr()+info.offset)).Interface()
		}
		return ps, nil
	}
	for _, mname := range mnames {
		mname = strings.TrimSpace(mname)
		info, ok := this[mname]
		if !ok {
			return map[string]interface{}{}, fmt.Errorf("%q is not the member name of %v", mname, vobj.Type())
		}
		ps[info.dbkey] = reflect.NewAt(info.vtype, unsafe.Pointer(vobj.UnsafeAddr()+info.offset)).Interface()
	}
	return ps, nil
}

// -------------------------------------------------------------------
// object cache
// -------------------------------------------------------------------
type s_ObjCache struct {
	sync.RWMutex
	colTags     []string
	colTypeTags []string
	objs        map[string]s_ObjInfo
}

// tobj = reflect.ValueOf(obj).Elem().Type()
func (this *s_ObjCache) get(tobj reflect.Type) s_ObjInfo {
	this.RLock()
	defer this.RUnlock()
	key := tobj.PkgPath() + ":" + tobj.String()
	if ms, ok := this.objs[key]; ok {
		return ms
	}
	return nil
}

// 解释并添加一个对象到缓存，tobj = reflect.ValueOf(obj).Elem().Type()
func (this *s_ObjCache) add(tobj reflect.Type) s_ObjInfo {
	this.Lock()
	defer this.Unlock()
	objInfo := newObjInfo(tobj)
	this.objs[tobj.PkgPath()+":"+tobj.String()] = objInfo
	return objInfo
}

// 清除缓存
func (this *s_ObjCache) clear() {
	this.Lock()
	defer this.Unlock()
	this.objs = map[string]s_ObjInfo{}
}

func (this *s_ObjCache) resetDBTags(tags []string) {
	this.Lock()
	defer this.Unlock()
	_colTags = tags
	this.objs = map[string]s_ObjInfo{}
}

func (this *s_ObjCache) resetDBTypeTags(tags []string) {
	this.Lock()
	defer this.Unlock()
	_colTypeTags = tags
	this.objs = map[string]s_ObjInfo{}
}

// 对象解构信息缓存
var _objcache = &s_ObjCache{
	objs: make(map[string]s_ObjInfo),
}

// -------------------------------------------------------------------
// package private
// -------------------------------------------------------------------
// 获取结构体成员 tag 与成员值或成员指针指针的映射：{tag:member ptr}
// 如果 ptr 参数为 true，则取成员的指针作为返回 map 的 value，否则取成员的值
func dbkeyMapValues(obj interface{}, members string, ptr bool) (tagMems map[string]interface{}, err error) {
	tagMems = make(map[string]interface{})
	tobj := reflect.TypeOf(obj)
	// 无类型 nil
	if tobj == nil {
		err = errors.New("object must be a not nil type value.")
		return
	}

	// 非指针
	if tobj.Kind() != reflect.Ptr {
		err = errors.New("object type must be a pointer of struct.")
		return
	}
	tobj = tobj.Elem()

	// 必须是结构体
	if tobj.Kind() != reflect.Struct {
		err = errors.New("object must be a not nil struct object pointer.")
		return
	}

	// 必须是非匿名结构体
	if tobj.Name() == "" {
		err = errors.New("object type must be a unanonymous struct.")
		return
	}

	vobj := reflect.ValueOf(obj)
	// 指针指向 nil
	if vobj.IsNil() {
		err = errors.New("object is not allow to be a nil value.")
		return
	}

	members = strings.TrimSpace(members)
	if members == "*" {
		members = members[1:]
	}
	// 是否获取所有字段的值
	all := members == ""
	mnames := make([]string, 0)
	if !all {
		// 指定部分字段，逗号隔开
		mnames = strings.Split(members, ",")
	}

	objInfo := _objcache.get(tobj)
	if objInfo == nil {
		objInfo = _objcache.add(tobj)
	}
	if ptr {
		return objInfo.dbkeyMapPMembers(vobj, mnames)
	}
	return objInfo.dbkeyMapMembers(vobj, mnames)
}

// 获取结构体对应的数据库字段和类型
// 通常创建表格只会调用一次，因此，不作缓存处理
func dbkeyTypes(obj interface{}) ([][2]string, error) {
	tobj := reflect.TypeOf(obj)
	// 无类型 nil
	if tobj == nil {
		return nil, errors.New("object mustn't be a nil type value.")
	}
	if tobj.Kind() == reflect.Ptr {
		tobj = tobj.Elem()
	}
	// 必须是结构体
	if tobj.Kind() != reflect.Struct {
		return nil, errors.New("object must be a struct object or a pointer of struct object.")
	}

	keyTypes := make([][2]string, 0)
	for i := 0; i < tobj.NumField(); i++ {
		field := tobj.Field(i)
		td := ""
		for _, tag := range _colTypeTags {
			td = field.Tag.Get(tag)
			if td != "" {
				break
			}
		}
		if td == "" {
			continue
		}

		key := ""
		for _, tag := range _colTags {
			key = field.Tag.Get(tag)
			if key != "" {
				break
			}
		}
		if key == "" {
			key = field.Name
		}
		keyTypes = append(keyTypes, [2]string{key, td})
	}
	return keyTypes, nil
}

// -------------------------------------------------------------------
// package public
// -------------------------------------------------------------------
// 清除缓存
func ClearObjectCache() {
	_objcache.clear()
}

// 重新设置 db tag
func ResetDBTags(tags ...string) {
	_objcache.resetDBTags(tags)
}

// 重新设置 db type tag
func ResetDBTypeTags(tags ...string) {
	_objcache.resetDBTypeTags(tags)
}
