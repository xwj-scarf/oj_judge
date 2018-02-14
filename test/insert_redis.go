package main

import ( "math/rand"
	"fmt"
	"time"
    "github.com/garyburd/redigo/redis"
    "encoding/json"
	"database/sql"
    _"github.com/go-sql-driver/mysql"
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

	for {
		for i:=0;i<5;i++ {
			op := rand.Intn(3)
			remark[op] ++
			insert_to_redis(op)
		}
		count  = count + 5
		if count >= 500 {
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
    stmt, err := db.Prepare(`insert into submit_info (pid,uid,update_time) values(?,?,?)`)
    defer stmt.Close()
    if err != nil {
        fmt.Println(err)
        return 
    }
    res,err := stmt.Exec("1","1",now)
    if err != nil {
        fmt.Println(err)
        return 
    }

    id, _ := res.LastInsertId()
    fmt.Println(id) 

    code1 := &C{
        Code: code,         
		Pid:     1,
		Uid:    1,
		Sid:   int(id),
    }

    data,err := json.Marshal(code1)
    if err != nil {
        fmt.Println(err)
    }
    conn.Do("lpush","test",data)    
}
