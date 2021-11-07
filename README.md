# HTTPserve
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/rykkard/httpserve)
[![License](https://img.shields.io/badge/license-MIT-_red.svg)](https://opensource.org/licenses/MIT)
[![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/rykkard/httpserve/issues)

Small tool to serve files over HTTP which provides more verbose log output.

## Installation

```bash
$ go install -v 'github.com/rykkard/httpserve@latest'
```

## Usage

```
$ httpserve -h
Small tool to serve files/directories over HTTP with more verbosity.

Usage:
   httpserve [OPTIONS] <filenames|directories>

Options:
   -p, --port <port>           port to serve on (default: 8000)
   -b, --bind <interface>      interface to bind (default: 0.0.0.0)
   --cors                      enable cors
   --list                      enable listing on root paths (/)
   --auth <user:pass>          enable basic authentication
   -v, --verbose               enable more verbose (headers)
   -s, --silent                enable silent mode
   -h, --help                  show help
```

## Examples

- Just listen mode
```
$ httpserve
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
<..>
```
- Serving files
```
$ httpserve LICENSE.md README.md
[*] Stagging resources
.
└── LICENSE.md
└── README.md
[*] Serving HTTP on 0.0.0.0 port 8000
<..>
```

- Serving directory files with any other file
```
$ httpserve /dev/shm/ README.md
[*] Stagging resources
/dev/shm
└── /dev/shm/somefile
.
└── README.md
[*] Serving HTTP on 0.0.0.0 port 8000
<..>
```
