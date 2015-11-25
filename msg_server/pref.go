package main

import (
	log "goProject/log"
	"net/http"
	_ "net/http/pprof"
)

// StartPprof start a golang pprof.
func StartPprof(addr string) {
	go func() {
		var err error
		if err = http.ListenAndServe(addr, nil); err != nil {
			log.Errorf("http.ListenAndServe(\"%s\") error(%v)", addr, err)
			return
		}
	}()
}
