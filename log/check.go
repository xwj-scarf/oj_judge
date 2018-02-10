package main

import (
	"io/ioutil"
	"fmt"
	"strings"
)

func main() {
        b, err := ioutil.ReadFile("4.log")
        if err != nil {
            fmt.Print(err)
        }
        str := string(b)
        fmt.Println(strings.Count(str,`ce.txt 159 -rw-r--r--`))
}
