package ctx

import (
	"log"

	"github.com/0xSherlokMo/banking-system-challenge/account"
	"github.com/0xSherlokMo/banking-system-challenge/memorydb"
	"go.uber.org/zap"
)

type DefaultContext struct {
	db     *memorydb.MemoryDB[*account.Account]
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

func (d *DefaultContext) MemoryDB() *memorydb.MemoryDB[*account.Account] {
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
