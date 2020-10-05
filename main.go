package main

import (
	"context"
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
	"github.com/h2non/filetype"
	"github.com/jedib0t/go-pretty/table"
	"github.com/justinas/alice"
)

func main() {
	handlerChain := alice.New(requestHandler)

	// enable silent mode
	log.SetFlags(0)
	if args.silentMode {
		log.SetOutput(ioutil.Discard)
	} else {
		loggingHandler := createLoggingHandler(os.Stdout)
		handlerChain = handlerChain.Append(loggingHandler)
	}

	// enable basic authentication
	if len(args.authString) != 0 {
		creds := strings.SplitN(args.authString, ":", 2)
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
	if len(args.resources) == 0 {
		mux.Handle("/", handlerChain.ThenFunc(welcome))
	} else {
		log.Println("[*] Stagging resources")
		graphs := make(map[string]gotree.Tree)

		for _, resource := range args.resources {
			fi, err := os.Stat(resource)
			if err != nil {
				continue
			}

			switch mode := fi.Mode(); {
			case mode.IsDir():
				func() {
					defer func() {
						if err := recover(); err != nil {
							//log.Println("panic occurred:", err)
						}
					}()
					fileServer := http.FileServer(http.Dir(resource))
					mux.Handle("/", handlerChain.Then(fileServer))

					//build dir graph
					dirname := filepath.Clean(resource)
					graph, exists := graphs[dirname]
					if !exists {
						graph = gotree.New(dirname)
						graphs[dirname] = graph
					}
					_ = filepath.Walk(resource,
						func(path string, info os.FileInfo, err error) error {
							if err != nil {
								return err
							}
							if info.Mode().IsRegular() {
								data := fmt.Sprintf("%v", path)
								graph.Add(data)
							}
							return nil
						})
				}()
			case mode.IsRegular():
				func() {
					defer func() {
						if err := recover(); err != nil {
							//log.Println("panic occurred:", err)
						}
					}()
					pattern := fmt.Sprint("/", filepath.Base(resource))
					mux.Handle(pattern, handlerChain.ThenFunc(createServeFileHandler(resource)))

					//build file graph
					dirname := path.Dir(resource)
					data := fmt.Sprintf("%v", resource)
					graph, exists := graphs[dirname]
					if !exists {
						graph = gotree.New(dirname)
						graphs[dirname] = graph
					}
					graph.Add(data)
				}()
			default:
			}
		}
		for _, graph := range graphs {
			log.Print(graph.Print())
		}
	}

	log.Printf("[*] Serving HTTP on %v port %v\n", args.bindInterface, args.port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	server := http.Server{Addr: fmt.Sprint(args.bindInterface, ":", args.port), Handler: mux}
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

		if args.verboseEnable {
			log.Printf("Host: %v\n", r.Host)
			for header, values := range r.Header {
				for _, value := range values {
					log.Printf("%v: %v\n", header, value)
				}
			}
			log.Println()
		}

		if len(body) > 0 {
			kind, _ := filetype.Match(body)
			if kind != filetype.Unknown {
				// kind.Extension, kind.MIME.Value
				message := simpleRowRender(strings.Replace(messageBinData, "%MIME%", kind.MIME.Value, -1))
				log.Println(message)
			} else {
				log.Println(string(body))
			}
		}
	})
}

func responseHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if args.corsEnable {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		if !args.listEnable && strings.HasSuffix(r.URL.Path, "/") {
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

func createServeFileHandler(filename string) func(w http.ResponseWriter, r *http.Request) {
	fileHandler := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	}
	return fileHandler
}

func welcome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, message200)
}

func isASCIIPrintable(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func simpleRowRender(info string) string {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.AppendRows([]table.Row{
		{info},
	})
	text := t.Render()
	return text
}
