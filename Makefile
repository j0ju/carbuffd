# - makefile -
# LICENSE: GPLv2, see attached License
# Author: Joerg Jungermann

INSTDIR = /opt/go/bin

all: \
  carbuffd

clean:
	rm -f carbuffd.386  carbuffd.amd64 carbuffd.arm carbuffd

install: carbuffd $(INSTDIR) $(INSTDIR)/carbuffd

$(INSTDIR):
	install -o root -g root -d $@

$(INSTDIR)/carbuffd: carbuffd /etc/systemd/system/carbuffd.service /etc/carbuffd.env
	install -o root -g root -s $< $@

/etc/systemd/system/carbuffd.service: carbuffd.service
	install -m 0644 -o root -g root $< $@

/etc/carbuffd.env: carbuffd.env
	[ -f /etc/carbuffd.env ] || \
		install -m 0644 -o root -g root $< $@

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
