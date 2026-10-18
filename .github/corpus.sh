#!/bin/sh
# corpus.sh: the tableaux corpus that tabloio's tests read in CI.
#
#   sh .github/corpus.sh fetch <directory>
#
# fetch  clones the repository corpus.lock.yaml names into <directory>, checks
#        out the locked commit and its submodules, points the clone's own tag
#        corpus/tooling at the locked tooling commit, and builds the corpus
#        under <directory>/corpus/build. It fails when the clone lacks either
#        commit. CI runs it; it needs the network, sh and git.
#
# The tests then read TABLEAUX_CORPUS=<directory>/corpus/build and
# TABLO_CORPUS=<directory>/corpus.

set -eu
HERE=$(cd "$(dirname "$0")" && pwd)
LOCK="$HERE/corpus.lock.yaml"

# field <key>: the value of a top-level scalar of the lock.
field() { sed -n "s/^$1: *\([^ #]*\).*/\1/p" "$LOCK"; }

# has <directory> <commit>: whether the clone holds the commit.
has() { git -C "$1" rev-parse -q --verify "$2^{commit}" > /dev/null; }

fetch() {
  dir=$1
  repository=$(field repository)
  commit=$(field commit)
  tooling=$(field tooling)
  git clone -q --no-checkout "$repository" "$dir"
  for c in "$commit" "$tooling"; do
    if ! has "$dir" "$c"; then
      echo "corpus.sh: $repository holds no commit $c" >&2
      exit 1
    fi
  done
  git -C "$dir" checkout -q --detach "$commit"
  git -C "$dir" submodule -q update --init
  git -C "$dir" tag -f corpus/tooling "$tooling" > /dev/null
  sh "$dir/corpus/build.sh" > /dev/null
  echo "corpus.sh: $dir holds $repository at $commit, its corpus built"
}

case "${1:-}" in
  fetch) [ $# -eq 2 ] || { echo "usage: corpus.sh fetch <directory>" >&2; exit 2; }
         fetch "$2" ;;
  *)     echo "usage: corpus.sh fetch <directory>" >&2; exit 2 ;;
esac
