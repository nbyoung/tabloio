"""Reference for design b618: the scalar rule, the status line edit and the task
template. It writes scalars.tsv, status-edits.txt and task.yaml beside itself
and checks each through PyYAML, with 200,000 random scalars besides. The Go
implementation follows the same rules and tests against the three files; this
script does not ship.

    python3 design/b618/reference.py
"""
import os, re, yaml, random, textwrap
OUT = os.path.dirname(os.path.abspath(__file__))
RES = {"true","false","yes","no","on","off","null","y","n"}
def scalar(v, flow=False):
    plain = (re.match(r'^[A-Za-z]', v) and v.lower() not in RES
             and ': ' not in v and ' #' not in v and not v.endswith(':') and not v.endswith(' '))
    if flow and not re.fullmatch(r"[A-Za-z0-9 ._/~()'+=@%:#-]+", v): plain = False
    if plain: return v
    return '"' + v.replace('\\','\\\\').replace('"','\\"') + '"'
ORDER = ["gate","state","reason","note"]
def extent(lines, key):
    for i,l in enumerate(lines):
        if re.match(re.escape(key)+r':( |\r?\n|$)', l):
            e = i+1
            while e < len(lines):
                if lines[e][:1] in (" ","\t"): e += 1; continue
                if lines[e].strip()=="":
                    j=e
                    while j < len(lines) and lines[j].strip()=="": j+=1
                    if j < len(lines) and lines[j][:1] in (" ","\t"): e=j; continue
                break
            return i,e
    return None
def edit(text, ops):
    lines = text.splitlines(keepends=True)
    eol = "\r\n" if lines and lines[0].endswith("\r\n") else "\n"
    for l in lines:
        if l.strip()=="" or l.startswith("#") or l[:1] in (" ","\t"): continue
        if not any(re.match(k+r':( |\r?\n|$)', l) for k in ORDER): raise ValueError("status-form: "+l)
    for k in ORDER:
        if sum(1 for l in lines if re.match(k+r':( |\r?\n|$)', l))>1: raise ValueError("status-form dup")
    for k in ORDER:
        if k not in ops: continue
        v = ops[k]; ex = extent(lines,k)
        if v is None:
            if ex: del lines[ex[0]:ex[1]]
            continue
        new = k+": "+scalar(v)+eol
        if ex: lines[ex[0]:ex[1]] = [new]; continue
        pos=None
        for p in reversed(ORDER[:ORDER.index(k)]):
            e=extent(lines,p)
            if e: pos=e[1]; break
        if pos is None:
            for p in ORDER[ORDER.index(k)+1:]:
                e=extent(lines,p)
                if e: pos=e[0]; break
        if pos is None: pos=len(lines)
        if pos>0 and not lines[pos-1].endswith("\n"): lines[pos-1]+=eol
        lines.insert(pos,new)
    return "".join(lines)
cases = [
 ("advance keeps comments and spacing", "# kept verbatim\ngate: defined\n\nstate:   nominal\nnote: The prototype is next\n", {"gate":"function","note":"Write commands shown"}, {"gate":"function","state":"nominal","note":"Write commands shown"}),
 ("hand-off inserts reason in canonical order", "gate: function\nstate: nominal\nnote: old\n", {"reason":"review","note":"design/b618.md awaits the owner's review"}, {"gate":"function","state":"nominal","reason":"review","note":"design/b618.md awaits the owner's review"}),
 ("note with a colon and a space is quoted", "gate: function\nstate: nominal\n", {"note":"prototype/171b/ shows live refresh: git-polling watcher for HEAD"}, {"gate":"function","state":"nominal","note":"prototype/171b/ shows live refresh: git-polling watcher for HEAD"}),
 ("folded note is replaced whole", "gate: defined\nstate: nominal\nnote: >\n  The prototype is\n\n  next\n", {"note":"Done"}, {"gate":"defined","state":"nominal","note":"Done"}),
 ("recursive next junction leaves the gate alone", "gate: function\nstate: stalled\nreason: blocked\nnote: Barometer ICs on 14-week backorder\n", {"gate":"design","state":None,"reason":None,"note":None}, {"gate":"design"}),
 ("no file yet", "", {"gate":"defined","state":"nominal"}, {"gate":"defined","state":"nominal"}),
 ("gate-only file gains a state", "gate: implementation\n", {"gate":"unit","state":"nominal","note":"Tests pass"}, {"gate":"unit","state":"nominal","note":"Tests pass"}),
 ("state before a late gate line", "# odd order\nstate: nominal\ngate: defined\n", {"reason":"blocked"}, {"gate":"defined","state":"nominal","reason":"blocked"}),
 ("a key that YAML reads as a boolean is quoted", "gate: defined\nstate: nominal\n", {"state":"on"}, {"gate":"defined","state":"on"}),
 ("last gate completes and drops the reason", "gate: validate\nstate: nominal\nreason: review\nnote: Release notes await review\n", {"gate":"release","state":"complete","reason":None,"note":None}, {"gate":"release","state":"complete"}),
]
out=[]
for name,before,ops,want in cases:
    after = edit(before,ops)
    assert yaml.safe_load(after)==want,(name,after)
    # untouched lines stay
    out.append("=== "+name+"\n--- before\n"+before+"--- set\n"+"".join((k+"="+ops[k]+"\n") if ops[k] is not None else (k+"-\n") for k in ORDER if k in ops)+"--- after\n"+after)
