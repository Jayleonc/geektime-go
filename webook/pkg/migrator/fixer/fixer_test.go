package fixer

import (
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"testing"
)

func TestName(t *testing.T) {
	strings := []string{"apple", "banana", "cherry"}
	indexedStrings := slice.Map(strings, func(idx int, s string) string {
		return fmt.Sprintf("%d: %s", idx, s)
	})
	fmt.Println(indexedStrings)
}

// 定义映射函数类型
type MapFunc func(int, string) int

// 实现将字符串转换为长度的映射函数
func stringToLength(idx int, s string) int {
	return len(s)
}
