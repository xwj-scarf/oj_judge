package judgeServer

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"encoding/json"
)

type RedisServer struct {
	conn redis.Conn
	address string
}

func (self *RedisServer) RedisInit() {
	var err error
	self.conn, err = redis.Dial("tcp",self.address)
	if err != nil {
		fmt.Println("redis connect error!")
		return 
	}
	fmt.Println("redis connect success")
}

func (self *RedisServer) SetRedisAddress(address string) {
	self.address = address
}

func (self *RedisServer) GetRedisTask(num int) []*SubmitInfo{
	var to_do_list []*SubmitInfo
	for i:=0;i<num;i++ {
		lpop,_ := redis.Bytes(self.conn.Do("lpop","test"))
		to_do := &SubmitInfo{}
		if lpop == nil {
			fmt.Println("no submit to do")
			return to_do_list
		}
		err := json.Unmarshal(lpop,&to_do)
		if err != nil {
			fmt.Println("unmarshal error!")
			return to_do_list
		}
		fmt.Println(to_do.Code)
		to_do_list = append(to_do_list,to_do)	
	}
	return to_do_list
}
