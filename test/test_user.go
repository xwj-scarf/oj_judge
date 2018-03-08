package main

import (
	"fmt"
	"net/http"
)

func main() {
	res ,err := http.Get("http://192.168.60.128/oj_web/")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(res)
				
}
