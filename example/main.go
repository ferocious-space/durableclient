package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ferocious-space/httpcache"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ferocious-space/durableclient"
)

func main() {
	zaplog, _ := zap.NewDevelopment()
	log := zapr.NewLogger(zaplog).WithName("main")

	url := "https://esi.evetech.net/latest/characters/90126489/?datasource=tranquility"
	dc := durableclient.NewDurableClient(log.WithName("httpClient"), "test", durableclient.OptionRetrier())

	cacheClient := dc.Clone(durableclient.OptionCache(httpcache.NewLRUCache(1<<20*50, 600)))
	retryClient := dc.Clone(durableclient.OptionConnectionPooling())
	downloadClient := dc.Clone()

	dc.Get("https://esi.evetech.net/latest/characters/90126489/?datasource=tranquility")
	f, err := os.Create("sde.zip")
	if err != nil {
		log.Error(err, "poop creating sde.zip")
		return
	}
	downloadSize, err := downloadClient.Download(
		"https://eve-static-data-export.s3-eu-west-1.amazonaws.com/tranquility/sde.zip",
		f,
	)
	if err != nil {
		log.Error(err, "donwload failed")
		return
	}
	zr, err := zip.NewReader(f, downloadSize)
	if err != nil {
		log.Error(err, "poop zip reader")
		return
	}
	defer f.Close()
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			log.Error(err, "cant open file", "name", f.Name)
			return
		}
		fpath := filepath.Join("./", "staticdata", f.Name)
		if !strings.HasPrefix(
			fpath,
			filepath.Clean(filepath.Join("./", "staticdata", "sde")+string(os.PathSeparator)),
		) {
			log.Error(fmt.Errorf("%s: illegal file path", fpath), "dunno")
			return
		}
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(fpath, os.ModePerm)
			rc.Close()
			if err != nil {
				log.Error(err, "create dir", "dir", fpath)
				return
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				log.Error(err, "create file dir", "dir", fpath)
				return
			}
			out, err := os.Create(fpath)
			if err != nil {
				log.Error(err, "create file", "file", fpath)
				return
			}
			wSize, err := io.Copy(out, rc)
			_ = out.Close()
			_ = rc.Close()
			if err != nil {
				log.Error(err, "write file", "file", fpath)
				return
			}
			if uint64(wSize) != f.UncompressedSize64 {
				log.Error(errors.New("file size dot match"), "decompress", "file", fpath)
			}
			log.Info("unzipped", "file", fpath)
		}
	}

	b := int64(0)
	log.Info("NO Cache Transport")
	r, e := retryClient.Get(url)
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
