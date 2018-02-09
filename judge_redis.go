package main

import (
	"github.com/garyburd/redigo/redis"
	"fmt"
	"encoding/json"
)

type C struct {
	Code string 
	Pid int	
	SubmitId int
}

func main() {
	conn,_ := redis.Dial("tcp","127.0.0.1:6379")
	code1 := &C{
		Code:	 `#include<iostream>
				 #include<cstdio>
				  using namespace std;
				  int main(){
					int a,b;
					while(scanf("%d%d",&a,&b)!=EOF) {
						printf("%d %d\n",a+1,b+1);
					}				
}`,
		Pid:	 1,
	
	}
	data,err := json.Marshal(code1)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(data))
	conn.Do("lpush","test",data)	

}
