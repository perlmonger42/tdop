tdop
========================================================================

This project is an exploration of Top Down Operator Precedence parsing.
It will be seeded with a version of [Douglas Crockford's demonstration](
http://crockford.com/javascript/tdop/index.html), translated into Go from the
original JavaScript. It won't be good Go style in the beginning, because it
will be a fairly direct transliteration of Crockford's JavaScript.

From there I hope to grow it into a real Go program that can be used as a base
for further research.


Go Version
========================================================================

This repository is built with Go 1.12.


Build Instructions
========================================================================

```bash
    cd ~/wherever/you/keep/your/stuff
    mkdir -p tdop/src/perlmonger42
    git clone git@github.com:perlmonger42/tdop.git tdop/src/perlmonger42/tdop
    cd ./tdop
    export GOPATH="$( cd "$(pwd)"; pwd )"
    export PATH="$GOPATH/bin":$PATH
    cd ./src/perlmonger42/tdop
    go get golang.org/x/tools/cmd/stringer # install prerequisite
    go generate ./... && go test ./... && go install . && $GOPATH/bin/tdop
```
