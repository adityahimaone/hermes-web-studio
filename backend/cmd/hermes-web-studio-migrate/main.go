package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/adityahimaone/hermes-web-studio/backend/internal/migrate"
)

func main() {
	state := flag.String("state-dir", "", "Hermes WebUI state directory")
	restore := flag.Bool("restore-latest", false, "restore the newest backup")
	flag.Parse()
	if *state == "" {
		log.Fatal("-state-dir is required")
	}
	var path string
	var err error
	if *restore {
		path, err = migrate.RestoreLatest(*state)
	} else {
		path, err = migrate.Backup(*state)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(os.Stdout, path)
}
