package main

import (
	"fmt"
	"net/http/httputil"

	"github.com/ferocious-space/httpcache"
	"go.uber.org/zap"

	"github.com/ferocious-space/durableclient"
)

func main() {
	lgr, _ := zap.NewDevelopment()
	cli := durableclient.NewCachedClient("test", httpcache.NewLRUCache(1<<20*50, 600), lgr.Named("cached"))
	cli2 := durableclient.NewClient("test2", lgr.Named("noncached"))
	a, e := cli.Get("https://scalewp.io")
	if e != nil {
		lgr.Panic("", zap.Error(e))
	}
	bin, _ := httputil.DumpResponse(a, false)
	fmt.Println(string(bin))
	b, e := cli.Get("https://scalewp.io")
	if e != nil {
		lgr.Panic("", zap.Error(e))
	}
	bin, _ = httputil.DumpResponse(b, false)
	fmt.Println(string(bin))
	c, e := cli2.Get("https://scalewp.io")
	if e != nil {
		lgr.Panic("", zap.Error(e))
	}
	bin, _ = httputil.DumpResponse(c, false)
	fmt.Println(string(bin))
}
