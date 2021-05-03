package main

import (
	"io"
	"io/ioutil"

	"github.com/ferocious-space/httpcache/LruCache"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"github.com/ferocious-space/durableclient"
)

func main() {
	zaplog, _ := zap.NewDevelopment()
	log := zapr.NewLogger(zaplog).WithName("main")

	url := "https://esi.evetech.net/latest/characters/901264898/?datasource=tranquility"

	normalClient := durableclient.NewDurableClient(durableclient.OptionLogger(log.WithName("normal")))
	cacheClient := durableclient.NewDurableClient(durableclient.OptionLogger(log.WithName("cached")), durableclient.OptionCache(LruCache.NewLRUCache(1<<20*50)))

	b := int64(0)
	log.Info("NO Cache Transport")
	r, e := normalClient.Get(url)
	if e != nil {
		log.Error(e, "Get")
	} else {
		b, _ = io.Copy(ioutil.Discard, r.Body)
		log.Info("data", "bytes", b)
		_ = r.Body.Close()
	}

	log.Info("Cache Transport")
	rc, ec := cacheClient.Get(url)
	if ec != nil {
		log.Error(ec, "Get")
	} else {
		b, _ = io.Copy(ioutil.Discard, rc.Body)
		log.Info("data", "bytes", b)
		_ = rc.Body.Close()
	}

	log.Info("Cache Transport")
	rc, ec = cacheClient.Get(url)
	if ec != nil {
		log.Error(ec, "Get")
	} else {
		b, _ = io.Copy(ioutil.Discard, rc.Body)
		log.Info("data", "bytes", b)
		_ = rc.Body.Close()
	}

}
