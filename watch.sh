while true; do
  sleep .25
  #echo === Scanning... ===
  if [ -n "$(find . -name '*.go' -newer ./tdop -print | head -n 1)" ]; then
    clear
    echo Reformatting...
    find . -name '*.go' -newer ./tdop -print -exec go fmt '{}' \;
    echo Rebuilding...
    go generate && go test ./... && go build && ./tdop
    touch ./tdop
  fi
done
