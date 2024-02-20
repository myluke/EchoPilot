package helper

// EqualsConstraint 约束指定了类型T必须能够与自己进行等值比较。
type EqualsConstraint[T comparable] interface {
}

// ValueInSlice 函数检查值v是否存在于切片slice中。
// 这里使用了泛型，允许对任意可比较的类型进行操作。
func ValueInSlice[T comparable](v T, slice []T) bool {
	for _, s := range slice {
		if v == s {
			return true
		}
	}
	return false
}
