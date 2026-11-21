#!/bin/sh
# Checks that each test is named after what it tests (CONTRIBUTING.md,
# "Nombres"), so the test of a file is found from its name and back.
set -eu
cd "$(dirname "$0")/.."
status=0

# Backend: <file>_test.go tests <file>.go, next to it. export_test.go only opens
# internals of the package to its tests.
for test in $(find services/backend/src -name '*_test.go' ! -name export_test.go | sort); do
  if [ ! -f "${test%_test.go}.go" ]; then
    echo "$test: there is no ${test%_test.go}.go to test"
    status=1
  fi
done

# Web: tests/<path>.test.ts(x) tests src/<path>.ts(x). tests/catalog/stories
# renders every story of the design catalog, which has no file of its own.
for test in $(find web/tests \( -name '*.test.ts' -o -name '*.test.tsx' \) ! -path web/tests/catalog/stories.test.tsx | sort); do
  source="web/src/${test#web/tests/}"
  source="${source%.test.*}"
  if [ ! -f "$source.ts" ] && [ ! -f "$source.tsx" ]; then
    echo "$test: there is no $source.ts or $source.tsx to test"
    status=1
  fi
done

exit $status
