package cron

import (
	"fmt"
	"github.com/crazy-choose/go/log"
	"github.com/robfig/cron/v3"
	"os"
	"time"
)

type CustomCron struct {
	cron      *cron.Cron
	entryMap  []cron.EntryID
	entryComm []string
}

var impl *CustomCron

func NewCron() *CustomCron {
	nyc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil
	}
	_cron := cron.New(cron.WithSeconds(), cron.WithLocation(nyc))
	impl = &CustomCron{
		cron:      _cron,
		entryMap:  make([]cron.EntryID, 0),
		entryComm: make([]string, 0),
	}

	return impl
}

func Impl() *CustomCron {
	if impl == nil {
		NewCron()
	}
	return impl
}

func (impl *CustomCron) AddFunc(f func(), command string) {
	if impl.cron != nil {
		c, e := impl.cron.AddFunc(command, f)
		if e != nil {
			log.Info("CustomCron AddFun err:", e)
			os.Exit(0)
		}
		impl.entryMap = append(impl.entryMap, c)
		impl.entryComm = append(impl.entryComm, command)
	}
}

func (impl *CustomCron) Start() {
	impl.cron.Start()
	log.Info("[ CustomCron ] Start...")
}

func (impl *CustomCron) Stop() {
	impl.cron.Stop()
	log.Info("[ CustomCron ] Stop...")
}

func (impl *CustomCron) Print() {
	fmt.Printf("[CustomCron]entryComm:%v", impl.entryComm)
	fmt.Printf("[CustomCron]entryMap:%v", impl.entryMap)
}

func (impl *CustomCron) Log() {
	log.Info("[CustomCron]entryComm:%v", impl.entryComm)
	log.Info("[CustomCron]entryMap:%v", impl.entryMap)
}
