#!/bin/sh
# The shell harness of the Tableaux snapshot workflow (design 1422, T1 to
# T14). It builds throwaway repositories, a bare origin and clones, under a
# temporary directory, puts stub tablo and tabloio scripts first on the path,
# and points install at fixture archives through TABLEAUX_RELEASES. A stub
# logs its arguments; the stub tablo fails a --ref whose tree holds
# .tableaux/BROKEN. CI runs it beside go test:
#
#   snapshot/test.sh
#
# It needs git, curl, shasum, tar and jq; a case whose tool is missing skips
# with the reason. T14 runs the real releases from GitHub over the network,
# so it runs alone and only when asked:
#
#   snapshot/test.sh --real
set -eu

mode=${1:-}
case $mode in
'' | --real) ;;
*)
    echo "usage: snapshot/test.sh [--real]" >&2
    exit 2
    ;;
esac

here=$(cd "$(dirname "$0")" && pwd)
snapshot=$here/snapshot.sh
hook=$(dirname "$here")/hooks/pre-commit

work=$(mktemp -d "${TMPDIR:-/tmp}/snapshot-test.XXXXXX")
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/stub" "$work/home" "$work/tmp"

# No identity or setting of the host reaches Git.
for v in $(env | sed -n 's/^\(GIT_[A-Z_]*\)=.*/\1/p'); do unset "$v"; done
export HOME="$work/home" XDG_CONFIG_HOME="$work/home" GIT_CONFIG_NOSYSTEM=1 TMPDIR="$work/tmp"
export GIT_AUTHOR_NAME=Ada GIT_AUTHOR_EMAIL=ada@example.org
export GIT_COMMITTER_NAME=Ada GIT_COMMITTER_EMAIL=ada@example.org
export STUB_LOG="$work/stub.log"

# A case runs in a subshell, so a failure lands in a file the end counts.
: >"$work/failures"
pass() { echo "ok   $1"; }
fail() {
    echo "FAIL $1: $2"
    echo "$1" >>"$work/failures"
}
skip() { echo "skip $1: $2"; }
has() { command -v "$1" >/dev/null 2>&1; }
finish() {
    failures=$(wc -l <"$work/failures")
    [ "$failures" -eq 0 ] || {
        echo "$failures failed:" $(cat "$work/failures")
        exit 1
    }
    exit 0
}

# ---- the real releases: T14 ----

