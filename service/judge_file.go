package judgeServer

import (
	"os"
	"strconv"
	"io"
)

func (self *JudgeServer) CreateFile(message,container_id,file_name string) error{
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

func (self *JudgeServer) CopyFile(container_id string, pid int,file_name string) error{
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


