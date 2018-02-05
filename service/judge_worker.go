package judgeServer

import (
	"time"
	"fmt"
	"os"
)

type JudgeWorker struct {
	manager *JudgeServer
}

func (self *JudgeWorker) Run() {
	//time.Sleep(5*time.Second)
	go self.GetTask()
	fmt.Println("run .......")

}


func (self *JudgeWorker) GetTask() {
	for {
		is_idle := 0
		var idle_container_id  []string
		for k,v := range self.manager.container_pool{
			if v.is_work == false {
				is_idle = is_idle + 1
				idle_container_id = append(idle_container_id,k)
			}
		}	
		
		//from redis get is_idle task
		to_do := self.manager.GetRedisTask(is_idle)

		for k,v := range to_do {
			go self.Assign(v,idle_container_id[k])
		}	
		time.Sleep(10*time.Second)
	}
}

func (self *JudgeWorker) Assign(taskinfo *SubmitInfo, container_id string) {
	
	self.manager.container_pool[container_id].is_work = true
	fmt.Println(taskinfo)
	fmt.Println(container_id)
	file,err := os.Create(self.manager.tmp_path+"/"+container_id+"/"+"code.cpp")
	if err != nil {
		fmt.Println("create file error!")
		return
	}
	_,err2 := file.WriteString(taskinfo.Code)
    //fmt.Println(n)
    if err2 != nil {
		fmt.Println("writestring to file error!")
		return
    }
			
}