# With --real the harness installs tabloio from its GitHub release, at the
# newest tag that HEAD reaches unless TABLOIO_VERSION names another, and the
# tablo release that tabloio names, as the action picks them. It reruns T3, T5, T6
# and T11 with them on a copy of the corpus entry version-trunk-empty from
# TABLEAUX_CORPUS: its version.yaml gives an error with a position, P3, and
# a warning without one, P5, and a trunk set to main makes it valid. It then
# times the hook on a clone of the umbrella at TABLEAUX_DIR.
if [ "$mode" = --real ]; then
    unset TABLEAUX_RELEASES
    for tool in curl shasum tar jq; do
        has "$tool" || {
            skip T14 "$tool is missing"
            finish
        }
    done
    entry=${TABLEAUX_CORPUS:-}/version-trunk-empty
    [ -n "${TABLEAUX_CORPUS:-}" ] && [ -d "$entry/.tableaux" ] || {
        skip T14 "TABLEAUX_CORPUS names no built corpus with version-trunk-empty"
        finish
    }
    real=$work/real
    mkdir -p "$real"
    tabloio_version=${TABLOIO_VERSION:-$(git -C "$here" describe --tags --abbrev=0 2>/dev/null || true)}
    tablo_version=${TABLO_VERSION:-}
    if ! "$snapshot" install tabloio "$tabloio_version" "$real" ||
        ! tablo_version=${tablo_version:-$("$real/tabloio" version --json | jq -r .data.tablo)} ||
        ! "$snapshot" install tablo "$tablo_version" "$real"; then
        fail T14 "install tabloio $tabloio_version and tablo $tablo_version"
        finish
    fi
    echo "T14 installs $("$real/tabloio" version) and $("$real/tablo" version)"
    PATH="$real:$PATH"
    cp -a "$entry" "$work/entry"
    yaml=.tableaux/version.yaml

    (
        cd "$work"
        got=$("$snapshot" validate entry) && rc=0 || rc=$?
        if [ "$rc" -ne 1 ] || [ "$(echo "$got" | wc -l)" -ne 3 ] ||
            ! echo "$got" | grep -q "^::error file=entry/$yaml,line=2,col=8::P3 " ||
            ! echo "$got" | grep -q "^::warning file=entry/$yaml,line=1,col=1::P5 " ||
            [ "$(echo "$got" | tail -n 1)" != "::error::tablo validate exits 1; the snapshot waits for a valid project" ]; then
            fail T14/T3 "exit $rc, got
$got"
        else
            pass "T14/T3 validate turns tablo's error and warning into annotations"
        fi
    ) || fail T14/T3 "the case stops early"

    trunk=$(git -C "$work/entry" symbolic-ref --short HEAD)
    git clone -q --bare "$work/entry" "$work/real-origin.git"
    git clone -q "$work/real-origin.git" "$work/real-clone" 2>/dev/null
    (
        cd "$work/real-clone"
        source=$(git rev-parse HEAD)
        "$snapshot" render . STATUS.md tableau --level glance >/dev/null
        "$snapshot" commit STATUS.md "$trunk" >/dev/null
        who=$(git log -1 --format='%an <%ae>|%cn <%ce>')
        bot='github-actions[bot] <41898283+github-actions[bot]@users.noreply.github.com>'
        files=$(git diff-tree --no-commit-id --name-only -r HEAD)
        trailers=$(git log -1 --format=%B | git interpret-trailers --parse)
        if [ "$who" != "$bot|$bot" ]; then
            fail T14/T5 "the commit is by $who"
        elif [ "$files" != STATUS.md ] || [ "$(head -n 1 STATUS.md)" != "# Global tableau" ]; then
            fail T14/T5 "the commit changes $files to $(head -n 1 STATUS.md)"
        elif [ -n "$trailers" ] || [ "$(git log -1 --format=%s)" != "Regenerate STATUS.md" ]; then
            fail T14/T5 "subject $(git log -1 --format=%s), trailers $trailers"
        elif [ "$(git log -1 --format=%b)" != "$(tabloio version) renders it from $source." ]; then
            fail T14/T5 "body $(git log -1 --format=%b)"
        elif [ "$(git rev-parse HEAD)" != "$(git -C "$work/real-origin.git" rev-parse "$trunk")" ]; then
            fail T14/T5 "origin's $trunk is not the snapshot commit"
        else
            pass "T14/T5 commit commits tabloio's snapshot as the bot, touching the output alone, with no trailer"
        fi

        before=$(git rev-list --count HEAD)
        "$snapshot" render . STATUS.md tableau --level glance >/dev/null
        if out=$("$snapshot" commit STATUS.md "$trunk") && [ "$(git rev-list --count HEAD)" = "$before" ] && [ "$out" = "STATUS.md is current" ]; then
            pass "T14/T6 tabloio renders the same bytes after the snapshot commit, so commit makes none"
        else
            fail T14/T6 "commits $before then $(git rev-list --count HEAD): $out"
        fi
    ) || fail T14/T5 "the case stops early"

    good=$(sed 's/^trunk: ""$/trunk: main/' "$entry/$yaml")
    bad=$(
        cat "$entry/$yaml"
        echo "# a staged change"
    )
    t11=
    for staged in good bad; do
        for tree in same opposite; do
            rm -rf "$work/real-hooked"
            cp -a "$work/entry" "$work/real-hooked"
            cp "$hook" "$work/real-hooked/.git/hooks/pre-commit"
            chmod +x "$work/real-hooked/.git/hooks/pre-commit"
            (
                cd "$work/real-hooked"
                [ "$good" != "$(cat $yaml)" ]
                if [ $staged = good ]; then
                    printf '%s\n' "$good" >$yaml
                    git add $yaml
                    [ $tree = same ] || printf '%s\n' "$bad" >$yaml
                else
                    printf '%s\n' "$bad" >$yaml
                    git add $yaml
                    [ $tree = same ] || printf '%s\n' "$good" >$yaml
                fi
                if git commit -q -m "Edit the project" >"$work/out" 2>&1; then
                    [ $staged = good ]
                else
                    [ $staged = bad ] && grep -q ' P3 ' "$work/out"
                fi
            ) || t11="$t11 $staged-staged/$tree-tree"
        done
    done
    if [ -z "$t11" ]; then
        pass "T14/T11 the hook validates the staged tree with tablo, not the working tree"
    else
        fail T14/T11 "wrong verdict for$t11"
    fi

    # The hook's time: twenty runs on the Tableaux plan, one change staged
    # under .tableaux, against 200 ms at the median.
    ms() {
        n=$(date +%s%N)
        case $n in '' | *[!0-9]*) return 1 ;; esac
        echo $((n / 1000000))
    }
    if [ -z "${TABLEAUX_DIR:-}" ]; then
        skip T14/hook "TABLEAUX_DIR names no umbrella"
    elif ! ms >/dev/null; then
        skip T14/hook "date gives no nanoseconds"
    else
        u=$work/umbrella
        git clone --quiet --no-local "$TABLEAUX_DIR" "$u"
        git -C "$u" -c protocol.file.allow=always submodule update --init --quiet
        cp "$hook" "$u/.git/hooks/pre-commit"
        chmod +x "$u/.git/hooks/pre-commit"
        echo "# a staged change" >>"$u/$yaml"
        git -C "$u" add "$yaml"
        : >"$work/times"
        hookrc=0
        i=0
        while [ $i -lt 20 ]; do
            a=$(ms)
            (cd "$u" && .git/hooks/pre-commit) >"$work/out" 2>&1 || hookrc=$?
            b=$(ms)
            echo $((b - a)) >>"$work/times"
            i=$((i + 1))
        done
        sort -n "$work/times" >"$work/sorted"
        low=$(sed -n 1p "$work/sorted") high=$(sed -n 20p "$work/sorted")
        median=$((($(sed -n 10p "$work/sorted") + $(sed -n 11p "$work/sorted")) / 2))
        load=$(uptime | sed 's/.*load averages*: //')
        times="median $median ms, range $low to $high ms over 20 runs, load average $load"
        if [ "$hookrc" -ne 0 ]; then
            fail T14/hook "the hook exits $hookrc: $(cat "$work/out")"
        elif [ "$median" -ge 200 ]; then
            fail T14/hook "$times"
        else
            pass "T14/hook the hook validates the Tableaux plan in $times"
        fi
    fi
    finish