open(os.path.join(OUT,"status-edits.txt"),"w").write("".join(out))
# refusal
for bad in ["{ gate: design }\n","gate: a\ngate: b\n","---\ngate: a\n","gate: a\nextra: 1\n"]:
    try: edit(bad,{"gate":"x"}); print("NOT REFUSED",bad)
    except ValueError: pass
# scalars
vals = ["Barometer ICs on 14-week backorder","prototype/171b/ shows live refresh: git-polling watcher","design/b618.md on branch wave/b618 awaits the owner's review at the design gate","40e8","07e0","1000","1e10","2026-10-05","true","No","on","y","~","waits for #7","waits #7","ends with a colon:","key:value","say \"hi\" to C:\\temp","a 'quote' here","- dash first","[bracket first","a, b [c] {d}","👓 symbol first","note 👓 inside","&anchor","*alias","!tag","%percent","@at","`tick","> fold","| literal","? question","é accent first"]
rows=[]
for v in vals:
    b=scalar(v); f=scalar(v,True)
    assert yaml.safe_load("note: "+b+"\n")=={"note":v}
    assert yaml.safe_load("- { url: "+f+", text: "+f+" }\n")==[{"url":v,"text":v}],v
    rows.append(v+"\t"+b+"\t"+f+"\n")
open(os.path.join(OUT,"scalars.tsv"),"w").write("value\tblock\tflow\n"+"".join(rows))
random.seed(2); alpha=list("ab:# \"'\\-[]{},&*!|>%@`?~01e./()+=")
n=0
for _ in range(200000):
    v="".join(random.choice(alpha) for _ in range(random.randint(1,8)))
    if v!=v.strip(): continue
    n+=1
    assert yaml.safe_load("- { url: "+scalar(v,True)+", text: "+scalar(v,True)+" }\n")==[{"url":v,"text":v}],v
    assert yaml.safe_load("title: "+scalar(v)+"\n")=={"title":v},v
print("flow+block ok",n)
# task file
def task(title,desc,assignee,refs,parent,order):
    s="title: "+scalar(title)+"\ndescription: >\n"
    for l in textwrap.wrap(" ".join(desc.split()),78,break_long_words=False,break_on_hyphens=False): s+="  "+l+"\n"
    s+="assignee: "+scalar(assignee)+"\n"
    if refs:
        s+="references:\n"
        for u,t in refs:
            s+="  - { url: "+scalar(u,True)+(", text: "+scalar(t,True) if t else "")+" }\n"
    s+='parent: { id: "%s", order: %d }\n'%(parent,order)
    return s
t=task("Rain gauge: tipping bucket","A tipping-bucket gauge on the mast that counts 0.2 mm tips, debounces them in firmware and reports the hourly total to the gateway.","ada@example.org",[("docs/rain-gauge.md",None),("https://example.org/datasheets/rg11.pdf","Gauge datasheet, revision 2")],"07e0",3)
d=yaml.safe_load(t); assert d["parent"]=={"id":"07e0","order":3} and d["title"]=="Rain gauge: tipping bucket", d
open(os.path.join(OUT,"task.yaml"),"w").write(t)
