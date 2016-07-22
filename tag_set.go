package tagcache

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

type TagSet struct {
	store CacheStore
	names []string
}

func NewTagSet(store CacheStore, names []string) *TagSet {

	t := &TagSet{store, names}
	t.SetNames(names)
	return t
}

func (this *TagSet) SetNames(names []string) {
	this.names = names
}

func (this *TagSet) AddNames(names []string) {
	names = append(this.names, names...)
	m := make(map[string]struct{})

	for i, l := 0, len(names); i < l; i++ {
		m[names[i]] = struct{}{}
	}

	filterNames := make([]string, len(m))
	i := 0
	for k, _ := range m {
		filterNames[i] = k
		i++
	}

	this.names = filterNames
}

// 刷新所有 tag key
func (this *TagSet) Reset() error {
	for _, name := range this.names {
		this.ResetTag(name)
	}
	return nil
}

// 取tag id
func (this *TagSet) TagId(name string) string {
	id := this.store.Get(this.TagKey(name))
	if len(id) == 0 {
		return this.ResetTag(name)
	}

	return id
}

// 取所有的tagid
func (this *TagSet) TagIds() []string {
	l := len(this.names)
	if l == 0 {
		return []string{}
	}

	//  排序
	sort.Strings(this.names)

	ids := make([]string, l)
	for i, name := range this.names {
		ids[i] = this.TagId(name)
	}

	return ids
}

// 取命名空间
func (this *TagSet) GetNamespace() string {
	ids := this.TagIds()
	if len(ids) == 0 {
		return ""
	}
	return strings.Join(ids, "|")
}

// 重置tagID 版本+1
func (this *TagSet) ResetTag(name string) string {

	seq, _ := this.store.Incr(this.TagKey(name))

	// redis 起过最大值不会归零
	if seq > (math.MaxInt64 - 10) {
		seq = 0
		this.store.Set(this.TagKey(name), "0", 0)
	}

	return strconv.FormatInt(seq, 10)
}

// Tag key
func (this *TagSet) TagKey(name string) string {
	return fmt.Sprintf("tag:%s:key", name)
}
