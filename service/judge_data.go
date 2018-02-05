package judgeServer

import (
    "github.com/docker/docker/client"
)

type ClientInfo struct{
	client *client.Client
	is_work bool
}

type SubmitInfo struct {
	Code string
	Pid int
	Sid int
}
