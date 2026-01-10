package connection

import (
	"strconv"

	"github.com/CaioMtho/pinoql-mcp/internal/errors"
	_ "github.com/CaioMtho/pinoql-mcp/internal/errors"
)

type Config struct {
	Dialect  Dialect
	DSN      string
	ReadOnly bool
}

func FromRaw(dialectString string, dsn string, readonlyString string) (*Config, error) {
	if !IsValidDialect(dialectString) {
		return nil, &errors.InvalidDialectError{DialectInput: dialectString, ValidDialects: GetDialects()}
	}

	dialect := Dialect(dialectString)

	readonly := true

	if readonlyString != "" {
		if b, err := strconv.ParseBool(readonlyString); err != nil {
			readonly = b
		}
	}

	config := &Config{Dialect: dialect, DSN: dsn, ReadOnly: readonly}
	return config, nil
}
