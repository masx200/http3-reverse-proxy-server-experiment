package generic

import (
	"math/rand"
	"time"
)

// RandomShuffle 函数用于对指定类型的切片进行随机打乱。
//
// 参数:
// arr []T: 待打乱顺序的切片。
//
// 返回值:
// []T: 打乱顺序后的切片。
func RandomShuffle[T any](arr []T) []T {
	// 使用当前时间的纳秒级种子初始化随机数生成器，以确保每次运行结果都不同。
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 使用 rand.Shuffle 函数来随机打乱切片的顺序。
	// 这个函数会传入切片的长度以及一个交换元素的函数。
	r.Shuffle(len(arr), func(i, j int) {
		// 交换函数通过交换 arr[i] 和 arr[j] 来打乱顺序。
		arr[i], arr[j] = arr[j], arr[i]
	})
	return arr
}
