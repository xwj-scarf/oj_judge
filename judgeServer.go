package main

import (
	"github.com/Unknwon/goconfig"
	"fmt"
	"./service"
)

var (
	max_num int
	judge_time_out int
	image_name string
	tmp_path string
	input_path string
	redis_address string
)


func main() {
	cfg, err := goconfig.LoadConfigFile("config.ini")
	if err != nil {
		fmt.Println(err)
	}

	max_num = cfg.MustInt("Docker", "MaxNum")
	image_name = cfg.MustValue("Docker","ImageName")	
	tmp_path = cfg.MustValue("File","TmpPath")	
	input_path = cfg.MustValue("File","InputPath")
	redis_address = cfg.MustValue("Redis","Addr")
	judge_time_out = cfg.MustInt("Judge","TimeOut")

	fmt.Println(image_name)
	var server judgeServer.JudgeServer
	server.SetMaxDockerNum(max_num)	
	server.SetImageName(image_name)
	server.SetTmpPath(tmp_path)
	server.SetInputPath(input_path)
	server.SetRedisAddress(redis_address)
	server.SetJudgeTimeOut(judge_time_out)
	server.Init()
	server.Run()
}
