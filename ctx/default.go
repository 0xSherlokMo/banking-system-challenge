package ctx

import (
	"log"

	"github.com/0xSherlokMo/banking-system-challenge/account"
	"github.com/0xSherlokMo/banking-system-challenge/memorydb"
	"go.uber.org/zap"
)

type Database[T memorydb.IdentifiedRecord] interface {
	Set(key string, record T, opts memorydb.Opts)
	Setnx(key string, record T) error
	GetM(terms []string, opts memorydb.Opts) []T
	Get(key string, opts memorydb.Opts) (T, error)
	Keys() []memorydb.Key
	Length() int
	Lock(key string) error
	Unlock(key string) error
}
type DefaultContext struct {
	db     Database[*account.Account]
	logger *zap.SugaredLogger
}

func NewDefaultContext() *DefaultContext {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("cannot setup logger")
	}
	sugarLogger := logger.Sugar()
	return &DefaultContext{
		logger: sugarLogger,
	}
}

func (d *DefaultContext) WithMemoryDB() *DefaultContext {
	if d.db != nil {
		return d
	}
	d.db = memorydb.Default[*account.Account]()
	return d
}

func (d *DefaultContext) MemoryDB() Database[*account.Account] {
	if d.db == nil {
		d.WithMemoryDB()
	}
	return d.db
}

func (d *DefaultContext) Logger() *zap.SugaredLogger {
	return d.logger
}

func (d *DefaultContext) Exit() {
	d.logger.Sync()
}
