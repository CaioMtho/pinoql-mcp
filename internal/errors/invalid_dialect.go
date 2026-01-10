package errors

import (
	"fmt"
	"strings"
)

type InvalidDialectError struct {
	DialectInput  string
	ValidDialects []string
}

func (e InvalidDialectError) Error() string {
	return fmt.Sprintf("Dialect %v is invalid. Valid dialects: %v", e.DialectInput, strings.Join(e.ValidDialects[:], ", "))
}
