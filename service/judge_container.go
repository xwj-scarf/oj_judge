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
)

func (self *JudgeServer) ComplieCodeInContainer(containerId string) error{
	self.judge_mutex.RLock()
	client_info, ok := self.container_pool[containerId]	
	if !ok {
		self.judge_mutex.RUnlock()
		return errors.New("get container client error!")
	}
	self.judge_mutex.RUnlock()
	cli := client_info.client
	ctx := context.Background()
 
    respexec,err := cli.ContainerExecCreate(ctx,containerId,types.ExecConfig{
        Cmd: []string{"sh","/tmp/complie.sh"},
		Detach:false,
    })

    if err != nil {
		fmt.Println(err)
        return err
    }

    respexecruncode,err1 := cli.ContainerExecAttach(ctx,respexec.ID,types.ExecStartCheck{
        Tty:false,
		Detach:false,
		
    })
	defer respexecruncode.Close() 
	if err1 != nil {
        fmt.Println(err1)
		return err1
    }

	timer := time.Now().Unix()
	for {
		execInfo,err:= cli.ContainerExecInspect(ctx,respexec.ID)
		if err != nil {
			fmt.Println(err)
			return err
		}
		if execInfo.Running == true {
			time.Sleep(1*time.Second)
		} else {
			break
		}
		now := time.Now().Unix()
		if now - timer > int64(self.judge_time_out) {
			return errors.New("complie time out!")
		}
	}
	return nil
}

func (self *JudgeServer) ChangePermission(containerId string,file_path string) error {
	self.judge_mutex.RLock()
	client_info, ok := self.container_pool[containerId]
	if !ok {
		self.judge_mutex.RUnlock()
		return errors.New("get container client error")
	}
	self.judge_mutex.RUnlock()
	cli := client_info.client
	ctx := context.Background()

    respexecruncode,err := cli.ContainerExecCreate(ctx,containerId,types.ExecConfig{
		Detach:false,
		User: "root",
        Cmd: []string{"chmod","777",file_path},
    })
    if err != nil {
        fmt.Println(err)
        return err
    }
	
    resprunexecruncode,err := cli.ContainerExecAttach(ctx,respexecruncode.ID,types.ExecStartCheck{
        Tty:false,
		Detach: false,
    })
	defer resprunexecruncode.Close()
    if err != nil {
        fmt.Println(err)
		return err
    }
	timer := time.Now().Unix()
	for {
		execInfo,err:= cli.ContainerExecInspect(ctx,respexecruncode.ID)
		if err != nil {
			return err
		}
		if execInfo.Running == true {
			time.Sleep(1*time.Second)
		} else {
			break
		}
		now := time.Now().Unix()
		if now - timer > int64(self.judge_time_out) {
			self.restartContainer(containerId)
			return errors.New("change permission error")
		}
	}
	return nil
}

func (self *JudgeServer) RunInContainer(containerId string) error{
	self.judge_mutex.RLock()
	client_info, ok := self.container_pool[containerId]
	if !ok {
		self.judge_mutex.RUnlock()
		return errors.New("get container client error")
	}
	self.judge_mutex.RUnlock()
	cli := client_info.client
	ctx := context.Background()

    respexecruncode,err := cli.ContainerExecCreate(ctx,containerId,types.ExecConfig{
		Detach:false,
        Cmd: []string{"sh","/tmp/do.sh"},
    })
    if err != nil {
        fmt.Println(err)
        return err
    }
	
    resprunexecruncode,err := cli.ContainerExecAttach(ctx,respexecruncode.ID,types.ExecStartCheck{
        Tty:false,
		Detach: false,
    })
	defer resprunexecruncode.Close()
    if err != nil {
        fmt.Println(err)
		return err
    }
	timer := time.Now().Unix()
	for {
		execInfo,err:= cli.ContainerExecInspect(ctx,respexecruncode.ID)
		if err != nil {
			return err
		}
		if execInfo.Running == true {
			time.Sleep(1*time.Second)
		} else {
			break
		}
		now := time.Now().Unix()
		if now - timer > int64(self.judge_time_out) {
			self.restartContainer(containerId)
			return errors.New("run code time out!")
		}
	}
	return nil
}

