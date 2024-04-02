package generic

// MapInterface 是一个泛型映射接口，支持基本的映射操作。
type MapInterface[T comparable, Y any] interface {
	Iterable[PairInterface[T, Y]]
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
	ForEach(func(Y, T, MapInterface[T, Y]))
}

func (m *MapImplement[T, Y]) ForEach(callback func(Y, T, MapInterface[T, Y])) {
	for key, value := range m.data {
		callback(value, key, m)
	}
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
func MapImplementFromMap[T comparable, Y any](entries map[T]Y) MapInterface[T, Y] {
	var m = &MapImplement[T, Y]{
		data: make(map[T]Y),
	}
	for k, entry := range entries {
		m.Set(k, entry)
	}
	return m
}

type MapIterator[T comparable, Y any] struct {
	entries []PairInterface[T, Y]
	index   int64
	size    int64
}

// Next implements Iterator.
func (m *MapIterator[T, Y]) Next() (PairInterface[T, Y], bool) {

	if m.index < m.size {
		entry := m.entries[m.index]
		m.index++
		return entry, true
	}
	return nil, false
}

func (m *MapImplement[T, Y]) Iterator() Iterator[PairInterface[T, Y]] {
	return &MapIterator[T, Y]{entries: m.Entries(), size: m.Size(), index: 0}
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
