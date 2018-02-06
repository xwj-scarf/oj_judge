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
	defer func (container_id string) {
		self.Manager.container_pool[container_id].is_work = false
	}(container_id)

	self.Manager.container_pool[container_id].is_work = true
	//fmt.Println(taskinfo)
	//fmt.Println(container_id)

	//create code.cpp
	err := self.Manager.CreateFile(taskinfo.Code,container_id,"code.cpp")
	if err != nil {
		fmt.Println("create code file error!")
		return
	}

	//copy input.txt
	err = self.Manager.CopyFile(container_id,taskinfo.Pid,"input.txt")
	if err != nil {
		fmt.Println("copy input file error!")
		return
	}
	
	//copy output.txt
	err = self.Manager.CopyFile(container_id,taskinfo.Pid,"output.txt")
	if err != nil {
		fmt.Println("copy output file error!")
		return
	}

	//send code to container
	err = self.Manager.SendToContainer("code.cpp" ,container_id)
	if err != nil {
		fmt.Println("send code to container error!")
		return
	}

	//send input to container
	err = self.Manager.SendToContainer("input.txt" ,container_id)
	if err != nil {
		fmt.Println("send input to container error!")
		return
	}

	//complie code in container
	err = self.Manager.ComplieCodeInContainer(container_id) 
	if err != nil {
		fmt.Println("complie code in container error!")
		//TODO   Write to Mysql  mark ce times+1 
		return
	 }

	//run code in container
	err = self.Manager.RunInContainer(container_id) 
	if err != nil {
		fmt.Println("run code in container error!")
		//TODO   Write to Mysql  mark re times+1 
		return
	 }
	
	//judge output 
	err = self.Manager.JudgeOutput(container_id)
	if err != nil {
		fmt.Println("judge output error!")
		//TODO   Write to Mysql  mark wa times+1 
		return
	 }

	//Write to Mysql mark ac times+1
}




