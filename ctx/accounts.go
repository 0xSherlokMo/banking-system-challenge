package ctx

import (
	"encoding/json"
	"net/http"

	"github.com/0xSherlokMo/banking-system-challenge/account"
)

const (
	jsonFileURL = "https://gist.githubusercontent.com/paytabs-engineering/c470210ebb19511a4e744aefc871974f/raw/6296df58428c89b8f852a6a83b0a5d0ac38289b6/accounts-mock.json"
)

func (d *DefaultContext) LoadAccounts() *DefaultContext {
	d.Logger().Infow("Loading accounts", "url", jsonFileURL)
	res, err := http.Get(jsonFileURL)
	if err != nil {
		d.Logger().Fatalw(
			"cannot load accounts",
			"error", err,
		)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		d.Logger().Fatalw(
			"unexpected http GET status",
			"status", res.StatusCode,
		)
	}

	var accounts []account.Account
	err = json.NewDecoder(res.Body).Decode(&accounts)
	if err != nil {
		d.Logger().Fatalw(
			"cannot decode accounts",
			"error", err,
		)
	}

	for _, account := range accounts {
		d.MemoryDB().Setnx(account.GetID(), &account)
	}

	d.Logger().Infow("Loaded accounts in memory", "accounts", d.MemoryDB().Length())

	return d
}
