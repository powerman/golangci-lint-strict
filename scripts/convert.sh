#!/bin/bash
die() { echo "$@" >&2; exit 1;}

test $# -eq 2 || die "Usage: $0 config-version schema-version"
SOURCE="$1/reference/.golangci.yml"
OUTPUT="$1/.golangci.yml"
SCHEMA="jsonschema/golangci.v$2.jsonschema.json"
test -f "$SOURCE" || die "source file not exists: $SOURCE"
test -f "$SCHEMA" || die "schema file not exists: $SCHEMA"

# Strip non-EOL comments.
yq -P -oy "$SOURCE" | grep -E -v '^\s*#' > "$OUTPUT"
# Strip all invalid empty properties.
DEL_INVALID="$(jv "$SCHEMA" "$OUTPUT" | grep ': got null, want' | sed $'s,.*\'\\(.*\\)\'.*,del(\\1) |,; s,/,.,g')"
yq -i "$DEL_INVALID ." "$OUTPUT"
# Strip all properties with empty object value.
sed -i -e '/: {}$/d' -e '/^\s*$/d' "$OUTPUT"

sed -i -e "1i# Origin: https://github.com/powerman/golangci-lint-strict version $1" "$OUTPUT"

jv "$SCHEMA" "$OUTPUT"
