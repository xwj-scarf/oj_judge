package main

import ( 
	_"math/rand"
	"fmt"
	"time"
    "github.com/garyburd/redigo/redis"
    "encoding/json"
	"database/sql"
    _"github.com/go-sql-driver/mysql"
	"sync"
)
type C struct {
	Code string
	Uid int
	Pid int
	Sid int
}

var db *sql.DB
var err error
var conn redis.Conn
func main() {

	var mutex *sync.RWMutex
	mutex = new (sync.RWMutex)
    conn,_ = redis.Dial("tcp","127.0.0.1:6379")
    db,err = sql.Open("mysql","root:123456@tcp(127.0.0.1:3306)/oj?charset=utf8")
    if err != nil {
        return
    }
    defer db.Close()

	count := 0
	remark := make(map[int]int)
	remark[0]=0
	remark[1]=0
	remark[2]=0

	//insert_to_redis(1)
	
	for {
		for i:=0;i<500;i++ {
			op := i%3
			mutex.Lock()			
			remark[op] ++
			mutex.Unlock()
			insert_to_redis(op)
		}
		count  = count + 5
		if count >= 10 {
			break
		}
		time.Sleep(1*time.Second)
	}
	fmt.Println("0 ce   1 ac   2 wa")
	fmt.Println(remark)
}



func insert_to_redis(op int) {
	var code string
	if op == 1 {
		code =     `#include<iostream>
                 #include<cstdio>
                  using namespace std;
                  int main(){
                    int a,b;
                    while(scanf("%d%d",&a,&b)!=EOF) {
                        printf("%d %d\n",a+1,b+1);
                    }               
				}`
	} 

	if op == 0 {
		   code = `#include<iostream>
                 #include<cstdio>
                  using namespace std;
                  int main(){
                    int a,b;
                    while(scanf("%d%d",&a,&b)!=EOF) {
                        printf("%d %d\n",a+1,b+1)
                    }               
			}`
	}

	if op == 2 {
		   code = `#include<iostream>
                 #include<cstdio>
                  using namespace std;
                  int main(){
                    int a,b;
                    while(scanf("%d%d",&a,&b)!=EOF) {
                        printf("%d %d\n",a+2,b+2);
                    }               
			}`

	}

	now := time.Now().Unix()

	stmt,_ := db.Prepare(`update problem_info set total_num = total_num + 1 where pid = ?`)
	_,err123 := stmt.Exec(1)
	if err123 != nil {
		fmt.Println(err123)
	}
	defer stmt.Close()

    stmt1, err := db.Prepare(`insert into submit_info (pid,uid,time_use,memory_use,add_time,update_time) values(?,?,?,?,?,?)`)
    defer stmt1.Close()
    if err != nil {
        fmt.Println(err)
        return 
    }
    res1,err := stmt1.Exec("1","8",0,0,now,now)
    if err != nil {
        fmt.Println(err)
        return 
    }

    id, _ := res1.LastInsertId()
    fmt.Println(id) 

    code1 := &C{
        Code: code,         
		Pid:     1,
		Uid:    8,
		Sid:   int(id),
    }

    data,err := json.Marshal(code1)
    if err != nil {
        fmt.Println(err)
    }
    conn.Do("lpush","test",data)    
}
