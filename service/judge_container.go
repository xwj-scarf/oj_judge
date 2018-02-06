package judgeServer

import (
    "archive/tar"
    "github.com/docker/docker/api/types"
    "golang.org/x/net/context"
    "bytes"
    "io/ioutil"
	"errors"
)


func (self *JudgeServer) SendToContainer(file_name, containerId string) error{

	client_info, ok := self.container_pool[containerId]		
	if !ok {
		return errors.New("get container client error!")
	}
	cli := client_info.client
	ctx := context.Background()
 
	filePath := self.tmp_path + "/" + containerId + "/" + file_name 
	destPath := "/tmp/"
    code,e1 := ioutil.ReadFile(filePath)
    if e1 != nil {
        return e1
    }

    buf_code:=bytes.NewBuffer(code)

    buf0 := new(bytes.Buffer)
    tw := tar.NewWriter(buf0)
    defer tw.Close()

    tarHeader := &tar.Header{
        Name: filePath,
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

