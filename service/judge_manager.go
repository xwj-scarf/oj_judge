package judgeServer

import (
	"fmt"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
	"golang.org/x/net/context"
	"os"
	"time"
)

type JudgeServer struct{
	max_docker_num int
	image_name string
	tmp_path	string
	container_pool map[string]*ClientInfo                //container id --> client
	worker JudgeWorker
	RedisServer
}

func (self *JudgeServer) SetTmpPath(path string) {
	self.tmp_path = path
}

func (self *JudgeServer) SetMaxDockerNum(num int) {
	self.max_docker_num = num
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
	self.container_pool = make(map[string]*ClientInfo)

	for i:=0;i<self.max_docker_num;i++ {
	    ctx := context.Background()
    	cli, err := client.NewEnvClient()
    	if err != nil {
        	panic(err)
    	}

    	imageName := self.image_name
    	resp, err := cli.ContainerCreate(ctx, &container.Config{
        	Image: imageName,
        	Cmd: []string{"/bin/bash"},
        	Tty: true,
        	AttachStdout:true, 
     	   	AttachStderr:true,
   		 }, nil, nil, "")

    	if err != nil {
        	panic(err)
    	}
		if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			panic(err)
		}
		fmt.Println(resp.ID)
		self.container_pool[resp.ID] = &ClientInfo{
			client: cli,
			is_work: false,	
		}

		tmp_path := self.tmp_path + "/"+resp.ID
		err = os.MkdirAll(tmp_path,0777)
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
	time.Sleep(1000*time.Second)
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
