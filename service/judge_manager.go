package judgeServer

import (
	"fmt"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/client"
	"golang.org/x/net/context"
	"os"
)

type JudgeServer struct{
	worker JudgeWorker
	RedisServer
	mysql judgeMysql
	judge_time_out int
	JudgeContainerManager
	JudgeFileManager
}

func (self *JudgeServer) SetMysqlInfo(mysqlinfo MysqlInfo) {
	self.mysql.Init(mysqlinfo)
}

func (self *JudgeServer) SetJudgeTimeOut(time_out int) {
	self.judge_time_out = time_out
}

func (self *JudgeServer) Init() {
	self.JudgeContainerManager.manager = self
	self.JudgeFileManager.manager = self
	self.worker.manager = self
	self.mysql.manager = self
	self.JudgeContainerManager.Init()
	self.JudgeFileManager.Init()
	self.RedisInit()
}

func (self *JudgeServer) Run() {
	defer self.Stop()

	for k,v := range self.container_pool {
		fmt.Println(k)
		fmt.Println(v)
	}
	self.worker.Run()
}


//when process done,kill containers
func (self *JudgeServer)Stop() {
	self.mysql.Stop()
    ctx := context.Background()

	cli, err := client.NewEnvClient()
    if err != nil {
        panic(err)
    }

    containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
    if err != nil {
        panic(err)
    }

    for _, container := range containers {
        fmt.Print("Stopping container ", container.ID[:10], "... ")
        if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
            panic(err)
        }
        fmt.Println("Success")
    }
	os.RemoveAll(self.tmp_path)
}
