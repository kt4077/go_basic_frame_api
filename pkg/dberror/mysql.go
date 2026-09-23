// Package dberror 提供与具体业务无关的数据库错误识别。
package dberror

import (
	"errors"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

func IsDuplicateKey(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
