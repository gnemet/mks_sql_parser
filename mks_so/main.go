package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"mks_sql/pkg/processor"
)

func init() {
	// Config loaded lazily
}

//export mksSql
func mksSql(cSql *C.char, cInput *C.char) *C.char {
	sqlText := C.GoString(cSql)
	inputJSON := C.GoString(cInput)

	result := processor.ProcessSql(sqlText, inputJSON, false)

	return C.CString(result)
}

func main() {
	// Need a main function to make CGO happy
}
