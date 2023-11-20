package main

import (
	"errors"
	"fmt"
)

func main() {
	fmt.Println(SumNumber([]int{1, 2, 3}))
	fmt.Println(Max([]float64{23.12, 22.3, 131.21}))
	fmt.Println(Min([]float64{23.12, 2.3, 131.21}))
	fmt.Println(Filter([]int{24, 23, 13121, 13, 53, 658, 2, 3, 5, 19}, func(i int) bool { return i%2 == 0 }))
	val, index, err := Find(2, []float64{23.12, 2.3, 131.21, 2})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v : %v \n", index, val)
}

type Number interface {
	~int | uint | int32 | float64
}

// 使用泛型求和
func SumNumber[T Number](vals []T) T {
	var res T
	for _, v := range vals {
		res = res + v
	}
	return res
}

// 使用泛型求最大值和最小值
func Max[T Number](vals []T) T {
	max := vals[0]

	for _, v := range vals {
		if v > max {
			max = v
		}
	}
	return max
}

func Min[T Number](vals []T) T {
	min := vals[0]

	for _, v := range vals {
		if v < min {
			min = v
		}
	}
	return min
}

// 过滤出符合条件的切片
func Filter[T Number](vals []T, c func(T) bool) []T {
	var res []T

	for _, v := range vals {
		if c(v) {
			res = append(res, v)
		}
	}
	return res
}

// 查找
func Find[T Number](val T, vals []T) (T, int, error) {
	for i, v := range vals {
		if val == v {
			return val, i, nil
		}
	}
	return 0, 0, errors.New("没有找到该元素")
}
