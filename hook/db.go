package main

import (
	"log"
	"time"

	"github.com/dgraph-io/badger/v3"
)

func openDB(dir string) (*badger.DB, error) {
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			lsm, vlog := db.Size()
			log.Printf("Badger LSM %d VLOG %d\n", lsm, vlog)
			if lsm > 1024*1024*8 || vlog > 1024*1024*32 {
				err := db.RunValueLogGC(0.5)
				log.Printf("Badger RunValueLogGC %v\n", err)
			}
			time.Sleep(5 * time.Minute)
		}
	}()

	return db, nil
}
