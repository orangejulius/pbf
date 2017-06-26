package command

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/missinglink/pbf/handler"
	"github.com/missinglink/pbf/leveldb"
	"github.com/missinglink/pbf/lib"
	"github.com/missinglink/pbf/parser"
	"github.com/missinglink/pbf/proxy"

	"github.com/codegangsta/cli"
)

// Pelias cli command
func Pelias(c *cli.Context) error {

	// validate args
	var argv = c.Args()
	if len(argv) != 1 {
		fmt.Println("invalid arguments, expected: {pbf}")
		os.Exit(1)
	}

	// create parser
	p := parser.NewParser(argv[0])

	// -- bitmask --

	// bitmask is mandatory
	var bitmaskPath = c.String("bitmask")

	// bitmask file doesn't exist
	if _, err := os.Stat(bitmaskPath); err != nil {
		fmt.Println("bitmask file doesn't exist")
		os.Exit(1)
	}

	// debug
	log.Println("loaded bitmask:", bitmaskPath)

	// read bitmask from disk
	masks := lib.NewBitmaskMap()
	masks.ReadFromFile(bitmaskPath)

	// -- leveldb --

	// leveldb directory is mandatory
	var leveldbPath = c.String("leveldb")

	// stat leveldb destination
	lib.EnsureDirectoryExists(leveldbPath, "leveldb")

	// open database connection
	conn := &leveldb.Connection{}
	conn.Open(leveldbPath)
	defer conn.Close()

	// create parser handler
	var handle = &handler.DenormlizedJSON{
		Mutex:           &sync.Mutex{},
		Conn:            conn,
		ComputeCentroid: true,
		ExportLatLons:   false,
	}

	// create filter proxy
	var filter = &proxy.WhiteList{
		Handler:      handle,
		NodeMask:     masks.Nodes,
		WayMask:      masks.Ways,
		RelationMask: masks.Relations,
	}

	// create store proxy
	var store = &proxy.StoreRefs{
		Handler: filter,
		Conn:    conn,
		Masks:   masks,
	}

	p.Parse(store)

	// find first way offset
	// offset, err := store.Index.FirstOffsetOfType("way")
	// if nil != err {
	// 	log.Printf("target type: %s not found in file\n", "way")
	// 	os.Exit(1)
	// }

	// Parse will block until it is done or an error occurs.
	// p.ParseFrom(filter, offset)

	return nil
}
