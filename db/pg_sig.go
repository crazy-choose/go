package db

import (
	"errors"
	"github.com/crazy-choose/go/log"
	"github.com/crazy-choose/go/meta"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	cfgSig           meta.Postgres
	impl             *Session
	ErrUninitialized = errors.New("pg client not initialized")
)

func InitPgSig(opt meta.Postgres) bool {
	if !verifyConfig(opt) {
		return false
	}
	cfgSig = opt
	pgConfig := postgres.Config{
		PreferSimpleProtocol: true,
		DSN:                  genDSN(opt), // DSN data source name
	}
	if db, err := gorm.Open(postgres.New(pgConfig)); err != nil {
		log.Error(err.Error())
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(opt.MaxIdleConns)
		sqlDB.SetMaxOpenConns(opt.MaxOpenConns)
		impl = &Session{
			db,
		}
		log.Info("InitPg success")
	}
	return true
}

func PostgresSig() *Session {
	if impl == nil {
		InitPgSig(cfgSig)
	}
	return impl
}

//func (s *Session) Transaction(fc func(tx *Session) error, opts ...*sql.TxOptions) error {
//	return s.DB.Transaction(func(tx *gorm.DB) error {
//		return fc(&Session{tx})
//	}, opts...)
//}
//
//func Transaction(fc func(tx *Session) error, opts ...*sql.TxOptions) (err error) {
//	if impl == nil {
//		Init(cfg)
//	}
//	return impl.DB.Transaction(func(tx *gorm.DB) error {
//		return fc(&Session{tx})
//	}, opts...)
//}
//
//type Pagination struct {
//	Limit      int         `json:"limit,omitempty;query:limit"`
//	Page       int         `json:"page,omitempty;query:page"`
//	Sort       string      `json:"sort,omitempty;query:sort"`
//	TotalRows  int64       `json:"total_rows"`
//	TotalPages int         `json:"total_pages"`
//	Rows       interface{} `json:"rows"`
//}
//
//func (p *Pagination) GetOffset() int {
//	return (p.GetPage() - 1) * p.GetLimit()
//}
//
//func (p *Pagination) GetLimit() int {
//	if p.Limit == 0 {
//		p.Limit = 10
//	}
//	return p.Limit
//}
//
//func (p *Pagination) GetPage() int {
//	if p.Page == 0 {
//		p.Page = 1
//	}
//	return p.Page
//}
//
//func (p *Pagination) GetSort() string {
//	if p.Sort == "" {
//		p.Sort = "Id desc"
//	}
//	return p.Sort
//}
//
//func Paginate(pagination *Pagination, txn *gorm.DB) func(db *gorm.DB) *gorm.DB {
//	var totalRows int64
//	txn.Count(&totalRows)
//	pagination.TotalRows = totalRows
//	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.Limit)))
//	pagination.TotalPages = totalPages
//	return func(db *gorm.DB) *gorm.DB {
//		return db.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
//	}
//}
