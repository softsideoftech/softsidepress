package softmail

import (
	_ "github.com/lib/pq"
	"net/http"
	//"database/sql"
	//"log"
	"fmt"
	"database/sql"
	"log"
)


func QueryDb(w http.ResponseWriter, r *http.Request) {
	connStr := "user=softside dbname=softside sslmode=disable password=local_dev_passwordasdf"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("SELECT * FROM list_members limit 60")

	if err != nil {
		log.Println("Error: %v", err)
	}
	println("HELLO WORLD!!!")


	cols, err := rows.Columns()
	if err != nil {
		fmt.Println("Failed to get columns", err)
		return
	}

	// Result is your slice string.
	rawResult := make([][]byte, len(cols))
	result := make([]string, len(cols))

	dest := make([]interface{}, len(cols)) // A temporary interface{} slice
	for i, _ := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			fmt.Println("Failed to scan row", err)
			return
		}

		for i, raw := range rawResult {
			if raw == nil {
				result[i] = "\\N"
			} else {
				result[i] = string(raw)
			}
		}

		fmt.Printf("%#v\n", result)
	}


}



