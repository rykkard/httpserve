package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"unicode"

	"github.com/disiqueira/gotree"
	"github.com/goji/httpauth"
	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
)

const (
	version        string = "v1.4.15"
	message200     string = "It works !"
	message404     string = "404 page not found"
	messageBinData string = "" +
		"+-----------------------------------------+\n" +
		"| NOTE: binary data not shown in terminal |\n" +
		"+-----------------------------------------+"
)

var (
	port          uint   = 8000
	bindInterface string = "0.0.0.0"
	corsEnable    bool   = false
	listEnable    bool   = false
	silentMode    bool   = false
	authString    string = ""
	limit         uint64 = ^uint64(0)
	content       string = "."
)

func init() {
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

	flag.UintVar(&port, "port", port, "")
	flag.UintVar(&port, "p", port, "")

	flag.StringVar(&bindInterface, "bind", bindInterface, "")
	flag.StringVar(&bindInterface, "b", bindInterface, "")

	flag.BoolVar(&corsEnable, "cors", corsEnable, "")
	flag.BoolVar(&listEnable, "list", listEnable, "")

	flag.BoolVar(&silentMode, "silent", silentMode, "")
	flag.BoolVar(&silentMode, "s", silentMode, "")

	flag.StringVar(&authString, "auth", authString, "")

	//TODO server shutdown based on request limit?
	//flag.Uint64Var(&limit, "limit", limit, "")
	//flag.Uint64Var(&limit, "l", limit, "")

	flag.Parse()
}

func main() {
	if flag.Arg(0) != "" {
		content = filepath.Clean(flag.Arg(0))
	}

	fi, err := os.Stat(content)
	if err != nil {
		log.Fatalln(err)
	}

	handlerChain := alice.New(requestHandler)

	// enable silent mode
	log.SetFlags(0)
	if silentMode {
		log.SetOutput(ioutil.Discard)
	} else {
		loggingHandler := createLoggingHandler(os.Stdout)
		handlerChain = handlerChain.Append(loggingHandler)
	}

	// enable basic authentication
	if len(authString) != 0 {
		creds := strings.SplitN(authString, ":", 2)
		user := creds[0]
		pass := ""
		if len(creds) == 2 {
			pass = creds[1]
		}
		authHandler := httpauth.SimpleBasicAuth(user, pass)
		handlerChain = handlerChain.Append(authHandler)
	}

	handlerChain = handlerChain.Append(responseHandler)

	mux := http.NewServeMux()
	switch mode := fi.Mode(); {
	case mode.IsDir():
		log.Println("[*] Stagging directory:", content)
		treeView := gotree.New(content)
		_ = filepath.Walk(content,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.Mode().IsRegular() {
					data := fmt.Sprintf("%v", path)
					treeView.Add(data)
				}
				return nil
			})
		log.Println(treeView.Print())
		fileServer := http.FileServer(http.Dir(content))
		mux.Handle("/", handlerChain.Then(fileServer))
	case mode.IsRegular():
		log.Println("[*] Stagging file:", content)
		treeView := gotree.New(path.Dir(content))
		data := fmt.Sprintf("%v", content)
		treeView.Add(data)
		log.Println(treeView.Print())
		pattern := fmt.Sprint("/", filepath.Base(content))

		mux.Handle("/", handlerChain.ThenFunc(welcome))
		mux.Handle(pattern, handlerChain.ThenFunc(serveFile))
	default:
		mux.Handle("/", handlerChain.ThenFunc(welcome))
	}

	log.Printf("[*] Serving HTTP on %v port %v\n", bindInterface, port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	server := http.Server{Addr: fmt.Sprint(bindInterface, ":", port), Handler: mux}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	select {
	case <-ctx.Done():
	case <-sig:
	}
	log.Println("\n[*] Shutdown HTTP service")
	server.Shutdown(ctx)
}

func requestHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
		}

		next.ServeHTTP(w, r)

		if len(body) > 0 {
			data := string(body)
			if isASCIIPrintable(data) {
				log.Println(data)
			} else {
				log.Println(messageBinData)
			}
		}
	})
}

func responseHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if corsEnable {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		if !listEnable && strings.HasSuffix(r.URL.Path, "/") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, message200)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func createLoggingHandler(dst io.Writer) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return handlers.LoggingHandler(dst, h)
	}
}

func welcome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, message200)
}

func serveFile(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, content)
}

func isASCIIPrintable(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}
