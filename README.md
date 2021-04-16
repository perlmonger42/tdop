tdop
========================================================================

This project is an exploration of Top Down Operator Precedence parsing.
It will be seeded with a version of [Douglas Crockford's demonstration](
http://crockford.com/javascript/tdop/index.html), translated into Go from the
original JavaScript. It won't be good Go style in the beginning, because it
will be a fairly direct transliteration of Crockford's JavaScript.

From there I hope to grow it into a real Go program that can be used as a base
for further research.

Note (added 2021-04-16): There's a branch named `simple-ast` that probably ought to be merged into `master`. I think it was driven by a desire for a better separation between tokens and AST nodes. But that work was done in December 2020, and I'm not wanting to spend the time now to reconstruct my thoughts at that time. Especially on a project I'm abandoning. But I did want to get the work committed, so I did push `simple-ast` to github. Bottom line: if ever I revisit this repo, consider merging `simple-ast` into `master`.

Note (added 2021-04-16): This project was superceded by `tdop-simple`, which was later renamed `go-tdop`.

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
