#!/usr/bin/env python3
"""The reference generator of the drafts in this directory.

It writes the brief of one work item from a fixture: the envelope that
`tablo view work-queue --brief` is assumed to emit (design/e4c7.md, A1 to A7).
The Go package internal/brief reproduces its output byte for byte; this
script does not ship.

    python3 brief.py e4c7-design.json > e4c7-design.txt

Exit status: 0 with the brief on standard output; 1 with one line on
standard error when the data holds no work item or the work waits.
"""
import json
import sys

WIDTH = 76  # runes per line, the indent included


def wrap(text, indent):
    """Fill text greedily to WIDTH. A blank line in text ends a paragraph."""
    out = []
    for n, para in enumerate(text.strip().split("\n\n")):
        if n:
            out.append("")
        line = ""
        for word in para.split():
            if not line:
                line = " " * indent + word
            elif len(line) + 1 + len(word) <= WIDTH:
                line += " " + word
            else:
                out.append(line)
                line = " " * indent + word
        if line:
            out.append(line)
    return out


def short(commit):
    return commit[:7]


def source(field, task_id):
    """The parenthesis that names where a junction field resolves from."""
    s = field.get("source") or {}
    kind = s.get("kind", "")
    if kind == "task":
        if s["id"] == task_id:
            return "(this task)"
        return "(%s, %s)" % (s["title"], s["id"])
    if kind == "assignee":
        return "(the assignee; the junction states none)"
    if kind == "default":
        return "(the assignee, by default)"
    return ""


def reference(r):
    return r["url"] + (", " + r["text"] if r.get("text") else "")


def fold(entries):
    """Group consecutive entries alike but for the task; a run of five folds."""
    runs = []
    for e in entries:
        key = (e["from"], e["to"], e.get("text", ""), e.get("condition", ""))
        if runs and runs[-1][0] == key:
            runs[-1][1].append(e)
        else:
            runs.append((key, [e]))
    out = []
    for _, run in runs:
        if len(run) >= 5:
            out.append(run)
        else:
            out.extend([e] for e in run)
    return out


def who(run):
    if len(run) == 1:
        return "%s %s" % (run[0]["id"], run[0]["title"])
    return "%d tasks, %s to %s," % (len(run), run[0]["id"], run[-1]["id"])


def unblocks(run):
    return "%s at %s, %s" % (who(run), run[0]["to"], run[0]["text"])


def also(run):
    e = run[0]
    return "%s from %s to %s, %s: %s" % (who(run), e["from"], e["to"], e["text"], e["condition"])


def requires(e):
    where = ""
    if e.get("url"):
        where = " in %s at %s" % (e["url"], short(e["commit"]))
    return "Requires: %s %s%s from %s to %s, %s: %s" % (
        e["id"], e["title"], where, e["from"], e["to"], e["text"], e["condition"])


def way(b):
    """How the gate passes: defined, handoff, self or none."""
    j = b["junction"]
    contributor, reviewer = j["contributor"]["value"], j["reviewer"]["value"]
    if b["gate"]["key"] == "defined":
        return "defined"
    if reviewer and reviewer != contributor:
        return "handoff"
    if reviewer:
        return "self"
    return "none"


