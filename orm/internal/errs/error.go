package errs

import (
	"errors"
	"fmt"
)

var (
	// ErrPointerOnly 只支持一级指针作为输入
	// 看到这个 error 说明你输入了其它的东西
	// 我们并不希望用户能够直接使用 err == ErrPointerOnly
	// 所以放在我们的 internal 包里
	ErrPointerOnly = errors.New("orm: 只支持指向结构体的一级指针")

	ErrNoRows = errors.New("orm: 没有数据")

	// ErrInsertZeroRow 代表插入 0 行
	ErrInsertZeroRow = errors.New("orm: 插入 0 行")
)

// NewErrUnknownField 返回代表未知字段的错误
// 一般意味着你可能输入的是列名，或者输入了错误的字段名
func NewErrUnknownField(fd string) error {
	return fmt.Errorf("orm: 未知字段 %s", fd)
}

func NewErrUnknownColumn(name string) error {
	return fmt.Errorf("orm: 未知列 %s", name)
}

// NewErrUnsupportedExpressionType 返回一个不支持该 expression 错误信息
func NewErrUnsupportedExpressionType(exp any) error {
	return fmt.Errorf("orm: 不支持的表达式 %v", exp)
}

// 后面可以考虑支持错误码
// func NewErrUnsupportedExpressionType(exp any) error {
// 	return fmt.Errorf("orm-50001: 不支持的表达式 %v", exp)
// }

// 后面还可以考虑用 AST 分析源码，生成错误排除手册，例如
// @ErrUnsupportedExpressionType 40001
// 发生该错误，主要是因为传入了不支持的 Expression 的实际类型
// 一般来说，这是因为中间件

func NewErrInvalidTagContent(pair string) error {
	return fmt.Errorf("orm: 非法标签值 %s", pair)
}

func NewErrUnsupportedAssignable(expr any) error {
	return fmt.Errorf("orm: 不支持的赋值表达式类型 %v", expr)
}

func NewErrFailedToRollbackTx(bizErr error, rbErr error, panicked bool) error {
	return fmt.Errorf("orm: 事务闭包回滚失败，业务错误: %w，回滚错误 %s，"+
		"是否 panic: %t", bizErr, rbErr, panicked)
	// return fmt.Errorf("orm: 事务闭包回滚失败，业务错误: %s，回滚错误 %w，" +
	// 	"是否 panic: %t", bizErr, rbErr, panicked)
}

func NewErrUnsupportedTable(table any) error {
	return fmt.Errorf("orm: 不支持的TableReference类型 %v", table)
}
