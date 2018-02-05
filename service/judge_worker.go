package judgeServer

import (
	"time"
	"fmt"
)

type JudgeWorker struct {
	Manager *JudgeServer
}

func (self *JudgeWorker) Run() {
	go self.GetTask()
	fmt.Println("run .......")

}


func (self *JudgeWorker) GetTask() {
	for {
		is_idle := 0
		var idle_container_id  []string
		for k,v := range self.Manager.container_pool{
			if v.is_work == false {
				is_idle = is_idle + 1
				idle_container_id = append(idle_container_id,k)
			}
		}	
		
		//from redis get is_idle task
		to_do := self.Manager.GetRedisTask(is_idle)

		for k,v := range to_do {
			go self.Assign(v,idle_container_id[k])
		}	
		time.Sleep(10*time.Second)
	}
}

func (self *JudgeWorker) Assign(taskinfo *SubmitInfo, container_id string) {
	
	self.Manager.container_pool[container_id].is_work = true
	fmt.Println(taskinfo)
	fmt.Println(container_id)

	err := self.Manager.CreateFile(taskinfo.Code,container_id,"code.cpp")
	if err != nil {
		fmt.Println("create code file error!")
		return
	}

	//get standard input and output
	//err = self.Manager.GetStandardIOP(taskinfo.Pid)
	//if err != nil {
	//	fmt.Println("get standard input and output error!")
	//	return
	//}
}




