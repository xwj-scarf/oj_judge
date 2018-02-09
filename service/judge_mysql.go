package judgeServer

import(
	"database/sql"
	_"github.com/go-sql-driver/mysql"
)

type judgeMysql struct {
	Manager *JudgeServer
	db *sql.DB
	
}

func (self *judgeMysql) Init() {
		
}
