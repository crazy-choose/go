package custom

import (
	"github.com/crazy-choose/go/log"
	"github.com/robfig/cron/v3"
	"os"
	"time"
)

type Cron struct {
	cron      *cron.Cron
	entryMap  []cron.EntryID
	entryComm []string
}

var impl *Cron

func NewCron() *Cron {
	nyc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil
	}
	_cron := cron.New(cron.WithSeconds(), cron.WithLocation(nyc))
	impl = &Cron{
		cron:      _cron,
		entryMap:  make([]cron.EntryID, 0),
		entryComm: make([]string, 0),
	}

	return impl
}

func Impl() *Cron {
	if impl == nil {
		NewCron()
	}
	return impl
}

func (impl *Cron) AddFunc(f func(), command string) {
	if impl.cron != nil {
		c, e := impl.cron.AddFunc(command, f)
		if e != nil {
			log.Info("Cron AddFun err:", e)
			os.Exit(0)
		}
		impl.entryMap = append(impl.entryMap, c)
		impl.entryComm = append(impl.entryComm, command)
	}
}

func (impl *Cron) Start() {
	impl.cron.Start()
	log.Info("[ Cron ] Start...")
}

func (impl *Cron) Stop() {
	impl.cron.Stop()
	log.Info("[ Cron ] Stop...")
}

func (impl *Cron) Log() {
	log.Info("[Cron]entryComm:%v", impl.entryComm)
	log.Info("[Cron]entryMap:%v", impl.entryMap)
}
