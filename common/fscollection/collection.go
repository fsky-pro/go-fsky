/**
@copyright: fantasysky 2016
@website: https://www.fsky.pro
@brief: collection functions
@author: fanky
@version: 1.0
@date: 2022-05-07
**/

package fscollection

import "slices"

// -------------------------------------------------------------------
// slice utils
// -------------------------------------------------------------------
// 复制一个 slice
func SliceCopy[T any](items []T) []T {
	newItems := make([]T, len(items))
	copy(newItems, items)
	return newItems
}

// 判断两个 slice 中的元素是否完全相等
func SliceEqual[T comparable](items1, items2 []T) bool {
	if len(items1) != len(items2) { 
		return false 
	}
	for idx, item := range items1 {
		if item != items2[idx] {
			return false
		}
	}
	return true
}

// 判断 slice 中是否存在指定元素
func SliceHas[T comparable](items []T, value T) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}

// 判断 slice 中是否存指定属性一致的元素
func SliceHasFunc[T any](items []T, f func(e T) bool) bool {
	for _, e := range items {
		if f(e) {
			return true
		}
	}
	return false
}

// 重组元素
func SliceFunc[E, R any](items []E, f func(e E) R) []R {
	ret := []R{}
	for _, e := range items {
		ret = append(ret, f(e))
	}
	return ret
}

// ---------------------------------------------------------
// 找出指定元素在 slice 中的索引，不存在则返回 -1
func SliceIndexOf[T comparable](items []T, value T) int {
	for index, item := range items {
		if item == value {
			return index
		}
	}
	return -1
}

// ---------------------------------------------------------
// 获取符合条件的元素
func SliceGetsFunc[T any](items []T, f func(T)bool) []T {
	newItems := []T{}
	for _, item := range items {
		if f(item) {
			newItems = append(newItems, item)
		}
	}
	return newItems
}

// 删除指定的子元素
func SliceRemoves[T comparable](items []T, es ...T) []T {
	newItems := []T{}
	for _, item := range items {
		if !slices.Contains(es, item){
			newItems = append(newItems, item)
		}
	}
	return newItems
}

// 根据函数参数返回，删除指定的元素，如果 f 返回 true，则删除
func SliceRemoveFunc[T any](items []T, f func(e T) bool) []T {
	newItems := []T{}
	for _, item := range items {
		if !f(item) {
			newItems = append(newItems, item)
		}
	}
	return newItems
}

// ---------------------------------------------------------
// 获取两个 slice 的交集部分
func SliceIntersection[T comparable](items1 []T, items2 []T) []T {
	items := make([]T, 0)
	for _, item := range items1 {
		if slices.Contains(items2, item) {
			items = append(items, item)
		}
	}
	return items
}

// 获取 items1 中存在，items2 中不存在的集合(即 items2 的补集)
func SliceDifference[T comparable](items1 []T, items2 []T) []T {
	items := make([]T, 0)
	for _, item := range items1 {
		if !slices.Contains(items2, item) {
			items = append(items, item)
		}
	}
	return items
}

// 去除重复的元素
func SliceUnique[T comparable](items []T) []T {
	m := map[T]any{}
	for _, item := range items {
		m[item] = nil
	}
	return MapKeysToSlice(m)
}

// 翻转列表
func SliceReverse[T any](items []T) []T {
	newItems := []T{}
	for i := len(items) - 1; i >= 0; i-- {
		newItems = append(newItems, items[i])
	}
	return newItems
}

// -------------------------------------------------------------------
// map utils
// -------------------------------------------------------------------
// 获取 map 中的值，不存在则返回默认值
func MapGet[K comparable, V any](items map[K]V, key K, def V) V {
	if v, ok := items[key]; ok {
		return v
	}
	return def
}

// 判断 map 中是否存在指定的 key
func MapHasKey[K comparable, V any](items map[K]V, key K) bool {
	_, ok := items[key]
	return ok
}

// 判断 map 中是否存在指定的 value
// 注意：要求 map 的值类型，必须是可比较类型
func MapHasValue[K comparable, V comparable](items map[K]V, value V) bool {
	for _, v := range items {
		if v == value {
			return true
		}
	}
	return false
}

// 用 map m2 更新 map m1
func MapUpdate[K comparable, V any](m1 map[K]V, m2 map[K]V) {
	for k, v := range m2 {
		m1[k] = v
	}
}

// 将 map 的 key 和 value 交换
func MapSwapKeyValue[K comparable, V comparable](m map[K]V) map[V]K {
	vk := map[V]K{}
	for k, v := range m {
		vk[v] = k
	}
	return vk
}

// 将 map 的所有 key 转换为 slice
func MapKeysToSlice[K comparable, V any](m map[K]V) []K {
	keys := []K{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// 将 map 的所有 value 转换为 slice
func MapValuesToSlice[K comparable, V any](m map[K]V) []V {
	values := []V{}
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// 复制一个 map
func MapCopy[K comparable, V any](m map[K]V) map[K]V {
	if m == nil { return nil }
	newMap := map[K]V{}
	for k, v := range m{
		newMap[k] =v
	}
	return m
}
