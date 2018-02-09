package judgeServer

import (
	"fmt"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/client"
	"golang.org/x/net/context"
	"os"
)

type JudgeServer struct{
	max_docker_num int
	image_name string
	tmp_path	string
	input_path  string
	container_pool map[string]*ClientInfo                //container id --> client
	worker JudgeWorker
	RedisServer
	mysql judgeMysql
	judge_time_out int
}

func (self *JudgeServer) SetTmpPath(path string) {
	self.tmp_path = path
}

func (self *JudgeServer) SetInputPath(path string) {
	self.input_path = path
}

func (self *JudgeServer) SetMaxDockerNum(num int) {
	self.max_docker_num = num
}

func (self *JudgeServer) SetJudgeTimeOut(time_out int) {
	self.judge_time_out = time_out
}

func (self *JudgeServer) SetImageName(image_name string) {
	self.image_name = image_name
}

func (self *JudgeServer) Init() {
	self.RedisInit()
}

func (self *JudgeServer) Run() {
	defer self.Stop()
	self.worker.Manager = self
	self.mysql.Manager = self
	self.container_pool = make(map[string]*ClientInfo)

	for i:=0;i<self.max_docker_num;i++ {
		respID,cli := self.CreateContainer(self.image_name)
		self.container_pool[respID] = &ClientInfo{
			client: cli,
			is_work:false,
		}

		tmp_path := self.tmp_path + "/"+respID
		err := os.MkdirAll(tmp_path,0777)
		if err != nil {
			fmt.Println("create tmp_path error!")
			return
		} 
	}

	for k,v := range self.container_pool {
		fmt.Println(k)
		fmt.Println(v)
	}
	self.worker.Run()
}


//when process done,kill containers
func (self *JudgeServer)Stop() {
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
