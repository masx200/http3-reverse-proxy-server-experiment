package generic

// MapInterface 是一个泛型映射接口，支持基本的映射操作。
type MapInterface[T comparable, Y any] interface {
	// Clear 清空映射中的所有元素。
	Clear()
	// Delete 从映射中删除指定的键。
	Delete(T)
	// Get 返回指定键的值，如果键不存在，则返回false。
	Get(T) (Y, bool)
	// Set 设置指定键的值。
	Set(T, Y)
	// Has 检查映射中是否存在指定的键。
	Has(T) bool
	// Values 返回映射中所有值的切片。
	Values() []Y
	// Kes 返回映射中所有键的切片。
	Keys() []T
	// Size 返回映射中元素的数量。
	Size() int64
	// Entries 返回映射中所有键值对的切片。
	Entries() []PairInterface[T, Y]
}

type MapImplement[T comparable, Y any] struct {
	data map[T]Y
}

func NewMapImplement[T comparable, Y any](entries ...PairInterface[T, Y]) MapInterface[T, Y] {
	var m = &MapImplement[T, Y]{
		data: make(map[T]Y),
	}
	for _, entry := range entries {
		m.Set(entry.GetFirst(), entry.GetSecond())
	}
	return m
}

func (m *MapImplement[T, Y]) Clear() {
	m.data = make(map[T]Y)
}

func (m *MapImplement[T, Y]) Delete(key T) {
	delete(m.data, key)
}

func (m *MapImplement[T, Y]) Get(key T) (Y, bool) {
	value, ok := m.data[key]
	return value, ok
}

func (m *MapImplement[T, Y]) Set(key T, value Y) {
	m.data[key] = value
}

func (m *MapImplement[T, Y]) Has(key T) bool {
	_, ok := m.data[key]
	return ok
}

func (m *MapImplement[T, Y]) Values() []Y {
	values := make([]Y, 0, len(m.data))
	for _, value := range m.data {
		values = append(values, value)
	}
	return values
}

func (m *MapImplement[T, Y]) Keys() []T {
	keys := make([]T, 0, len(m.data))
	for key := range m.data {
		keys = append(keys, key)
	}
	return keys
}

func (m *MapImplement[T, Y]) Size() int64 {
	return int64(len(m.data))
}

func (m *MapImplement[T, Y]) Entries() []PairInterface[T, Y] {
	entries := make([]PairInterface[T, Y], 0, len(m.data))
	for key, value := range m.data {
		entries = append(entries, NewPairImplement[T, Y](key, value))
	}
	return entries
}
