// Copyright 2018 gf Author(https://gitee.com/johng/gf). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://gitee.com/johng/gf.

package garray

import (
    "gitee.com/johng/gf/g/container/gtype"
    "gitee.com/johng/gf/g/container/internal/rwmutex"
)

// 默认按照从低到高进行排序
type SortedArray struct {
    mu          *rwmutex.RWMutex             // 互斥锁
    cap         int                          // 初始化设置的数组容量
    array       []interface{}                // 底层数组
    unique      *gtype.Bool                  // 是否要求不能重复
    compareFunc func(v1, v2 interface{}) int // 比较函数，返回值 -1: v1 < v2；0: v1 == v2；1: v1 > v2
}

func NewSortedArray(cap int, compareFunc func(v1, v2 interface{}) int, safe...bool) *SortedArray {
    return &SortedArray{
        mu          : rwmutex.New(safe...),
        unique      : gtype.NewBool(),
        array       : make([]interface{}, 0, cap),
        compareFunc : compareFunc,
    }
}

// 添加加数据项
func (a *SortedArray) Add(values...interface{}) {
    if len(values) == 0 {
        return
    }
    a.mu.Lock()
    defer a.mu.Unlock()
    for _, value := range values {
        index, cmp := a.binSearch(value, false)
        if a.unique.Val() && cmp == 0 {
            continue
        }
        if index < 0 {
            a.array = append(a.array, value)
            continue
        }
        // 加到指定索引后面
        if cmp > 0 {
            index++
        }
        rear   := append([]interface{}{}, a.array[index : ]...)
        a.array = append(a.array[0 : index], value)
        a.array = append(a.array, rear...)
    }
}

// 获取指定索引的数据项, 调用方注意判断数组边界
func (a *SortedArray) Get(index int) interface{} {
    a.mu.RLock()
    defer a.mu.RUnlock()
    value := a.array[index]
    return value
}

// 删除指定索引的数据项, 调用方注意判断数组边界
func (a *SortedArray) Remove(index int) interface{} {
    a.mu.Lock()
    defer a.mu.Unlock()
    // 边界删除判断，以提高删除效率
    if index == 0 {
        value  := a.array[0]
        a.array = a.array[1 : ]
        return value
    } else if index == len(a.array) - 1 {
        value  := a.array[index]
        a.array = a.array[: index]
        return value
    }
    // 如果非边界删除，会涉及到数组创建，那么删除的效率差一些
    value  := a.array[index]
    a.array = append(a.array[ : index], a.array[index + 1 : ]...)
    return value
}

// 将最左端(索引为0)的数据项移出数组，并返回该数据项
func (a *SortedArray) PopLeft() interface{} {
    a.mu.Lock()
    defer a.mu.Unlock()
    value  := a.array[0]
    a.array = a.array[1 : ]
    return value
}

// 将最右端(索引为length - 1)的数据项移出数组，并返回该数据项
func (a *SortedArray) PopRight() interface{} {
    a.mu.Lock()
    defer a.mu.Unlock()
    index  := len(a.array) - 1
    value  := a.array[index]
    a.array = a.array[: index]
    return value
}

// 数组长度
func (a *SortedArray) Len() int {
    a.mu.RLock()
    length := len(a.array)
    a.mu.RUnlock()
    return length
}

// 返回原始数据数组
func (a *SortedArray) Slice() []interface{} {
    array := ([]interface{})(nil)
    if a.mu.IsSafe() {
        a.mu.RLock()
        array = make([]interface{}, len(a.array))
        for k, v := range a.array {
            array[k] = v
        }
        a.mu.RUnlock()
    } else {
        array = a.array
    }
    return array
}

// 查找指定数值的索引位置，返回索引位置(具体匹配位置或者最后对比位置)及查找结果
// 返回值: 最后比较位置, 比较结果
func (a *SortedArray) Search(value interface{}) (index int, result int) {
    return a.binSearch(value, true)
}

// 查找指定数值的索引位置，返回索引位置(具体匹配位置或者最后对比位置)及查找结果
// 返回值: 最后比较位置, 比较结果
func (a *SortedArray) binSearch(value interface{}, lock bool)(index int, result int) {
    if len(a.array) == 0 {
        return -1, -2
    }
    if lock {
        a.mu.RLock()
        defer a.mu.RUnlock()
    }
    min := 0
    max := len(a.array) - 1
    mid := 0
    cmp := -2
    for {
        mid = int((min + max) / 2)
        cmp = a.compareFunc(value, a.array[mid])
        switch cmp {
            case -1 : max = mid - 1
            case  0 :
            case  1 : min = mid + 1
        }
        if cmp == 0 || min >= max {
            break
        }
    }
    return mid, cmp
}

// 设置是否允许数组唯一
func (a *SortedArray) SetUnique(unique bool) {
    oldUnique := a.unique.Val()
    a.unique.Set(unique)
    if unique && oldUnique != unique {
        a.doUnique()
    }
}

// 清理数组中重复的元素项
func (a *SortedArray) doUnique() {
    a.mu.Lock()
    defer a.mu.Unlock()
    i := 0
    for {
        if i == len(a.array) - 1 {
            break
        }
        if a.compareFunc(a.array[i], a.array[i + 1]) == 0 {
            a.array = append(a.array[ : i + 1], a.array[i + 1 + 1 : ]...)
        } else {
            i++
        }
    }
}

// 清空数据数组
func (a *SortedArray) Clear() {
    a.mu.Lock()
    if len(a.array) > 0 {
        a.array = make([]interface{}, 0, a.cap)
    }
    a.mu.Unlock()
}

// 使用自定义方法执行加锁修改操作
func (a *SortedArray) LockFunc(f func(array []interface{})) {
    a.mu.Lock(true)
    defer a.mu.Unlock(true)
    f(a.array)
}

// 使用自定义方法执行加锁读取操作
func (a *SortedArray) RLockFunc(f func(array []interface{})) {
    a.mu.RLock(true)
    defer a.mu.RUnlock(true)
    f(a.array)
}