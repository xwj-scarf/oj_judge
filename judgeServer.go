package main

import (
	"github.com/Unknwon/goconfig"
	"fmt"
	"./service"
)

var (
	max_num int
	image_name string
	tmp_path string
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
	redis_address = cfg.MustValue("Redis","Addr")

	fmt.Println(image_name)
	var server judgeServer.JudgeServer
	server.SetMaxDockerNum(max_num)	
	server.SetImageName(image_name)
	server.SetTmpPath(tmp_path)
	server.SetRedisAddress(redis_address)
	server.Init()
	server.Run()
}
