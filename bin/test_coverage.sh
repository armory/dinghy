#!/usr/bin/env bash

set +e

TMPFILE=coveragereport.txt
PCT=${PCT:-"50"}

# This is here because the build/test container we use doesn't have bc in it.
if [[ "$NEEDS_BC" == "yes" ]]; then
  apt update
  apt-get install bc
fi

# Run coverage
bin/coverage.sh > $TMPFILE
rc=$?

if [[ $rc == 0 ]]; then
  actual=`tail -1 $TMPFILE | awk '{print $3}' | sed 's/%//'`
  # NOTE:  Using bc because bash can't handle decimals intrinsically.
  rc=`echo "$actual < $PCT" | bc`
  if [[ $rc == 1 ]]; then
    echo "FAIL -- coverage too low (${actual}% < ${PCT}%)"
    exit 1
  else
    echo "PASS -- coverage ${actual}%"
  fi
  rm $TMPFILE
else
  echo "FAIL -- test failure, see ${TMPFILE}"
fi
exit $rc

