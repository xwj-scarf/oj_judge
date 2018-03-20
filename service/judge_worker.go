package judgeServer

import (
	"fmt"
	"time"
	"os"
    "io/ioutil"
	"errors"
	"crypto/md5"
	"io"
	"strconv"
    "encoding/hex"
	"strings"
)

type JudgeWorker struct {
	manager *JudgeServer
}

func (self *JudgeWorker) Run() {
	go self.GetTask()
	go self.checkContainerHealth()
	fmt.Println("run .......")

	for {
		time.Sleep(1*time.Second)
	}
}

func (self *JudgeWorker) checkContainerHealth() {
	for {
		self.manager.judge_mutex.RLock()
		for k,_ := range self.manager.container_pool {
			if !self.manager.checkContainerInspect(k) {
				self.manager.restartContainer(k)
			}			
		}
		self.manager.judge_mutex.RUnlock()
		time.Sleep(10*time.Second)
	}
}

func (self *JudgeWorker) GetTask() {
	for {
		is_idle := 0
		var idle_container_id  []string
		self.manager.judge_mutex.RLock()
		for k,v := range self.manager.container_pool{
			if v.is_work == false {
				is_idle = is_idle + 1
				idle_container_id = append(idle_container_id,k)
			}
		}
		self.manager.judge_mutex.RUnlock()	
		//from redis get is_idle task
		to_do := self.manager.GetRedisTask(is_idle)

		for k,v := range to_do {
			go self.Assign(v,idle_container_id[k])
		}
	
		time.Sleep(2*time.Second)
	}
}

func (self *JudgeWorker) Assign(taskinfo *SubmitInfo, container_id string) {
	defer func (container_id string) {
		self.manager.DelFile(container_id)
		self.manager.DelFileInContainer(container_id)
		self.manager.judge_mutex.Lock()
		self.manager.container_pool[container_id].is_work = false
		self.manager.judge_mutex.Unlock()

	}(container_id)

	self.manager.judge_mutex.Lock()
	self.manager.container_pool[container_id].is_work = true
	self.manager.judge_mutex.Unlock()

	//create code.cpp
	err := self.manager.CreateFile(taskinfo.Code,container_id,"code.cpp")
	if err != nil {
		fmt.Println("create code file error!")
		return
	}

	//copy input.txt
	err = self.manager.CopyFile(container_id,taskinfo.Pid,"input.txt")
	if err != nil {
		fmt.Println("copy input file error!")
		return
	}
	
	//copy output.txt
	err = self.manager.CopyFile(container_id,taskinfo.Pid,"output.txt")
	if err != nil {
		fmt.Println("copy output file error!")
		return
	}

	//send code to container
	err = self.manager.SendToContainer("code.cpp" ,container_id)
	if err != nil {
		fmt.Println("send code to container error!")
		return
	}

	//send input to container
	err = self.manager.SendToContainer("input.txt" ,container_id)
	if err != nil {
		fmt.Println("send input to container error!")
		self.manager.mysql.MarkUserStatus(0,0,taskinfo.Sid,taskinfo.Cid,taskinfo.Uid,Error)
		return
	}

	self.manager.ChangePermission(container_id,"/tmp/code.cpp")
	self.manager.ChangePermission(container_id,"/tmp/input.txt")

	//complie code in container
	err = self.manager.ComplieCodeInContainer(container_id) 
	if err != nil {
		fmt.Println("complie code in container error!")
		self.manager.mysql.MarkUserStatus(0,0,taskinfo.Sid,taskinfo.Cid,taskinfo.Uid,Ce)
		return
	 }

	self.manager.ChangePermission(container_id,"/tmp/code")

	err = self.manager.CopyFromContainer(container_id,"ce.txt")
	if err != nil {
		fmt.Println("copy from container error!")
		self.manager.mysql.MarkUserStatus(0,0,taskinfo.Sid,taskinfo.Cid,taskinfo.Uid,Error)
		return
	}

	err = self.JudgeIsCe(container_id) 
	if err != nil {
		fmt.Println("code is ce!")
		self.manager.mysql.MarkUserStatus(0,0,taskinfo.Sid,taskinfo.Cid,taskinfo.Uid,Error)
		return
	}

	//run code in container
	err = self.manager.RunInContainer(container_id) 
	if err != nil {
		fmt.Println("run code in container error!")
		self.manager.mysql.MarkUserStatus(0,0,taskinfo.Sid,taskinfo.Cid,taskinfo.Uid,Error)
		return
	 }

	//copy is_runtime_error.txt in container
	err = self.manager.CopyFromContainer(container_id,"runtime.txt")
	if err != nil {
		fmt.Println("copy runtime.txt from container error")
		self.manager.mysql.MarkUserStatus(0,0,taskinfo.Sid,taskinfo.Cid,taskinfo.Uid,Error)
		return
	}

	err = self.JudgeIsRunTimeError(container_id)
	if err != nil {
		fmt.Println(err)
		self.manager.mysql.MarkUserStatus(0,0,taskinfo.Sid,taskinfo.Cid,taskinfo.Uid,Re)
		return	
	}

	//copy time and memory use in container
	err = self.manager.CopyFromContainer(container_id,"time.txt")
	if err != nil {
		fmt.Println("copy time.txt from container error")
		self.manager.mysql.MarkUserStatus(0,0,taskinfo.Sid,taskinfo.Cid,taskinfo.Uid,Error)
		return
	}

	err = self.manager.CopyFromContainer(container_id,"m.txt")
	if err != nil {
		fmt.Println(err)
		fmt.Println("copy mem.txt from container error")
		self.manager.mysql.MarkUserStatus(0,0,taskinfo.Sid,taskinfo.Cid,taskinfo.Uid,Error)
		return
	}

	//judge is_time_out and is_memory_out 
	use_time,use_memory,is_tleormle,err3 := self.JudgeIsTimeOutAndMemoryOut(container_id,taskinfo.Pid,taskinfo.Sid,taskinfo.Cid,taskinfo.Uid)
	if err3 != nil {
		fmt.Println(err3)
		self.manager.mysql.MarkUserStatus(0,0,taskinfo.Sid,taskinfo.Cid,taskinfo.Uid,Error)
		return
	}
	
	if is_tleormle {
		return
	}

	//copy output from container
	err = self.manager.CopyFromContainer(container_id,"output.txt")
	if err != nil {
		fmt.Println("copy output.txt from container error")
		self.manager.mysql.MarkUserStatus(0,0,taskinfo.Sid,taskinfo.Cid,taskinfo.Uid,Error)
		return
	}

	//judge output 
	err = self.JudgeIsAc(container_id,taskinfo.Pid)
	if err != nil {
		fmt.Println("judge output error!")
		self.manager.mysql.MarkUserStatus(use_memory,use_time,taskinfo.Sid,taskinfo.Cid,taskinfo.Uid,Wa)
		return
	 }
	self.manager.mysql.MarkUserStatus(use_memory,use_time,taskinfo.Sid,taskinfo.Cid,taskinfo.Uid,Ac)
}

