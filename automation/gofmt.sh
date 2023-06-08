#! /bin/bash

set -e

echo "Checking go format"
sources="collector/pkg/ collector/cmd/"
unformatted=$(gofmt -e -d -s -l $sources)
if [ ! -z "$unformatted" ]; then
    # Some files are not gofmt.
    echo >&2 "The following Go files must be formatted with gofmt:"
    for fn in $unformatted; do
        echo >&2 "  $fn"
    done
fi

exit 0