def render(b):
    task, gate, j, st = b["task"], b["gate"], b["junction"], b["status"]
    tid, gkey = task["id"], gate["key"]
    contributor = j["contributor"]["value"]
    model = j["model"]["value"]
    reviewer = j["reviewer"]["value"]
    ref = b["ref"]
    named = ref["name"] != "HEAD" and not ref["commit"].startswith(ref["name"])
    at = ("on %s at %s" % (ref["name"], short(ref["commit"]))) if named else "at " + short(ref["commit"])
    base = ref["name"] if named else short(ref["commit"])
    L = []

    L.append("Brief: %s %s at %s" % (tid, task["title"], gkey))
    L.append("")
    L += wrap("This brief stands alone. It is one item of the work queue of %s %s, %s: %s."
              % (b["person"], at, ref["date"], b["kind"]), 0)
    L.append("")

    # 1. The item
    L.append("1. The item")
    L.append("")
    L.append("  Task         %s %s" % (tid, task["title"]))
    L.append("  Gate         %s %s, %s: %s" % (gate["symbol"], gkey, gate["name"], gate["criteria"]))
    rows = [("Contributor", contributor, source(j["contributor"], tid))]
    if model:
        rows.append(("Model", model, source(j["model"], tid)))
    if reviewer:
        rows.append(("Reviewer", reviewer, source(j["reviewer"], tid)))
    else:
        rows.append(("Reviewer", "none", ""))
    pad = max(len(v) for _, v, _ in rows)
    for label, value, src in rows:
        L.append(("  %-13s%-*s   %s" % (label, pad, value, src)).rstrip())
    if model:
        L.append("")
        L += wrap("The model is the plan's statement. Run on a model that matches it, as an "
                  "identifier or by prefix, and record the model you in fact run in the "
                  "Model: trailer.", 2)
    L.append("")

    # 2. The task
    L.append("2. The task")
    L.append("")
    L += wrap(task["description"], 2)
    L.append("")
    if task.get("references"):
        for r in task["references"]:
            L.append("  Reference: " + reference(r))
    else:
        L.append("  References: none")
    if j.get("references"):
        for r in j["references"]:
            L.append(("  Junction reference: %s   %s" % (reference(r), source(r, tid))).rstrip())
        L += wrap("A junction reference expands the gate's criteria for this task. "
                  "Read each one before you start.", 2)
    else:
        L.append("  Junction references: none")
    if task.get("path"):
        L += wrap("Under: " + " › ".join(p["title"] for p in task["path"]), 2)
    sup = b.get("supplier")
    if sup:
        L.append("")
        L += wrap("%s (%s), whose entry sends this work to you: %s"
                  % (sup["title"], sup["id"], " ".join(sup["description"].split())), 2)
    L.append("")

    # 3. The requirements
    L.append("3. The requirements")
    L.append("")
    if b.get("requires"):
        for e in b["requires"]:
            L += wrap(requires(e), 2)
    else:
        L.append("  Requires: nothing")
    deps = b.get("dependents") or []
    now = [e for e in deps if e["from"] == gkey]
    rest = [e for e in deps if e["from"] != gkey]
    if not deps:
        L += wrap("Unblocks: nothing; no task requires %s" % tid, 2)
    elif not now:
        L.append("  Unblocks: nothing at this gate")
    for run in fold(now):
        L += wrap("Unblocks: " + unblocks(run), 2)
    for run in fold(rest):
        L += wrap("Also required by: " + also(run), 2)
    parent = b.get("parent")
    if parent:
        pnow = [e for e in parent.get("dependents") or [] if e["from"] == gkey]
        if pnow:
            L += wrap("Its parent %s %s at %s is required by:" % (parent["id"], parent["title"], gkey), 2)
            for run in fold(pnow):
                L += wrap(unblocks(run), 4)
    L.append("")

    # 4. The status
    L.append("4. The status")
    L.append("")
    head = "%s %s %s %s" % (st["gate_symbol"], st["gate"], st["state_symbol"], st["state"])
    if st.get("reason"):
        head += " %s %s" % (st["reason_symbol"], st["reason"])
    head += ", %s, %s, %s" % (st["date"], st["recorder"], short(st["commit"]))
    if st.get("note"):
        L += wrap(head + ":", 2)
        L += wrap(st["note"], 2)
    else:
        L += wrap(head, 2)
    L.append("")

    # 5. The commit
    L.append("5. The commit you make when done")
    L.append("")
    w = way(b)
    lead = "Commit your work and %s, as %s, on a branch off %s. Do not merge. " % (st["path"], contributor, base)
    if w == "handoff":
        lead += ("The reviewer %s accepts this gate, so hand the work off: the status stays at "
                 "the gate it stands at and takes the reason review." % reviewer)
    elif w == "self":
        lead += ("You are the reviewer of this gate as well as its contributor, so the commit "
                 "that records the status passes the gate.")
    elif w == "defined":
        lead += ("Authorisation stands as the review of defined, so the commit that records the "
                 "status passes the gate and carries no Reviewed: trailer.")
    else:
        lead += "No one reviews this gate, so the commit that records the status passes it."
    L += wrap(lead, 2)
    L.append("")
    if w == "handoff":
        L.append("    gate: %s" % st["gate"])
        L.append("    state: nominal")
        L.append("    reason: review")
        L.append("    note: <one line that says what waits on the branch>")
    elif gate["follows"] == "recursive":
        L.append("    gate: %s" % gkey)
    else:
        L.append("    gate: %s" % gkey)
        L.append("    state: %s" % ("nominal" if gate["follows"] else "complete"))
        L.append("    note: <one line that says what the gate now holds>")
    L.append("")
    if w != "handoff" and gate["follows"] == "recursive":
        L += wrap("The next junction is recursive, so the file holds the gate alone. Edit "
                  "it by hand, and change no other file under .tableaux.", 2)
    else:
        L += wrap("Edit the file by hand so that it reads as above, with the note on one "
                  "line and no colon followed by a space inside it. Change no other file "
                  "under .tableaux.", 2)
    L.append("")
    if model:
        L += wrap("Author and committer are both %s. Set the identity in the environment, "
                  "which no configuration overrides, and check it on each commit:" % contributor, 2)
        L.append("")
        L.append("    export GIT_AUTHOR_EMAIL=%s" % contributor)
        L.append("    export GIT_COMMITTER_EMAIL=%s" % contributor)
        L.append("    export GIT_AUTHOR_NAME=\"<your model's display name>\"")
        L.append("    export GIT_COMMITTER_NAME=\"<your model's display name>\"")
        L.append("    git log -1 --format='%an <%ae> / %cn <%ce>%n%B'")
        L.append("")
        tail = ("End each commit message with one final paragraph of trailers, with no blank "
                "line between them and nothing after them.")
        if w == "self":
            tail += (" The commit that changes the status file carries all three; an earlier "
                     "commit carries the last two.")
        L += wrap(tail, 2)
        L.append("")
        if w == "self":
            L.append("    Reviewed: %s %s" % (tid, gkey))
        L.append("    Model: <the identifier your harness reports, matching %s>" % model)
        L.append("    Co-Authored-By: <your model's display name> <%s>" % contributor)
    elif w == "self":
        L += wrap("End the message of the commit that changes the status file with one "
                  "final paragraph of one trailer:", 2)
        L.append("")
        L.append("    Reviewed: %s %s" % (tid, gkey))
    else:
        L += wrap("A person contributes here, so the commit carries no trailer.", 2)
    L.append("")
    if w == "handoff":
        L += wrap("The reviewer merges the branch and passes the gate with a commit that "
                  "carries the trailer Reviewed: %s %s" % (tid, gkey), 2)
    else:
        L += wrap("Whoever dispatches you merges the branch.", 2)
    L.append("")

    # 6. The commands
    L.append("6. The commands that reproduce this brief and show the task")
    L.append("")
    L.append("  tabloio queue --person %s --ref %s --brief %s %s" % (b["person"], short(ref["commit"]), tid, gkey))
    L.append("  tabloio task %s --ref %s" % (tid, short(ref["commit"])))
    return "\n".join(L) + "\n"


def main():
    env = json.load(open(sys.argv[1], encoding="utf-8"))
    data = env["data"]
    b = data.get("brief")
    if not b:
        p = data["params"]
        sys.stderr.write("tabloio: no work for %s at %s %s; tabloio queue --person %s lists the items\n"
                         % (p["person"], p["brief"]["task"], p["brief"]["gate"], p["person"]))
        return 1
    if b["kind"] != "work ready":
        sys.stderr.write("tabloio: %s at %s waits and no brief starts it: %s\n"
                         % (b["task"]["id"], b["gate"]["key"], b["cause"]))
        return 1
    sys.stdout.write(render(b))
    return 0


if __name__ == "__main__":
    sys.exit(main())
