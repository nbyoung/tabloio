#!/bin/sh
# DRAFT: the design of tabloio task 1422, not the implementation.
# Lands as snapshot/snapshot.sh beside action.yml. Each subcommand is one step
# of the recipe, so the shell harness tests it without GitHub Actions.
#
#   snapshot.sh install <tool> <version> <bindir>
#   snapshot.sh validate <project>
#   snapshot.sh render <project> <output> <view args...>
#   snapshot.sh commit <output> <trunk>
set -eu

# The harness points this at a file:// fixture of release archives.
RELEASES=${TABLEAUX_RELEASES:-https://github.com/nbyoung}

cmd=$1; shift
case $cmd in
install)
    tool=$1 bin=$3
    case $2 in
    v[0-9]*) ver=${2#v} ;;
    *) echo "::error::$tool: '$2' names no release; set ${tool}-version"; exit 2 ;;
    esac
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case $(uname -m) in
    x86_64 | amd64) arch=amd64 ;;
    aarch64 | arm64) arch=arm64 ;;
    *) echo "::error::no $tool release for $(uname -m)"; exit 2 ;;
    esac
    asset=${tool}_${ver}_${os}_${arch}.tar.gz
    tmp=$(mktemp -d)
    curl -fsSL -o "$tmp/$asset" "$RELEASES/$tool/releases/download/v$ver/$asset"
    curl -fsSL -o "$tmp/checksums.txt" "$RELEASES/$tool/releases/download/v$ver/checksums.txt"
    (cd "$tmp" && grep " $asset\$" checksums.txt | shasum -a 256 -c -) >/dev/null ||
        { echo "::error::$asset fails its checksum"; exit 1; }
    tar -xzf "$tmp/$asset" -C "$bin" "$tool"
    ;;
validate)
    out=$(mktemp)
    rc=0
    tablo -C "$1" validate --json >"$out" || rc=$?
    jq -r --arg p "$1" '.diagnostics[] |
        "::\(if .severity == "error" then "error" else "warning" end) file=\($p)/\(.path),line=\(.line // 1),col=\(.col // 1)::\(.code) \(.message)"' "$out"
    [ "$rc" -eq 0 ] || echo "::error::tablo validate exits $rc; the snapshot waits for a valid project"
    exit "$rc"
    ;;
render)
    project=$1 output=$2; shift 2
    tabloio -C "$project" "$@" --markdown >"$output.new"
    mv "$output.new" "$output"
    ;;
commit)
    output=$1 trunk=$2
    source=$(git rev-parse HEAD)
    git add -- "$output"
    if git diff --cached --quiet -- "$output"; then
        echo "$output is current"
        exit 0
    fi
    export GIT_AUTHOR_NAME='github-actions[bot]' GIT_COMMITTER_NAME='github-actions[bot]'
    export GIT_AUTHOR_EMAIL='41898283+github-actions[bot]@users.noreply.github.com'
    export GIT_COMMITTER_EMAIL=$GIT_AUTHOR_EMAIL
    git commit -q -m "Regenerate $output" \
        -m "$(tabloio version) renders it from ${source}." -- "$output"
    if ! git push -q origin "HEAD:refs/heads/$trunk"; then
        git fetch -q origin "$trunk"
        if [ "$(git rev-parse FETCH_HEAD)" != "$source" ]; then
            echo "::notice::$trunk moved past $source; the run for its new head regenerates $output"
            exit 0
        fi
        echo "::error::the push of $output to $trunk fails"
        exit 1
    fi
    ;;
*)
    echo "usage: snapshot.sh install|validate|render|commit ..." >&2
    exit 2
    ;;
esac
