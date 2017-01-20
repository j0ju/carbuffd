#!/bin/sh /etc/rc.common

# LICENSE: GPLv2, see attached License
# Author: Joerg Jungermann

# carbuffd initscript for OpenWRT/LEDE using procd

START=99
STOP=10
USE_PROCD=1
EXTRA_COMMANDS="daemon"

LOGFILE=syslog
LOGLEVEL=3
INTERVAL=60
SIZE=10000
LADDR=:2003
# if this is not set in /etc/carbuffd.env, it just consumes the events
RADDR= # 203.0.113.23:2003

ENVFILE="/etc/carbuffd.env"
BIN="/opt/go/bin/carbuffd"

[ -r "$ENVFILE" ] && \
  . "$ENVFILE"

start_service() {
	procd_open_instance
	procd_set_param command "$initscript" daemon
	procd_set_param respawn
	procd_close_instance
}

daemon() {
	logger -t "${initscript##*/}:" "starting ..."
	exec "$BIN" -logfile "$LOGFILE" -loglevel "$LOGLEVEL" -interval "$INTERVAL" -size "$SIZE" "$LADDR" $RADDR
}