fi

PATH="$work/stub:$PATH"

# The stub tablo. validate --json prints $STUB_JSON and exits $STUB_EXIT;
# --ref <tree> validate fails when the tree holds .tableaux/BROKEN.
cat >"$work/stub/tablo" <<'EOF'
#!/bin/sh
echo "tablo $*" >>"$STUB_LOG"
dir=. ref=
while [ $# -gt 0 ]; do
    case $1 in
    -C) dir=$2; shift 2 ;;
    --ref) ref=$2; shift 2 ;;
    *) break ;;
    esac
done
[ "$1" = validate ] || { echo "stub tablo: unexpected $*" >&2; exit 2; }
if [ -n "$ref" ]; then
    if git -C "$dir" ls-tree -r --full-tree --name-only "$ref" | grep -q '\.tableaux/BROKEN$'; then
        echo "error S3 .tableaux/BROKEN: the stub's marker"
        exit 1
    fi
    exit 0
fi
cat "${STUB_JSON:-/dev/null}"
exit "${STUB_EXIT:-0}"
EOF
# The stub tabloio. version prints its version; a view prints $STUB_RENDER
# or exits $STUB_RENDER_EXIT.
cat >"$work/stub/tabloio" <<'EOF'
#!/bin/sh
echo "tabloio $*" >>"$STUB_LOG"
case $1 in
version) echo "tabloio v0.2.0"; exit 0 ;;
esac
[ "${STUB_RENDER_EXIT:-0}" -eq 0 ] || exit "$STUB_RENDER_EXIT"
cat "$STUB_RENDER"
EOF
chmod +x "$work/stub/tablo" "$work/stub/tabloio"

# ---- install: T1, T2 ----

host_os=$(uname -s | tr '[:upper:]' '[:lower:]')
case $(uname -m) in
x86_64 | amd64) host_arch=amd64 ;;
aarch64 | arm64) host_arch=arm64 ;;
*) host_arch=$(uname -m) ;;
esac

