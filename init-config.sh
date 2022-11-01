#!/usr/bin/env sh

CONFDIR="$HOME/.config/ethlogfilter";
CONFFILE="$CONFDIR/config.yaml";

mkdir -v -p $CONFDIR;

test -f $CONFFILE && exit 0;
cp -v config.yaml.example "$CONFFILE";
