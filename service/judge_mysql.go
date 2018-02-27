package judgeServer

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type judgeMysql struct {
	Manager   *JudgeServer
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

func (self *judgeMysql) MarkUserCe(sid int, cid int) {
	now := time.Now().Unix()
	var stmt *sql.Stmt
	var err error
	if cid <= 0 {
		stmt, err = self.db.Prepare(`update submit_info set status = ?, update_time = ? where id = ?`)
	} else {
		stmt, err = self.db.Prepare(`update contest_submit_info set status = ?,update_time = ? where id = ?`)
	}
	defer stmt.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := stmt.Exec(1, now, sid)
	if err != nil {
		fmt.Println(err)
		return
	}
	num, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
		return
	}
	if num <= 0 {
		fmt.Println("update error")
		return
	}
}

func (self *judgeMysql) MarkUserAc(sid int, use_time, use_memory int, cid int) {
	now := time.Now().Unix()
	var stmt *sql.Stmt
	var err error
	if cid <= 0 {
		stmt, err = self.db.Prepare(`update submit_info a inner join problem_info b on a.pid = b.pid set a.status = ?,a.time_use = ?, a.memory_use = ?, a.update_time = ?, b.ac_num = b.ac_num + 1 where a.id = ?`)
	} else {
		stmt, err = self.db.Prepare(`update contest_submit_info a inner join contest_problem_info b on a.pid = b.show_pid set a.status = ?,a.time_use = ?, a.memory_use = ?, a.update_time = ?, b.ac_num = b.ac_num + 1 where a.id = ? and b.contest_id = ?`)
	}
	defer stmt.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	if cid <= 0 {
		res, err := stmt.Exec(6, use_time, use_memory, now, sid)
		if err != nil {
			fmt.Println(err)
			return
		}
		num, err := res.RowsAffected()
		if err != nil {
			fmt.Println(err)
			return
		}
		if num <= 0 {
			fmt.Println("update error")
			return
		}
	} else {

		res, err := stmt.Exec(6, use_time, use_memory, now, sid, cid)
		if err != nil {
			fmt.Println(err)
			return
		}
		num, err := res.RowsAffected()
		if err != nil {
			fmt.Println(err)
			return
		}
		if num <= 0 {
			fmt.Println("update error")
			return
		}
	}
}

func (self *judgeMysql) MarkUserWa(sid int, use_time, use_memory int, cid int) {
	now := time.Now().Unix()
	var stmt *sql.Stmt
	var err error
	if cid <= 0 {
		stmt, err = self.db.Prepare(`update submit_info set status = ?,time_use = ?, memory_use = ?, update_time = ? where id = ?`)
	} else {
		stmt, err = self.db.Prepare(`update contest_submit_info set status = ?,time_use = ?, memory_use = ?,update_time = ? where id = ?`)
	}
	defer stmt.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := stmt.Exec(2, use_time, use_memory, now, sid)
	if err != nil {
		fmt.Println(err)
		return
	}
	num, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
		return
	}
	if num <= 0 {
		fmt.Println("update error")
		return
	}

}

func (self *judgeMysql) MarkError(sid int, cid int) {
	now := time.Now().Unix()
	var stmt *sql.Stmt
	var err error
	if cid <= 0 {
		stmt, err = self.db.Prepare(`update submit_info set status = ?, update_time = ? where id = ?`)
	} else {
		stmt, err = self.db.Prepare(`update contest_submit_info set status = ?,update_time = ? where id = ?`)
	}
	defer stmt.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := stmt.Exec(404, now, sid)
	if err != nil {
		fmt.Println(err)
		return
	}
	num, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
		return
	}
	if num <= 0 {
		fmt.Println("update error")
		return
	}
}

func (self *judgeMysql) MarkTle(time_use, sid int, cid int) {
	now := time.Now().Unix()

	var stmt *sql.Stmt
	var err error
	if cid <= 0 {
		stmt, err = self.db.Prepare(`update submit_info set time_use = ? ,status = ?, update_time = ? where id = ?`)
	} else {
		stmt, err = self.db.Prepare(`update contest_submit_info set time_use = ? ,status = ?, update_time = ? where id = ?`)
	}
	defer stmt.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := stmt.Exec(time_use, 3, now, sid)
	if err != nil {
		fmt.Println(err)
		return
	}
	num, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
		return
	}
	if num <= 0 {
		fmt.Println("update error")
		return
	}
}

func (self *judgeMysql) MarkMle(mem_use, sid int, cid int) {
	now := time.Now().Unix()
	var stmt *sql.Stmt
	var err error
	if cid <= 0 {
		stmt, err = self.db.Prepare(`update submit_info set memory_use = ?,status = ?, update_time = ? where id = ?`)
	} else {
		stmt, err = self.db.Prepare(`update contest_submit_info set memory_use = ?,status = ?, update_time = ? where id = ?`)
	}
	defer stmt.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := stmt.Exec(mem_use, 4, now, sid)
	if err != nil {
		fmt.Println(err)
		return
	}
	num, err := res.RowsAffected()
	if err != nil {
		fmt.Println(err)
		return
	}
	if num <= 0 {
		fmt.Println("update error")
		return
	}
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