func (self *JudgeWorker) JudgeIsAc(container_id string,pid int) error {
	container_output_path := self.manager.tmp_path + "/" + container_id + "/" + "output.txt"
	standard_output_path := self.manager.input_path + "/" + strconv.Itoa(pid) + "/" + "output.txt"
	
	container_output,err:= os.Open(container_output_path)
	if err != nil {
		fmt.Println("open container output.txt error!")
		fmt.Println(err)
		return err
	}

	standard_output,err := os.Open(standard_output_path)
	if err != nil {
		fmt.Println("open standard output.txt error!")
		fmt.Println(err)
		return err
	}

	md5_container_output := md5.New()
	io.Copy(md5_container_output,container_output)
	md5_container_output_md5 :=hex.EncodeToString(md5_container_output.Sum(nil)) 
	fmt.Println("md5 container output.txt is : ",string(md5_container_output_md5))

	md5_standard_output := md5.New()
	io.Copy(md5_standard_output,standard_output)
	md5_standard_output_md5 :=hex.EncodeToString(md5_standard_output.Sum(nil))

	fmt.Println("md5 standard output.txt is : ",string(md5_standard_output_md5))

	if string(md5_standard_output_md5) == string(md5_container_output_md5) {
		fmt.Println("ac!")
		return nil
	}
	return errors.New("wa")

}

func (self *JudgeWorker) JudgeIsRunTimeError(container_id string) error {
	dest_path := self.manager.tmp_path + "/" + container_id + "/" + "runtime.txt"

	fileInfo, err := os.Stat(dest_path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fileSize := fileInfo.Size() //获取size
	fmt.Println(fileSize)
	if fileSize == 0 {
		fmt.Println("no runtime error")
		return nil
	}else {
		b, err := ioutil.ReadFile(dest_path)
		if err != nil {
			fmt.Print(err)
			return err
		}
		str := string(b)
		fmt.Println(str)
		return errors.New("runtime error!")
	}
	return nil	
}

func (self *JudgeWorker) JudgeIsTimeOutAndMemoryOut(container_id string,pid,sid int,cid int,uid int) (int,int,bool,error) {
	time_limit,mem_limit,err:= self.manager.mysql.GetTimeAndMemoryLimit(pid)
	if err != nil {
		fmt.Println("get time and memory limit error!")
		return 0,0,false,errors.New("get time and memory limit error")
	}

	dest_time_path := self.manager.tmp_path + "/" + container_id + "/" + "time.txt"
	b,err1 := ioutil.ReadFile(dest_time_path)

	if err1 != nil {
		fmt.Println(err1)
		return 0,0,false,err1
	}

	fmt.Println(string(b))

	use_time,err5 := strconv.Atoi(strings.Replace(string(b),"\n","",-1))

	if err5 != nil {
		fmt.Println(err5)
	}
	fmt.Println("use_time is ",use_time)

	if use_time > time_limit {
		fmt.Println("time limit!")
		self.manager.mysql.MarkUserStatus(0,time_limit,sid,cid,uid,Tle)
		return time_limit,0,true,nil
	}

	dest_mem_path := self.manager.tmp_path + "/" + container_id + "/" + "m.txt"
	b, err1 = ioutil.ReadFile(dest_mem_path)
	if err1 != nil {
		fmt.Println(err1)
		return 0,0,false,err1
	}
	use_memory,_ := strconv.Atoi(strings.Replace(string(b),"\n","",-1))

	fmt.Println("use_mem is ",use_memory)

	if use_memory > mem_limit {
		fmt.Println("memory limit!")
		self.manager.mysql.MarkUserStatus(mem_limit,0,sid,cid,uid,Mle)
		return 0,mem_limit,true,nil
	}
	return use_time,use_memory,false,nil
	
}

func (self *JudgeWorker) JudgeIsCe(container_id string)error{
	dest_path := self.manager.tmp_path+"/"+container_id+"/"+"ce.txt"
	fileInfo, err := os.Stat(dest_path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fileSize := fileInfo.Size() //获取size
	fmt.Println(fileSize)
	if fileSize == 0 {
		fmt.Println("complie success")
		return nil
	}else {
		b, err := ioutil.ReadFile(dest_path)
		if err != nil {
			fmt.Print(err)
			return err
		}
		str := string(b)
		fmt.Println(str)
		return errors.New("complie error!")
	}
	return nil	
}
