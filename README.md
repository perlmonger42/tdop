tdop
========================================================================

This project is an exploration of Top Down Operator Precedence parsing.
It will be seeded with a version of [Douglas Crockford's demonstration](
http://crockford.com/javascript/tdop/index.html), translated into Go from the
original JavaScript. It won't be good Go style in the beginning, because it
will be a fairly direct transliteration of Crockford's JavaScript.

From there I hope to grow it into a real Go program that can be used as a base
for further research.

Go Modules
========================================================================
This repository is built with Go 1.12,
and opts-in to the [modules-based behavior introduced in Go 1.11](
https://github.com/golang/go/wiki/Modules#example).

Build Instructions
========================================================================

Change directory to the root of the `tdop` project and execute:

```bash
    go generate && go test ./... && go build
```
