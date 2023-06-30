package main

import (
	"bytes"
	"github.com/ferocious-space/httpcache/LruCache"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"time"

	"github.com/ferocious-space/durableclient"
)

func main() {
	srv := &http.Server{
		Handler: http.DefaultServeMux,
		Addr:    ":8090",
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			return
		}
	}()
	zaplog, _ := zap.NewDevelopment()
	log := zapr.NewLogger(zaplog).WithName("main")

	url := "https://esi.evetech.net/latest/characters/90126489/?datasource=tranquility"

	normalClient := durableclient.NewDurableClient(durableclient.OptionLogger(log.WithName("normal")))
	cacheClient := durableclient.NewDurableClient(
		durableclient.OptionLogger(log.WithName("cached")),
		durableclient.OptionCache(LruCache.NewLRUCache(1<<20*50)),
		durableclient.OptionConnectionPooling(true),
	)

	go func() {
		for {
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
				b, _ = io.Copy(io.Discard, rc.Body)
				log.Info("data", "bytes", b)
				_ = rc.Body.Close()
			}

			log.Info("Cache Transport")
			rc, ec = cacheClient.Get(url)
			if ec != nil {
				log.Error(ec, "Get")
			} else {
				b, _ = io.Copy(io.Discard, rc.Body)
				log.Info("data", "bytes", b)
				_ = rc.Body.Close()
			}
			rc, ec = cacheClient.Post("https://esi.evetech.net/latest/universe/ids/?datasource=tranquility&language=en", "application/json", bytes.NewBufferString("[ \"CCP Zoetrope\"]"))
			if ec != nil {
				log.Error(ec, "Get")
			} else {
				b, _ = io.Copy(io.Discard, rc.Body)
				log.Info("data", "bytes", b)
				_ = rc.Body.Close()
			}
			time.Sleep(30 * time.Second)
		}
	}()

	for {
		runtime.Gosched()
	}
}
