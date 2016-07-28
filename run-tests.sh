#!/bin/sh

rm -f test_vectors/observed/*.css
go run cmd/scss/*.go  --compile test_vectors/source:test_vectors/observed || exit $?


DIFF="$(which colordiff)"
if [ ! -x "$DIFF" ]
then
	DIFF="$(which diff)"
fi


if [ -x "$DIFF" ]
then
	echo
	$DIFF --exclude '.*.swp' -ur test_vectors/expected test_vectors/observed || exit $?
fi


