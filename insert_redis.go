package main

import (
	"os/exec"
	"math/rand"
	"time"
)

func main() {

	for {
		op := rand.Intn(3)
		var judge string
		if op == 0 {
			judge = "./judge_redis_ac"
		} 
		if op == 1 {
			judge = "./judge_redis_ce"
		}
		if op == 2 {
			judge = "./judge_redis_wa"
		}
		for i:=0;i<5;i++ {
			cmd := exec.Command(judge)
			cmd.Run()
		}
		time.Sleep(1*time.Second)
	}
}