func (self *JudgeServer) checkContainerInspect(containerId string) bool{
	self.judge_mutex.RLock()
	client_info,_ := self.container_pool[containerId]
	self.judge_mutex.RUnlock()
	cli := client_info.client
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

func (self *JudgeServer) restartContainer(containerId string) {
	self.judge_mutex.RLock()
	client_info,_ := self.container_pool[containerId]
	self.judge_mutex.RUnlock()
	cli := client_info.client
	ctx := context.Background()
	
	err := cli.ContainerRestart(ctx,containerId,nil)
	if err != nil {
		fmt.Println(err)
	}	
}

func (self *JudgeServer) JudgeOutput(containerId string) error{
	return nil
}

func (self *JudgeServer) CreateContainer(image_name string) (string,*client.Client){
	ctx := context.Background()
    cli, err := client.NewEnvClient()
    if err != nil {
        panic(err)
    }

    imageName := image_name
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
	return resp.ID,cli
}

func (self *JudgeServer) SendToContainer(file_name, containerId string) error{
	self.judge_mutex.RLock()
	client_info, ok := self.container_pool[containerId]		
	if !ok {
		self.judge_mutex.RUnlock()
		return errors.New("get container client error!")
	}
	self.judge_mutex.RUnlock()
	cli := client_info.client
	ctx := context.Background()
 
	filePath := self.tmp_path + "/" + containerId + "/" + file_name 
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
func (self *JudgeServer) CopyFromContainer(container_id,file_name string) error {
	file_path := "/tmp/" + file_name
	dest_path := self.tmp_path+"/"+container_id+"/"+file_name

	self.judge_mutex.RLock()
	client_info, ok := self.container_pool[container_id]		
	if !ok {
		self.judge_mutex.RUnlock()
		return errors.New("get container client error!")
	}
	self.judge_mutex.RUnlock()
	cli := client_info.client
	ctx := context.Background()

	fmt.Println(file_path)
    returnoutput,out,err1 := cli.CopyFromContainer(ctx,container_id,file_path)
	if err1 != nil {
		fmt.Println(err1)
		return err1
	}
    defer returnoutput.Close()
	fmt.Println(out)
	//if file_name == "ce.txt" {
	//	if out.Size == 0 {
	//		return nil
	//	}
	//}
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

func (self *JudgeServer) DelFileInContainer(containerId string) error{
	self.judge_mutex.RLock()
	client_info, ok := self.container_pool[containerId]	
	if !ok {
		self.judge_mutex.RUnlock()
		return errors.New("get container client error!")
	}
	self.judge_mutex.RUnlock()
	cli := client_info.client
	ctx := context.Background()
 
    respexec,err := cli.ContainerExecCreate(ctx,containerId,types.ExecConfig{
        Cmd: []string{"rm","/tmp/input.txt /tmp/output.txt /tmp/code.cpp"},
		User: "root",
		Detach:false,
    })

    if err != nil {
        return err
    }

    respexecruncode,err1 := cli.ContainerExecAttach(ctx,respexec.ID,types.ExecStartCheck{
        Tty:false,
		Detach:false,
		
    })
	defer respexecruncode.Close()
    if err1 != nil {
        fmt.Println(err1)
		return err1
    }

	timer := time.Now().Unix()
	for {
		execInfo,err:= cli.ContainerExecInspect(ctx,respexec.ID)
		if err != nil {
			return err
		}
		if execInfo.Running == true {
			time.Sleep(1*time.Second)
		} else {
			break
		}
		now := time.Now().Unix()
		if now - timer > int64(self.judge_time_out) {
			return errors.New("remove file error")
		}
	}
	return nil
}

