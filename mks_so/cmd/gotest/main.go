package main

import (
	"fmt"
	"io/ioutil"
	"mks_sql/pkg/processor"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <sql_file> <input_json_file>\n", os.Args[0])
		return
	}

	sqlFile := os.Args[1]
	jsonFile := os.Args[2]

	sqlBytes, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		fmt.Printf("Error reading SQL file: %s\n", err)
		return
	}

	jsonBytes, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("Error reading JSON file: %s\n", err)
		return
	}

	processor.InitConfig()
	result := processor.ProcessSql(string(sqlBytes), string(jsonBytes))
	fmt.Println(result)
}
