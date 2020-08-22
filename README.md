# HTTPServant
[![License](https://img.shields.io/badge/license-MIT-_red.svg)](https://opensource.org/licenses/MIT)
[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/rykkard/httpservant/issues)

Small tool made in Go to serve files over HTTP which provides more verbose log output.

## Install

```bash
$ go get -v 'github.com/rykkard/httpservant'
```

## Update

```bash
$ go get -v -u 'github.com/rykkard/httpservant'
```

## Usage

```
$ httpservant -h
Small tool to serve just one directory or file over HTTP.
It serves the current directory by default.

Usage:
   httpservant [OPTIONS] <filename|directory>

Options:
   -p, --port <port>           port to serve on (default: 8000)
   -b, --bind <interface>      interface to bind (default: 0.0.0.0)
   --cors                      enable cors
   --list                      enable listing on root paths (/)
   --auth <user:pass>          enable basic authentication
   -s, --silent                enable silent mode
   -h, --help                  show help
```
## Example

```
$ httpservant
[*] Stagging directory: .
.

[*] Serving HTTP on 0.0.0.0 port 8000
127.0.0.1 - - [13/Aug/2020:17:45:34 -0500] "GET / HTTP/1.1" 200 10
127.0.0.1 - - [13/Aug/2020:17:46:24 -0500] "POST / HTTP/1.1" 200 10
Hello world!

127.0.0.1 - - [13/Aug/2020:17:46:47 -0500] "POST / HTTP/1.1" 200 10
+-----------------------------------------+
| NOTE: binary data not shown in terminal |
+-----------------------------------------+
127.0.0.1 - - [13/Aug/2020:17:49:36 -0500] "POST / HTTP/1.1" 200 10
GG

^C
[*] Shutdown HTTP service
```
