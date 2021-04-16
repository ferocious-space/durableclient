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
	vctx := context.WithValue(context.Background(), "test", "value")
	dc := durableclient.NewDurableClient(vctx, log.WithName("httpClient"), "test")
	clonedC := dc.Clone()

	ccache := dc.WithCache(httpcache.NewLRUCache(1<<20*50, 600)).Client()
	c := clonedC.WithPool(true).Client(durableclient.WithContext(context.Background()), durableclient.WithRetrier())

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
	rc, ec = ccache.Get(url)
	if ec != nil {
		log.Error(ec, "Get")
	} else {
		b, _ = io.Copy(ioutil.Discard, rc.Body)
		log.Info("data", "bytes", b)
		_ = rc.Body.Close()
	}

}
