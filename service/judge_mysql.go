package judgeServer

import(
	"database/sql"
	_"github.com/go-sql-driver/mysql"
	"fmt"
	"time"
)

type judgeMysql struct { 
	Manager *JudgeServer 
	db *sql.DB
	mysqlInfo MysqlInfo	
}

func (self *judgeMysql) Init(mysqlinfo MysqlInfo) {
	self.mysqlInfo = mysqlinfo
	var err error
	info := self.mysqlInfo.User + ":" + self.mysqlInfo.Password + "@tcp(" + self.mysqlInfo.Host + ":" + self.mysqlInfo.Port + ")/" + self.mysqlInfo.Database + "?" + "charset=" + self.mysqlInfo.Charset
	fmt.Println(info)
	self.db,err = sql.Open("mysql",info)
    if err != nil {
		fmt.Println(err)
        return
    }
	self.db.SetMaxOpenConns(mysqlinfo.MaxOpenConns)
	self.db.SetMaxIdleConns(mysqlinfo.IdleConns)
	fmt.Println("init mysql success")
}

func (self *judgeMysql) Stop() {
	self.db.Close()
}

func (self *judgeMysql) MarkUserCe(sid int) {
	now := time.Now().Unix()
    stmt, err := self.db.Prepare(`update submit_status set status = ?, update_time = ? where id = ?`)
    defer stmt.Close()
    if err != nil {
        fmt.Println(err)
        return 
    }
    res,err := stmt.Exec(1,now,sid)
    if err != nil {
        fmt.Println(err)
        return 
    }
	num,err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
		return
	}
	if num <= 0 {
		fmt.Println("update error")
		return
	}
}

func (self *judgeMysql) MarkUserAc(sid int) {
	now := time.Now().Unix()
	stmt, err := self.db.Prepare(`update submit_status set status = ?, update_time = ? where id = ?`)
	defer stmt.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	res,err := stmt.Exec(6,now,sid)
	if err != nil {
		fmt.Println(err)
		return
	}
	num,err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
		return
	}
	if num <= 0 {
		fmt.Println("update error")
		return
	}
}

func (self *judgeMysql) MarkUserWa(sid int) {
	now := time.Now().Unix()
	stmt, err := self.db.Prepare(`update submit_status set status = ?, update_time = ? where id = ?`)
	defer stmt.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	res,err := stmt.Exec(2,now,sid)
	if err != nil {
		fmt.Println(err)
		return
	}
	num,err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
		return
	}
	if num <= 0 {
		fmt.Println("update error")
		return
	}

}

func (self *judgeMysql) MarkError(sid int) {
	now := time.Now().Unix()
	stmt, err := self.db.Prepare(`update submit_status set status = ?, update_time = ? where id = ?`)
	defer stmt.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	res,err := stmt.Exec(404,now,sid)
	if err != nil {
		fmt.Println(err)
		return
	}
	num,err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
		return
	}
	if num <= 0 {
		fmt.Println("update error")
		return
	}
}

func (self *judgeMysql) MarkUserTle() {

}
