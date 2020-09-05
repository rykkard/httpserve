package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type arguments struct {
	port          uint
	bindInterface string
	corsEnable    bool
	listEnable    bool
	silentMode    bool
	authString    string
	limit         uint64
	resource      string
}

func (args *arguments) parse() {
	flag.Usage = func() {
		h := []string{
			fmt.Sprintf("HTTPServant %v", version),
			"Small tool to serve just one directory or file over HTTP.",
			"It serves the current directory by default.",
			"",
			"Usage:",
			fmt.Sprintf("   %v [OPTIONS] <filename|directory>", filepath.Base(os.Args[0])),
			"",
			"Options:",
			"   -p, --port <port>           port to serve on (default: 8000)",
			"   -b, --bind <interface>      interface to bind (default: 0.0.0.0)",
			"   --cors                      enable cors",
			"   --list                      enable listing on root paths (/)",
			"   --auth <user:pass>          enable basic authentication",
			"   -s, --silent                enable silent mode",
			"   -h, --help                  show help",
			"",
		}
		fmt.Fprintf(os.Stderr, strings.Join(h, "\n"))
		os.Exit(1)
	}

	flag.UintVar(&args.port, "port", args.port, "")
	flag.UintVar(&args.port, "p", args.port, "")

	flag.StringVar(&args.bindInterface, "bind", args.bindInterface, "")
	flag.StringVar(&args.bindInterface, "b", args.bindInterface, "")

	flag.BoolVar(&args.corsEnable, "cors", args.corsEnable, "")
	flag.BoolVar(&args.listEnable, "list", args.listEnable, "")

	flag.BoolVar(&args.silentMode, "silent", args.silentMode, "")
	flag.BoolVar(&args.silentMode, "s", args.silentMode, "")

	flag.StringVar(&args.authString, "auth", args.authString, "")

	//TODO server shutdown based on request limit?
	//flag.Uint64Var(&args.limit, "limit", args.limit, "")
	//flag.Uint64Var(&args.limit, "l", args.limit, "")

	flag.Parse()
}

var args arguments

func init() {
	args.port = 8000
	args.bindInterface = "0.0.0.0"
	args.limit = ^uint64(0)
	args.resource = "."
	args.parse()
	if flag.Arg(0) != "" {
		args.resource = filepath.Clean(flag.Arg(0))
	}
}
