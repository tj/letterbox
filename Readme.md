# Letterbox

A tiny Go program to batch-process letter-boxing of photographs.

## Installation

From [gobinaries.com](https://gobinaries.com):

```sh
$ curl -sf https://gobinaries.com/tj/letterbox/cmd/letterbox | sh
```

From source:

```
$ go get github.com/tj/letterbox/cmd/letterbox
```

## Usage

```
Usage of letterbox:
  -aspect string
    	Output aspect ratio (default "16:9")
  -concurrency int
    	Concurrency of image processing (default 8)
  -force
    	Force image reprocess when it exists
  -output string
    	Image output directory (default "processed")
  -padding int
    	Output image padding in percentage
  -quality int
    	Output jpeg quality (default 90)
  -white
    	Output a white letterbox
```

## Examples

Example of 1:1

```
$ letterbox -aspect 1:1
```

![](https://apex-software.imgix.net/github/tj/letterbox/1-1.jpg?w=500&dpr=2)

Example of 4:3

```
$ letterbox -aspect 4:3
```

![](https://apex-software.imgix.net/github/tj/letterbox/4-3.jpg?w=500&dpr=2)

Example of 16:9 (the default)

```
$ letterbox -aspect 16:9
```

![](https://apex-software.imgix.net/github/tj/letterbox/16-9.jpg?w=500&dpr=2)

Example of explicitly listing images:

```
$ letterbox DSCF6719.jpg DSCF6718.jpg
```

![](https://apex-software.imgix.net/github/tj/letterbox/16-9.jpg?w=500&dpr=2)

Example of 1:1 with a white background and 6% padding:

```
$ letterbox -white -aspect 1:1 -padding 6
```

![](https://apex-software.imgix.net/github/tj/letterbox/1-1-white.jpg?w=500&dpr=2)

---

[![GoDoc](https://godoc.org/github.com/tj/letterbox?status.svg)](https://godoc.org/github.com/tj/letterbox)
![](https://img.shields.io/badge/license-MIT-blue.svg)
![](https://img.shields.io/badge/status-stable-green.svg)

<a href="https://apex.sh"><img src="http://tjholowaychuk.com:6000/svg/sponsor"></a>
