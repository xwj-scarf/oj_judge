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
)

func (self *JudgeServer) ComplieCodeInContainer(containerId string) error{
	client_info, ok := self.container_pool[containerId]	
	if !ok {
		return errors.New("get container client error!")
	}
	cli := client_info.client
	ctx := context.Background()
 
    respexec,err := cli.ContainerExecCreate(ctx,containerId,types.ExecConfig{
        Cmd: []string{"sh","/tmp/complie.sh"},
    })

    if err != nil {
        return err
    }

    respexecruncode,err1 := cli.ContainerExecAttach(ctx,respexec.ID,types.ExecStartCheck{
        Tty:true,
    })
	fmt.Println(respexecruncode)
    if err1 != nil {
        fmt.Println(err1)
		return err1
    }
	return nil
}

func (self *JudgeServer) RunInContainer(containerId string) error{
	client_info, ok := self.container_pool[containerId]
	if !ok {
		return errors.New("get container client error")
	}
	cli := client_info.client
	ctx := context.Background()

    respexecruncode,err := cli.ContainerExecCreate(ctx,containerId,types.ExecConfig{
        AttachStdout:true,
        AttachStderr:true,
        Cmd: []string{"sh","/tmp/do.sh"},
    })

    if err != nil {
        fmt.Println(err)
        return err
    }

    resprunexecruncode,err := cli.ContainerExecAttach(ctx,respexecruncode.ID,types.ExecStartCheck{
        Tty:true,
    })
	fmt.Println(resprunexecruncode)
    if err != nil {
        fmt.Println(err)
		return err
    }	
	return nil
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
	client_info, ok := self.container_pool[containerId]		
	if !ok {
		return errors.New("get container client error!")
	}
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

	client_info, ok := self.container_pool[container_id]		
	if !ok {
		return errors.New("get container client error!")
	}
	cli := client_info.client
	ctx := context.Background()

	fmt.Println(file_path)
    returnoutput,out,err1 := cli.CopyFromContainer(ctx,container_id,file_path)
	if err1 != nil {
		fmt.Println(err1)
		return err1
	}
    //defer returnoutput.Close()
	//fmt.Println(err1)
    fmt.Println(out)

    tr := tar.NewReader(returnoutput)
    _, err := tr.Next()
    if err != nil {
        fmt.Println(err)
		return err
    }

    file, err2 := os.Create(dest_path)
    if err2 != nil {
        fmt.Println(err2)
		return err2
    }
    defer file.Close()

    _, err = io.Copy(file, tr)
    if err != nil {
        fmt.Println(err)
		return err
    }
	return nil
}
