go-scss is an experiment in writing a compiler in Go. It compiles SASS/SCSS into CSS.

Building
--------
go-scss has few external dependencies, partly by design, partly because *this is the manner in which I roll*. To quickly grab the libraries you do need, run:
```
go get -u github.com/thijzert/go-termcolours
```
After that, simply run:
```
go build -o scss cmd/scss/*.go
```

Usage
-----
To compile a single file, run:
```
./scss FILE.scss
```
This will create FILE.css

Known issues
------------
go-scss currently only supports an embarrassingly tiny subset of SASS/SCSS.

License
-------
You may distribute this program and its source code under the terms of the BSD 3-clause license. You can find the details at https://www.tldrlegal.com/l/bsd3

Acknowledgements
----------------
go-scss includes a lexer based on Rob Pike's talk *[Lexical scanning in Go](https://www.youtube.com/watch?v=HxaD_trXwRE)* (a recommended watch), helpfully written down by @bbuck at https://github.com/bbuck/go-lexer. It is licensed under the terms of the MIT license.
