# - makefile -
# LICENSE: GPLv2, see attached License
# Author: Joerg Jungermann

all: carbuffd carbuffd.386 carbuffd.amd64 carbuffd.arm

clean:
	rm -f carbuffd.386  carbuffd.amd64 carbuffd.arm carbuffd

carbuffd: main.go
	go build -o $@
	strip -g $@

carbuffd.386: main.go
	GOARCH=386   go build -o $@
	strip -g $@

carbuffd.amd64: main.go
	GOARCH=amd64 go build -o $@
	strip -g $@

carbuffd.arm: main.go
	GOARCH=arm   go build -o $@
	strip -g $@

# vim: noet
