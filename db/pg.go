package db

import (
	"database/sql"
	"github.com/crazy-choose/go/log"
	"github.com/crazy-choose/go/meta"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"math"
)

var (
	cfgMap map[string]meta.Postgres
	pgsMap map[string]*Session // pg 连接map

)

func init() {
	cfgMap = make(map[string]meta.Postgres)
	pgsMap = make(map[string]*Session)
}

func InitPg(key string, opt meta.Postgres) bool {
	if !verifyConfig(opt) {
		return false
	}
	cfgMap[key] = opt
	pgConfig := postgres.Config{
		PreferSimpleProtocol: true,
		DSN:                  genDSN(opt), // DSN data source name
	}
	if db, err := gorm.Open(postgres.New(pgConfig), &gorm.Config{
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	}); err != nil {
		log.Error(err.Error())
		return false
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(opt.MaxIdleConns)
		sqlDB.SetMaxOpenConns(opt.MaxOpenConns)
		_pgs_ := &Session{
			db,
		}
		pgsMap[key] = _pgs_
		log.Info("InitPg success")
	}
	return true
}

func Postgres(key string) *Session {
	if pgsMap == nil {
		return nil
	}

	ret, _ := pgsMap[key]
	return ret
}

func verifyConfig(option meta.Postgres) bool {
	if option.DbName == "" {
		log.Error("dbname is empty")
		return false
	}
	return true
}

func genDSN(option meta.Postgres) string {
	return "host =" + option.Host + " user=" + option.Username + " password=" + option.Password + " dbname=" + option.DbName + " port=" + option.Port + " sslmode=disable TimeZone=Asia/Shanghai"
}

type Session struct {
	*gorm.DB
}

func (s *Session) Transaction(fc func(tx *Session) error, opts ...*sql.TxOptions) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		return fc(&Session{tx})
	}, opts...)
}

func Transaction(fc func(tx *Session) error, opts ...*sql.TxOptions) (err error) {
	if impl == nil {
		InitPgSig(cfgSig)
	}
	return impl.DB.Transaction(func(tx *gorm.DB) error {
		return fc(&Session{tx})
	}, opts...)
}

type Pagination struct {
	Limit      int         `json:"limit,omitempty;query:limit"`
	Page       int         `json:"page,omitempty;query:page"`
	Sort       string      `json:"sort,omitempty;query:sort"`
	TotalRows  int64       `json:"total_rows"`
	TotalPages int         `json:"total_pages"`
	Rows       interface{} `json:"rows"`
}

func (p *Pagination) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

func (p *Pagination) GetLimit() int {
	if p.Limit == 0 {
		p.Limit = 10
	}
	return p.Limit
}

func (p *Pagination) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}
	return p.Page
}

func (p *Pagination) GetSort() string {
	if p.Sort == "" {
		p.Sort = "Id desc"
	}
	return p.Sort
}

func Paginate(pagination *Pagination, txn *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRows int64
	txn.Count(&totalRows)
	pagination.TotalRows = totalRows
	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.Limit)))
	pagination.TotalPages = totalPages
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
	}
}
