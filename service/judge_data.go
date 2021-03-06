package judgeServer

import (
    "github.com/docker/docker/client"
)

type MysqlInfo struct {
	Host string
	Port string
	User string
	Password string
	Database string
	Charset string
	MaxOpenConns int
	IdleConns int
}

type ClientInfo struct{
	client *client.Client
	is_work bool
}

type SubmitInfo struct {
	Code string
	Pid int     //problem id
	Sid int		//submit id
	Uid int		//user id
	Cid int		//contest id
}

const (
	Judging = 0
	Ce = 1
	Wa = 2
	Tle = 3
	Mle = 4
	Re = 5
    Ac = 6
	Error = 404
)