# releases <dir> <tool> <version>: fixture archives for each os and arch,
# each holding a binary that names its platform, and checksums.txt.
releases() {
    d=$1/$2/releases/download/v$3
    mkdir -p "$d"
    for os in linux darwin; do
        for arch in amd64 arm64; do
            b=$work/build/$2-$os-$arch
            mkdir -p "$b"
            printf '#!/bin/sh\necho "%s %s %s %s"\n' "$2" "$3" "$os" "$arch" >"$b/$2"
            chmod +x "$b/$2"
            tar -czf "$d/$2_$3_${os}_$arch.tar.gz" -C "$b" "$2"
        done
    done
    (cd "$d" && shasum -a 256 ./*.tar.gz | sed 's| \./| |' >checksums.txt)
}

if has curl && has shasum && has tar; then
    rel=$work/releases
    releases "$rel" tablo 0.1.0
    export TABLEAUX_RELEASES="file://$rel"

    bin=$work/bin1
    mkdir -p "$bin"
    if out=$("$snapshot" install tablo v0.1.0 "$bin" 2>&1) && [ "$("$bin/tablo")" = "tablo 0.1.0 $host_os $host_arch" ]; then
        pass "T1 install fetches the host's archive into bindir"
    else
        fail T1 "install: $out; binary: $("$bin/tablo" 2>&1 || true)"
    fi

    asset=$rel/tablo/releases/download/v0.1.0/tablo_0.1.0_${host_os}_$host_arch.tar.gz
    cp "$asset" "$work/asset.keep"
    printf 'corrupt' >>"$asset"
    bin=$work/bin2
    mkdir -p "$bin"
    if "$snapshot" install tablo v0.1.0 "$bin" >"$work/out" 2>&1 || [ -e "$bin/tablo" ]; then
        fail T2 "a checksum mismatch installs: $(cat "$work/out")"
    elif ! grep -q "fails its checksum" "$work/out"; then
        fail T2 "a checksum mismatch says $(cat "$work/out")"
    else
        rm "$asset"
        if "$snapshot" install tablo v0.1.0 "$bin" >"$work/out" 2>&1 || [ -e "$bin/tablo" ]; then
            fail T2 "a missing asset installs"
        elif "$snapshot" install tablo main "$bin" >"$work/out" 2>&1 || [ -e "$bin/tablo" ] || ! grep -q "names no release" "$work/out"; then
            fail T2 "a ref that is no tag: $(cat "$work/out")"
        else
            pass "T2 install refuses a checksum mismatch, a missing asset and a ref that is no tag"
        fi
    fi
    cp "$work/asset.keep" "$asset"
else
    skip T1 "curl, shasum or tar is missing"
    skip T2 "curl, shasum or tar is missing"
fi

# ---- validate: T3, T4 ----

if has jq; then
    cat >"$work/diags.json" <<'EOF'
{"schema": "tablo/1", "command": "validate", "data": {"valid": false},
 "diagnostics": [
  {"severity": "error", "code": "S5", "path": ".tableaux/status/9f31.yaml", "line": 2, "col": 7, "message": "gate names no gate"},
  {"severity": "warning", "code": "P5", "path": ".tableaux/version.yaml", "message": "the trunk is undetermined"}]}
EOF
    want='::error file=proj/.tableaux/status/9f31.yaml,line=2,col=7::S5 gate names no gate
::warning file=proj/.tableaux/version.yaml,line=1,col=1::P5 the trunk is undetermined
::error::tablo validate exits 1; the snapshot waits for a valid project'
    got=$(STUB_JSON=$work/diags.json STUB_EXIT=1 "$snapshot" validate proj) && rc=0 || rc=$?
    if [ "$got" = "$want" ]; then
        pass "T3 validate turns diagnostics into annotations at file, line and column"
    else
        fail T3 "got
$got"
    fi
    cat >"$work/warns.json" <<'EOF'
{"diagnostics": [{"severity": "warning", "code": "H3", "path": ".tableaux/tasks/9f31.yaml", "line": 4, "col": 3, "message": "a model outside"}]}
EOF
    if [ "$rc" -ne 1 ]; then
        fail T4 "an error exits $rc"
    elif ! STUB_JSON=$work/warns.json STUB_EXIT=0 "$snapshot" validate proj >"$work/out"; then
        fail T4 "warnings alone fail: $(cat "$work/out")"
    elif ! grep -q '^::warning file=proj/.tableaux/tasks/9f31.yaml,line=4,col=3::H3 a model outside$' "$work/out"; then
        fail T4 "warnings alone print $(cat "$work/out")"
    else
        pass "T4 validate fails on an error and passes on warnings alone"
    fi
else
    skip T3 "jq is missing"
    skip T4 "jq is missing"
fi

# ---- render and commit: T5 to T9 ----

# origin: a bare origin with one commit on main and a clone of it in $1.
origin() {
    rm -rf "$work/origin.git" "$1"
    git init -q --bare -b main "$work/origin.git"
    git clone -q "$work/origin.git" "$1" 2>/dev/null
    (
        cd "$1"
        mkdir .tableaux
        echo "gates: []" >.tableaux/gates.yaml
        echo "old snapshot" >STATUS.md
        git add .
        git commit -q -m "Start"
        git push -q origin HEAD:refs/heads/main
    )
}
count() { git -C "$1" rev-list --count HEAD; }

clone=$work/clone
origin "$clone"
echo "new snapshot" >"$work/render"
export STUB_RENDER="$work/render"
(
    cd "$clone"
    source=$(git rev-parse HEAD)
    "$snapshot" render . STATUS.md tableau --level glance >/dev/null
    "$snapshot" commit STATUS.md main >/dev/null
    who=$(git log -1 --format='%an <%ae>|%cn <%ce>')
    bot='github-actions[bot] <41898283+github-actions[bot]@users.noreply.github.com>'
    files=$(git diff-tree --no-commit-id --name-only -r HEAD)
    trailers=$(git log -1 --format=%B | git interpret-trailers --parse)
    body=$(git log -1 --format=%b)
    if [ "$who" != "$bot|$bot" ]; then
        fail T5 "the commit is by $who"
    elif [ "$files" != STATUS.md ] || [ "$(cat STATUS.md)" != "new snapshot" ]; then
        fail T5 "the commit changes $files"
    elif [ -n "$trailers" ] || [ "$(git log -1 --format=%s)" != "Regenerate STATUS.md" ]; then
        fail T5 "subject $(git log -1 --format=%s), trailers $trailers"
    elif [ "$body" != "tabloio v0.2.0 renders it from $source." ]; then
        fail T5 "body $body"
    elif [ "$(git rev-parse HEAD)" != "$(git -C "$work/origin.git" rev-parse main)" ]; then
        fail T5 "origin's main is not the snapshot commit"
    elif ! grep -q "^tabloio -C . tableau --level glance --markdown$" "$STUB_LOG"; then
        fail T5 "render ran $(grep tabloio "$STUB_LOG")"
    else
        pass "T5 commit commits a changed snapshot as the bot, touching the output alone, with no trailer"
    fi

    before=$(count .)
    "$snapshot" render . STATUS.md tableau --level glance >/dev/null
    if out=$("$snapshot" commit STATUS.md main) && [ "$(count .)" = "$before" ] && [ "$out" = "STATUS.md is current" ]; then
        pass "T6 commit makes no commit when the snapshot is unchanged"
    else
        fail T6 "commits $before then $(count .): $out"
    fi
) || fail T5 "the case stops early"

origin "$clone"
git clone -q "$work/origin.git" "$work/other" 2>/dev/null
(
    cd "$work/other"
    echo "moved" >README.md
    git add README.md
    git commit -q -m "Move the trunk"
    git push -q origin HEAD:refs/heads/main
) || fail T7 "the case stops early"
(
    cd "$clone"
    "$snapshot" render . STATUS.md tableau --level glance >/dev/null
    if out=$("$snapshot" commit STATUS.md main 2>&1) &&
        [ "$(git -C "$work/origin.git" rev-parse main)" = "$(git -C "$work/other" rev-parse HEAD)" ] &&
        echo "$out" | grep -q '^::notice::main moved past'; then
        pass "T7 commit exits 0 without a commit on the trunk when the trunk moved"
    else
        fail T7 "$out"
    fi
) || fail T7 "the case stops early"
rm -rf "$work/other"

origin "$clone"
printf '#!/bin/sh\necho "the trunk is protected" >&2\nexit 1\n' >"$work/origin.git/hooks/pre-receive"
chmod +x "$work/origin.git/hooks/pre-receive"
(
    cd "$clone"
    "$snapshot" render . STATUS.md tableau --level glance >/dev/null
    if out=$("$snapshot" commit STATUS.md main 2>&1); then
        fail T8 "a rejected push passes: $out"
    elif ! echo "$out" | grep -q '^::error::the push of STATUS.md to main fails'; then
        fail T8 "$out"
    else
        pass "T8 commit fails when the push fails for another reason"
    fi
) || fail T8 "the case stops early"

origin "$clone"
(
    cd "$clone"
    if STUB_RENDER_EXIT=1 "$snapshot" render . STATUS.md tableau --level glance >/dev/null 2>&1; then
        fail T9 "a failed render passes"
    elif [ "$(cat STATUS.md)" != "old snapshot" ]; then
        fail T9 "a failed render writes $(cat STATUS.md)"
    else
        pass "T9 a render failure leaves the output untouched and fails the step"
    fi
) || fail T9 "the case stops early"

# ---- the hook: T10 to T13 ----

# hooked <dir> <project>: a repository with the hook installed and a project
# at <project>, "." for the root.
hooked() {
    rm -rf "$1"
    git init -q -b main "$1"
    (
        cd "$1"
        mkdir -p "$2/.tableaux"
        echo "gates: []" >"$2/.tableaux/gates.yaml"
        echo "# Project" >README.md
        git add .
        git commit -q -m "Start"
        cp "$hook" .git/hooks/pre-commit
        chmod +x .git/hooks/pre-commit
    )
}
calls() { grep -c '^tablo ' "$STUB_LOG" || true; }

repo=$work/hooked
hooked "$repo" .
(
    cd "$repo"
    : >"$STUB_LOG"
    echo "more" >>README.md
    git add README.md
    if git commit -q -m "Edit the README" && [ "$(calls)" = 0 ]; then
        pass "T10 the hook skips validation when nothing under .tableaux is staged"
    else
        fail T10 "$(cat "$STUB_LOG")"
    fi
) || fail T10 "the case stops early"

t11=
for staged in good bad; do
    for tree in same opposite; do
        hooked "$repo" .
        (
            cd "$repo"
            echo "one" >.tableaux/notes.yaml
            git add .tableaux/notes.yaml
            if [ $staged = bad ]; then
                touch .tableaux/BROKEN
                git add .tableaux/BROKEN
                [ $tree = same ] || rm .tableaux/BROKEN
            elif [ $tree = opposite ]; then
                touch .tableaux/BROKEN
            fi
            if git commit -q -m "Edit the project" >/dev/null 2>&1; then
                [ $staged = good ]
            else
                [ $staged = bad ]
            fi
        ) || t11="$t11 $staged-staged/$tree-tree"
    done
done
if [ -z "$t11" ]; then
    pass "T11 the hook validates the staged tree, not the working tree"
else
    fail T11 "wrong verdict for$t11"
fi

hooked "$repo" .
mkdir -p "$work/notablo"
ln -sf "$(command -v git)" "$work/notablo/git"
(
    cd "$repo"
    echo "one" >.tableaux/notes.yaml
    git add .tableaux/notes.yaml
    if out=$(PATH="$work/notablo" .git/hooks/pre-commit 2>&1); then
        fail T12 "the hook passes without tablo"
    elif ! echo "$out" | grep -q -- "--no-verify"; then
        fail T12 "$out"
    else
        pass "T12 the hook fails, naming the bypass, when tablo is absent"
    fi
) || fail T12 "the case stops early"

hooked "$repo" sub/dir
(
    cd "$repo"
    git config tableaux.project sub/dir
    : >"$STUB_LOG"
    echo "one" >.tableaux-not-it
    git add .tableaux-not-it
    git commit -q -m "Stage nothing of the project"
    skipped=$(calls)
    echo "one" >sub/dir/.tableaux/notes.yaml
    git add sub/dir/.tableaux/notes.yaml
    git commit -q -m "Edit the project"
    top=$(git rev-parse --show-toplevel)
    if [ "$skipped" = 0 ] && grep -q "^tablo -C $top/sub/dir --ref [0-9a-f]* validate$" "$STUB_LOG"; then
        pass "T13 the hook honours tableaux.project"
    else
        fail T13 "$(cat "$STUB_LOG")"
    fi
) || fail T13 "the case stops early"

echo "T14 runs the real releases with snapshot/test.sh --real"
finish
