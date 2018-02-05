package judgeServer

import (
	"os"
)

func (self *JudgeServer) CreateFile(message,container_id,file_name string) error{
	file,err := os.Create(self.tmp_path+"/"+container_id+"/"+file_name)
	if err != nil {
		return err
	}
	_,err1 := file.WriteString(message)
	if err1 != nil {
		return err1
	}
	return nil	
}
