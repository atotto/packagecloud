package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	packagecloud "github.com/atotto/packagecloud/api/v1"
	"github.com/google/subcommands"
)

var (
	PACKAGECLOUD_TOKEN = os.Getenv("PACKAGECLOUD_TOKEN")
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(pushPackageCommand, "package")
	subcommands.Register(deletePackageCommand, "package")
	subcommands.Register(promotePackageCommand, "package")

	flag.Parse()

	if PACKAGECLOUD_TOKEN == "" {
		fmt.Fprintf(flag.CommandLine.Output(), `
Please set an environment variable with the name PACKAGECLOUD_TOKEN, containing the value of a packagecloud API token.
You can find a packagecloud API token at https://packagecloud.io/api_token .
`)
		log.Println(`PACKAGECLOUD_TOKEN is empty`)
		os.Exit(2)
	}
	ctx = packagecloud.WithPackagecloudToken(ctx, PACKAGECLOUD_TOKEN)

	go func() {
		os.Exit(int(subcommands.Execute(ctx)))
	}()

	select {
	case <-sig:
		cancel()
	case <-ctx.Done():
	}
}
