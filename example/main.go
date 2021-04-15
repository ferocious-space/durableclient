package main

import (
	"context"
	"io"
	"io/ioutil"

	"github.com/ferocious-space/durableclient"
	"github.com/ferocious-space/httpcache"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

func main() {
	zaplog, _ := zap.NewDevelopment()
	log := zapr.NewLogger(zaplog).WithName("main")

	url := "https://esi.evetech.net/latest/characters/90126489/?datasource=tranquility"
	dc := durableclient.NewDurableClient(context.Background(), log.WithName("httpClient"), "test")
	c := dc.Client()
	ccache := dc.WithCache(httpcache.NewLRUCache(1<<20*50, 600)).Client()
	cc := dc.WithPool(true).Client()

	b := int64(0)
	log.Info("NO Cache Transport")
	r, e := c.Get(url)
	if e != nil {
		log.Error(e, "Get")
	} else {
		b, _ = io.Copy(ioutil.Discard, r.Body)
		log.Info("data", "bytes", b)
		_ = r.Body.Close()
	}

	log.Info("Cache Transport")
	rc, ec := ccache.Get(url)
	if ec != nil {
		log.Error(ec, "Get")
	} else {
		b, _ = io.Copy(ioutil.Discard, rc.Body)
		log.Info("data", "bytes", b)
		_ = rc.Body.Close()
	}

	log.Info("Cache Transport")
	rc, ec = cc.Get(url)
	if ec != nil {
		log.Error(ec, "Get")
	} else {
		b, _ = io.Copy(ioutil.Discard, rc.Body)
		log.Info("data", "bytes", b)
		_ = rc.Body.Close()
	}

}
