// 仓储公共辅助：唯一约束错误映射。
package store

import (
	"fmt"
	"strings"

	"task267-choreoreview/internal/model"
)

// mapUniqueErr 将 SQLite 唯一约束错误映射为 ErrDuplicate。
func mapUniqueErr(err error, msg string) error {
	if err == nil {
		return nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "unique") ||
		strings.Contains(strings.ToLower(err.Error()), "constraint") {
		return fmt.Errorf("%w: %s", model.ErrDuplicate, msg)
	}
	return err
}
