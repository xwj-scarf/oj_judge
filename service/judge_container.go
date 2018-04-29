package judgeServer

import (
    "archive/tar"
    "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
    "github.com/docker/docker/api/types/container"
    "golang.org/x/net/context"
    "bytes"
    "io/ioutil"
	"errors"
	"fmt"
	"os"
	"io"
	"time"
	"sync"
)

type JudgeContainerManager struct{
	manager *JudgeServer
    container_pool map[string]*ClientInfo                //container id --> client
	judge_mutex *sync.RWMutex
	image_name string
	max_docker_num int
}

func (self *JudgeContainerManager) SetMaxDockerNum(num int) {
    self.max_docker_num = num
}

func (self *JudgeContainerManager) SetImageName(image_name string) {
    self.image_name = image_name
}

func (self *JudgeContainerManager) Init() {
    self.judge_mutex = new(sync.RWMutex)
	self.container_pool = make(map[string]*ClientInfo)

    for i:=0;i<self.max_docker_num;i++ {
        respID,cli := self.CreateContainer(self.image_name)
        self.container_pool[respID] = &ClientInfo{
            client: cli,
            is_work:false,
        }
    }
}

func (self *JudgeContainerManager) getClientInfo(containerId string) *client.Client{
	self.judge_mutex.RLock()
	defer self.judge_mutex.RUnlock()
	client_info, ok := self.container_pool[containerId]
	if !ok {
		fmt.Println("get container client error!")
		return nil
	}
	return client_info.client
}

func (self *JudgeContainerManager) ComplieCodeInContainer(containerId string) error{
	return self.ContainerExec(containerId,"docker","sh","/tmp/complie.sh")
}

func (self *JudgeContainerManager) ChangePermission(containerId string,file_path string) error {
	return self.ContainerExec(containerId,"root","chmod","777",file_path)
}

func (self *JudgeContainerManager) RunInContainer(containerId string) error{
	return self.ContainerExec(containerId,"docker","sh","/tmp/do.sh")
}

func (self *JudgeContainerManager) checkContainerInspect(containerId string) bool{
	cli := self.getClientInfo(containerId)
	ctx := context.Background()

	inspect, err := cli.ContainerInspect(ctx,containerId) 
	if err != nil {
		fmt.Println(err)
		return false
	}
	if inspect.State.Running != true || inspect.State.ExitCode != 0 || inspect.State.Error != "" {
		fmt.Println(inspect.State)
		return false
	}
	return true
}

func (self *JudgeContainerManager) restartContainer(containerId string) {
	cli := self.getClientInfo(containerId)
	ctx := context.Background()
	
	err := cli.ContainerRestart(ctx,containerId,nil)
	if err != nil {
		fmt.Println(err)
	}	
}

func (self *JudgeContainerManager) JudgeOutput(containerId string) error{
	return nil
}

func (self *JudgeContainerManager) CreateContainer(image_name string) (string,*client.Client){
	ctx := context.Background()
    cli, err := client.NewEnvClient()
    if err != nil {
        panic(err)
    }

    imageName := image_name
	host_config := &container.HostConfig {
		//单位B
		Resources : container.Resources{
			Memory:500000000,
			CPUCount:1,
			MemorySwap:5000000000,         
			MemoryReservation :500000000,
			//KernelMemoryLimit :500000, 
			//MemorySwap:-1, 
		},	
	}
    resp, err := cli.ContainerCreate(ctx, &container.Config{
        Image: imageName,
        Cmd: []string{"/bin/bash"},
        Tty: true,
        AttachStdout:true, 
        AttachStderr:true,
     }, host_config, nil, "")

    if err != nil {
        panic(err)
    }
    if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
        panic(err)
    }
	return resp.ID,cli
}

func (self *JudgeContainerManager) SendToContainer(file_name, containerId string) error{
	cli := self.getClientInfo(containerId)
	ctx := context.Background()
 
	filePath := self.manager.tmp_path + "/" + containerId + "/" + file_name 
	destPath := "/tmp/"
	fmt.Println(filePath)
	fmt.Println(destPath)
    code,e1 := ioutil.ReadFile(filePath)
    if e1 != nil {
        return e1
    }

    buf_code:=bytes.NewBuffer(code)

    buf0 := new(bytes.Buffer)
    tw := tar.NewWriter(buf0)
    defer tw.Close()

    tarHeader := &tar.Header{
        Name: file_name,
        Size: int64(buf_code.Len()),
    }
    err5 := tw.WriteHeader(tarHeader)
    if err5 != nil {
		return err5
    }
    _, err := tw.Write(buf_code.Bytes())
    if err != nil {
		return err
    }

    tarreader := bytes.NewReader(buf0.Bytes())
	//defer tarreader.Close()
    err1 := cli.CopyToContainer(ctx,containerId,destPath,tarreader, types.CopyToContainerOptions{
            AllowOverwriteDirWithFile:true, 
    })
    if err1 != nil{
		return err1
    }
	return nil
}

//copy out.txt from container 
func (self *JudgeContainerManager) CopyFromContainer(container_id,file_name string) error {
	file_path := "/tmp/" + file_name
	dest_path := self.manager.tmp_path+"/"+container_id+"/"+file_name

	cli := self.getClientInfo(container_id)
	ctx := context.Background()

	fmt.Println(file_path)
    returnoutput,out,err1 := cli.CopyFromContainer(ctx,container_id,file_path)
	if err1 != nil {
		fmt.Println(err1)
		return err1
	}
    defer returnoutput.Close()
	fmt.Println(out)
	file,err2 := os.Create(dest_path)
	if err2 != nil {
		fmt.Println(err2)
		return err2
	}
	defer file.Close()
	tr := tar.NewReader(returnoutput)
	for {
			_, err := tr.Next()
            if err == io.EOF {
                break
            }
            if err != nil {
				fmt.Println(err)
				return err
            }
            buf := new(bytes.Buffer)
            buf.ReadFrom(tr)  
			_, err = io.Copy(file,buf)
			if err != nil {
				fmt.Println(err)
				return err
			}
            //wholeContent := buf.String() 
	}

	return nil
}

func (self *JudgeContainerManager) DelFileInContainer(containerId string) error{
	return self.ContainerExec(containerId,"root","rm","/tmp/input.txt","/tmp/output.txt","/tmp/code.cpp")
}

func (self *JudgeContainerManager) ContainerExec(containerId string, user string,cmd ...string) error {
    cli := self.getClientInfo(containerId)
    if cli == nil {
        return errors.New("get container client error!")
    }
    ctx := context.Background()

	container_exec_create, err := cli.ContainerExecCreate(ctx,containerId,types.ExecConfig{
		Cmd: cmd,
		User: user,
		Detach:false,
	})

	if err != nil {
		fmt.Println(err)
		return err
	}
	
	container_exec_attach, err := cli.ContainerExecAttach(ctx,container_exec_create.ID,types.ExecStartCheck{
		Tty:false,
		Detach:false,
	})
	defer container_exec_attach.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}
	
	timer := time.Now().Unix()
	for {
		exec_info,err:= cli.ContainerExecInspect(ctx,container_exec_create.ID)
		if err != nil {
			fmt.Println(err)
			return err
		}
		if exec_info.Running == true {
			time.Sleep(1*time.Second)
		} else {
			break
		}
		now := time.Now().Unix()
		if now - timer > int64(self.manager.judge_time_out) {
			fmt.Println("ContainerExec time out!")
			return errors.New("ContainerExec time out!")
		}
	}
	return nil
}

