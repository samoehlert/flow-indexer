package flowindexer

import (
	"log"
	"path/filepath"
	"runtime"
	"time"

	"github.com/JustinAzoff/flow-indexer/backend"
	"github.com/JustinAzoff/flow-indexer/ipset"
	"github.com/JustinAzoff/flow-indexer/store"
)

func Index(s store.IpStore, b backend.Backend, filename string) error {
	exists, err := s.HasDocument(filename)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("%s Already indexed\n", filename)
		return nil
	}
	ips := ipset.New()
	reader, err := backend.OpenDecompress(filename)
	if err != nil {
		return err
	}
	defer reader.Close()

	start := time.Now()
	lines, err := b.ExtractIps(reader, ips)
	duration := time.Since(start)
	log.Printf("%s: Read %d lines in %s\n", filename, lines, duration)

	if err != nil {
		log.Printf("%s: Non fatal read error: %s\n", filename, err)
	}
	start = time.Now()
	s.AddDocument(filename, *ips)
	duration = time.Since(start)
	log.Printf("%s: Wrote %d unique ips in %s\n", filename, len(ips.Store), duration)
	return nil
}

func RunIndex(dbpath string, args []string) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	mystore, err := store.NewStore("leveldb", dbpath)
	check(err)
	defer mystore.Close()

	for _, path := range args {
		mybackend := backend.NewBackend("bro")
		matches, err := filepath.Glob(path)
		check(err)
		for _, fp := range matches {
			err = Index(mystore, mybackend, fp)
			check(err)
		}
	}
}
