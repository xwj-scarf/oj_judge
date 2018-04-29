package judgeServer

import (
	"os"
	"strconv"
	"io"
	"fmt"
	"io/ioutil"
)

type JudgeFileManager struct {
	manager *JudgeServer
	tmp_path string
	input_path string
}

func (self *JudgeFileManager) Init() { 
	for containerId,_ := range self.manager.container_pool {
		tmp_path := self.tmp_path + "/"+containerId
        err := os.MkdirAll(tmp_path,0777)
        if err != nil {
            fmt.Println("create tmp_path error!")
            return
        } 
	}
}

func (self *JudgeFileManager) SetTmpPath(path string) {
    self.tmp_path = path
}

func (self *JudgeFileManager) SetInputPath(path string) {
    self.input_path = path
}

func (self *JudgeFileManager) CreateFile(message,container_id,file_name string) error{
	fmt.Println("start create file")
	file,err := os.Create(self.tmp_path+"/"+container_id+"/"+file_name)
	defer file.Close()
	if err != nil {
		return err
	}
	_,err1 := file.WriteString(message)
	if err1 != nil {
		return err1
	}
	return nil	
}

func (self *JudgeFileManager) CopyFile(container_id string, pid int,file_name string) error{
	file,err := os.Open(self.input_path+"/"+strconv.Itoa(pid)+"/"+file_name)
	defer file.Close()
	if err != nil {
		return err
	}
	
	dstPath := self.tmp_path + "/" + container_id + "/" + file_name
	dst, err1 := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, 0644)
	defer dst.Close()
	if err1 != nil {
		return err
	}
	_, err3 := io.Copy(dst,file)	
	return err3
}

func (self *JudgeFileManager) DelFile(container_id string) {
	file_path := self.tmp_path + "/" + container_id + "/"
    dir_list, err := ioutil.ReadDir(file_path)
    if err != nil {
        fmt.Println("read dir error")
        return
    }
    for _, v := range dir_list {
		err := os.Remove(file_path+v.Name())
		if err != nil {
			fmt.Println(err)
			return	
		}
        fmt.Println( v.Name())
    }
	return
}
