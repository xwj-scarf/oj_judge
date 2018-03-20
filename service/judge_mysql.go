package judgeServer

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type judgeMysql struct {
	manager   *JudgeServer
	db        *sql.DB
	mysqlInfo MysqlInfo
}

func (self *judgeMysql) Init(mysqlinfo MysqlInfo) {
	self.mysqlInfo = mysqlinfo
	var err error
	info := self.mysqlInfo.User + ":" + self.mysqlInfo.Password + "@tcp(" + self.mysqlInfo.Host + ":" + self.mysqlInfo.Port + ")/" + self.mysqlInfo.Database + "?" + "charset=" + self.mysqlInfo.Charset
	fmt.Println(info)
	self.db, err = sql.Open("mysql", info)
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

func (self *judgeMysql) MarkUserStatus(mem_use,time_use,sid,cid,uid,status int) {
	if cid <= 0 {
		self.MarkNormalStatus(mem_use,time_use,sid,cid,uid,status)
	} else {
		self.MarkContestStatus(mem_use,time_use,sid,cid,uid,status)
	}
}

func (self *judgeMysql) MarkNormalStatus(mem_use,time_use,sid,cid,uid,status int) {
	tx, err := self.db.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}
	if !self.UpdateSubmitInfo(mem_use,time_use,sid,cid,uid,status,tx) || !self.UpdateUserStatistic(uid,status,tx) {
		fmt.Println("update error!")
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (self *judgeMysql) UpdateSubmitInfo(mem_use,time_use,sid,cid,uid,status int,tx *sql.Tx) bool{
	now := time.Now().Unix()
	var SQL string
	if status != Ac {
		SQL = "update submit_info set memory_use = ?, time_use = ?, status = ?, update_time = ? where id = ?" 
	} else {
		SQL = "update submit_info a inner join problem_info b on a.pid = b.pid set a.memory_use = ?,a.time_use = ?, a.status = ?, a.update_time = ?, b.ac_num = b.ac_num + 1 where a.id = ?"
	}
	var args []interface{}
	args = append(args,mem_use)
	args = append(args,time_use)
	args = append(args,status)
	args = append(args,now)
	args = append(args,sid)
	return self.UpdateData(tx,SQL,args)
}

func (self *judgeMysql) UpdateUserStatistic(uid,status int, tx *sql.Tx) bool{
	var args []interface{}

	columes := self.GetColumes(status)
	SQL := "update user_statistic set " + columes + " = " + columes + " + 1 where uid = ? and is_contest = 0"
	args = append(args,uid)
	
	return self.UpdateData(tx,SQL,args)
}

func (self *judgeMysql) MarkContestStatus(mem_use,time_use,sid,cid,uid,status int) {
	tx, err := self.db.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}
	if !self.UpdateContestSubmitInfo(mem_use,time_use,sid,cid,uid,status,tx) || !self.UpdateContestUserStatistic(uid,status,tx) {
		fmt.Println("update error!")
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (self *judgeMysql) UpdateContestSubmitInfo(mem_use,time_use,sid,cid,uid,status int,tx *sql.Tx) bool{
	now := time.Now().Unix()
	var SQL string
	if status != Ac {
		SQL = "update contest_submit_info set memory_use = ?, time_use = ?, status = ?, update_time = ? where id = ?"
	} else {
		SQL = "update contest_submit_info a inner join contest_problem_info b on a.pid = b.show_pid set a.status = ?,a.time_use = ?, a.memory_use = ?, a.update_time = ?, b.ac_num = b.ac_num + 1 where a.id = ? and b.contest_id = ?"
	}
	var args []interface{}
	args = append(args,mem_use)
	args = append(args,time_use)
	args = append(args,status)
	args = append(args,now)
	args = append(args,sid)
	return self.UpdateData(tx,SQL,args)
}

func (self *judgeMysql) UpdateContestUserStatistic(uid,status int, tx *sql.Tx) bool{
	var args []interface{}

	columes := self.GetColumes(status)
	SQL := "update user_statistic set " + columes + " = " + columes + " + 1 where uid = ? and is_contest = 1"
	args = append(args,uid)
	
	return self.UpdateData(tx,SQL,args)
}

func (self *judgeMysql) GetColumes(status int) string {
	switch status {
		case Ce:
			return "ce_count"
		case Ac:
			return "ac_count"
		case Wa:
			return "wa_count"
		case Re:
			return "re_count"
		case Tle:
			return "tle_count"
		case Mle:
			return "mle_count"
		default:
			return "other_count" 
	}
	return "other_count"
}

func (self *judgeMysql) GetTimeAndMemoryLimit(pid int) (int, int, error) {
	rows, _ := self.db.Query("select time_limit,memory_limit from problem_info where pid = ?", pid)
	defer rows.Close()
	var time_limit int
	var memory_limit int
	for rows.Next() {
		if err := rows.Scan(&time_limit, &memory_limit); err != nil {
			fmt.Println(err)
			return 0, 0, err
		}
		fmt.Println(time_limit)
		fmt.Println(memory_limit)
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
		return 0, 0, err
	}
	return time_limit, memory_limit, nil
}

func (self *judgeMysql) UpdateData(tx *sql.Tx,SQL string, args []interface{}) bool{
	stmt, err := tx.Prepare(SQL)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer stmt.Close()
	res,err := stmt.Exec(args...)	
	if err != nil {
		fmt.Println(err)
		return false 
	}
	num,err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
		return false 
	}
	if num <= 0 {
		fmt.Println("update error!")
		return false
	}
	return true 
}
