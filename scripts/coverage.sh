#!/bin/bash
# Credit for this file goes to @sdboyer

fail=0
COVERALLS_TOKEN=t47LG6BQsfLwb9WxB56hXUezvwpED6D11
echo "Using Travis.CI"
echo "Coverage token $COVERALLS_TOKEN"
echo "mode: set" > acc.out



# Standard go tooling behavior is to ignore dirs with leading underscors
for dir in $(find . -maxdepth 10 -not -path './.git*' -not -path '*/_*' -type d);
do
  if ls $dir/*.go &> /dev/null; then
    go test -coverprofile=profile.out $dir || fail=1
    if [ -f profile.out ]
    then
      cat profile.out | grep -v "mode: set" >> acc.out
      rm profile.out
    fi
  fi
done

# Failures have incomplete results, so don't send
if [ -n "$COVERALLS_TOKEN" ] && [ "$fail" -eq 0 ]
then
	echo "goveralls -v -coverprofile=acc.out -service travis.ci -repotoken $COVERALLS_TOKEN"
 	goveralls -v -coverprofile=acc.out -service travis.ci -repotoken $COVERALLS_TOKEN
fi

rm -f acc.out

exit $fail