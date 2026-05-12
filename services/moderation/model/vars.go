package model

import (
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var ErrNotFound = sqlx.ErrNotFound

func joinWhere(where []string) string {
	return strings.Join(where, " and ")
}
