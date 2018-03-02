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
	Manager *JudgeServer
}

func (self *JudgeWorker) Run() {
	go self.GetTask()
	fmt.Println("run .......")

	for {
		time.Sleep(1*time.Second)
	}
}


func (self *JudgeWorker) GetTask() {
	for {
		is_idle := 0
		var idle_container_id  []string
		self.Manager.judge_mutex.RLock()
		for k,v := range self.Manager.container_pool{
			if v.is_work == false {
				is_idle = is_idle + 1
				idle_container_id = append(idle_container_id,k)
			}
		}
		self.Manager.judge_mutex.RUnlock()	
		//from redis get is_idle task
		to_do := self.Manager.GetRedisTask(is_idle)

		for k,v := range to_do {
			go self.Assign(v,idle_container_id[k])
		}
	
		time.Sleep(2*time.Second)
	}
}

func (self *JudgeWorker) Assign(taskinfo *SubmitInfo, container_id string) {
	defer func (container_id string) {
		self.Manager.DelFile(container_id)
		self.Manager.DelFileInContainer(container_id)
		self.Manager.judge_mutex.Lock()
		self.Manager.container_pool[container_id].is_work = false
		self.Manager.judge_mutex.Unlock()

	}(container_id)

	self.Manager.judge_mutex.Lock()
	self.Manager.container_pool[container_id].is_work = true
	self.Manager.judge_mutex.Unlock()

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
		self.Manager.mysql.MarkError(taskinfo.Sid,taskinfo.Cid)
		return
	}

	//complie code in container
	err = self.Manager.ComplieCodeInContainer(container_id) 
	if err != nil {
		fmt.Println("complie code in container error!")
		self.Manager.mysql.MarkUserCe(taskinfo.Sid,taskinfo.Cid)
		//TODO   Write to Mysql  mark failed times+1 
		return
	 }

	err = self.Manager.CopyFromContainer(container_id,"ce.txt")
	if err != nil {
		fmt.Println("copy from container error!")
		self.Manager.mysql.MarkError(taskinfo.Sid,taskinfo.Cid)
		return
	}

	err = self.JudgeIsCe(container_id) 
	if err != nil {
		fmt.Println("code is ce!")
		self.Manager.mysql.MarkUserCe(taskinfo.Sid,taskinfo.Cid)
		//mark
		return
	}

	//run code in container
	err = self.Manager.RunInContainer(container_id) 
	if err != nil {
		fmt.Println("run code in container error!")
		self.Manager.mysql.MarkError(taskinfo.Sid,taskinfo.Cid)
		//TODO   Write to Mysql  mark re times+1 
		return
	 }

	//copy is_runtime_error.txt in container
	err = self.Manager.CopyFromContainer(container_id,"runtime.txt")
	if err != nil {
		fmt.Println("copy runtime.txt from container error")
		self.Manager.mysql.MarkError(taskinfo.Sid,taskinfo.Cid)
		return
	}

	err = self.JudgeIsRunTimeError(container_id)
	if err != nil {
		fmt.Println(err)
		self.Manager.mysql.MarkUserRe(taskinfo.Sid,taskinfo.Cid)
		return	
	}

	//copy time and memory use in container
	err = self.Manager.CopyFromContainer(container_id,"time.txt")
	if err != nil {
		fmt.Println("copy time.txt from container error")
		self.Manager.mysql.MarkError(taskinfo.Sid,taskinfo.Cid)
		return
	}

	err = self.Manager.CopyFromContainer(container_id,"m.txt")
	if err != nil {
		fmt.Println(err)
		fmt.Println("copy mem.txt from container error")
		self.Manager.mysql.MarkError(taskinfo.Sid,taskinfo.Cid)
		return
	}

	//judge is_time_out and is_memory_out 
	use_time,use_memory,is_tleormle,err3 := self.JudgeIsTimeOutAndMemoryOut(container_id,taskinfo.Pid,taskinfo.Sid,taskinfo.Cid)
	if err3 != nil {
		fmt.Println(err3)
		self.Manager.mysql.MarkError(taskinfo.Sid,taskinfo.Cid)
		return
	}
	
	if is_tleormle {
		return
	}

	//copy output from container
	err = self.Manager.CopyFromContainer(container_id,"output.txt")
	if err != nil {
		fmt.Println("copy output.txt from container error")
		self.Manager.mysql.MarkError(taskinfo.Sid,taskinfo.Cid)
		return
	}

	//judge output 
	err = self.JudgeIsAc(container_id,taskinfo.Pid)
	if err != nil {
		fmt.Println("judge output error!")
		self.Manager.mysql.MarkUserWa(taskinfo.Sid,use_time,use_memory,taskinfo.Cid)
		//TODO   Write to Mysql  mark wa times+1 
		return
	 }
	self.Manager.mysql.MarkUserAc(taskinfo.Sid,use_time,use_memory,taskinfo.Cid)
}

func (self *JudgeWorker) JudgeIsAc(container_id string,pid int) error {
	container_output_path := self.Manager.tmp_path + "/" + container_id + "/" + "output.txt"
	standard_output_path := self.Manager.input_path + "/" + strconv.Itoa(pid) + "/" + "output.txt"
	
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
	dest_path := self.Manager.tmp_path + "/" + container_id + "/" + "runtime.txt"

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

func (self *JudgeWorker) JudgeIsTimeOutAndMemoryOut(container_id string,pid,sid int,cid int) (int,int,bool,error) {
	time_limit,mem_limit,err:= self.Manager.mysql.GetTimeAndMemoryLimit(pid)
	if err != nil {
		fmt.Println("get time and memory limit error!")
		return 0,0,false,errors.New("get time and memory limit error")
	}

	dest_time_path := self.Manager.tmp_path + "/" + container_id + "/" + "time.txt"
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
		self.Manager.mysql.MarkTle(time_limit,sid,cid)
		return time_limit,0,true,nil
	}

	dest_mem_path := self.Manager.tmp_path + "/" + container_id + "/" + "m.txt"
	b, err1 = ioutil.ReadFile(dest_mem_path)
	if err1 != nil {
		fmt.Println(err1)
		return 0,0,false,err1
	}
	use_memory,_ := strconv.Atoi(strings.Replace(string(b),"\n","",-1))

	fmt.Println("use_mem is ",use_memory)

	if use_memory > mem_limit {
		fmt.Println("memory limit!")
		self.Manager.mysql.MarkMle(mem_limit,sid,cid)
		return 0,mem_limit,true,nil
	}
	return use_time,use_memory,false,nil
	
}

func (self *JudgeWorker) JudgeIsCe(container_id string)error{
	dest_path := self.Manager.tmp_path+"/"+container_id+"/"+"ce.txt"
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
