#!/usr/bin/env bash
while true; do
  sleep .25
  #echo === Scanning... ===
  if [[  ( ! -f ./tdop ) || -n "$(find . -name '*.go' -newer ./tdop -print | head -n 1)" ]]; then
    clear
    echo Reformatting...
    find . -name '*.go' -newer ./tdop -print -exec go fmt '{}' \;
    echo Rebuilding...
    go fmt ./... && go generate ./... \
      && go build && echo Build succeeded \
      && go test ./... && echo Tests passed \
      && echo && echo && echo && ./tdop
    touch ./tdop
  fi
done
