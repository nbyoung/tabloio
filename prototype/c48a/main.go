// Command c48a renders the authority delegation and task assignment views as
// Markdown from the JSON that tablo's structural-view prototype emits.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Levels nest by inclusion: detail includes glance, provenance includes both.
const (
	glance = iota
	detail
	provenance
)

type view struct {
	View    string `json:"view"`
	Ref     string `json:"ref"`
	OnTrunk bool   `json:"on_trunk"`
	Params  struct {
		Person   string `json:"person"`
		Task     string `json:"task"`
		Proposed bool   `json:"proposed"`
	} `json:"params"`
	Glance     levelData `json:"glance"`
	Detail     levelData `json:"detail"`
	Provenance levelData `json:"provenance"`
}

// levelData holds the keys the two views use at any level.
type levelData struct {
	Rows    []row    `json:"rows"`
	Columns []string `json:"columns"`
	People  []person `json:"people"`
}

type row struct {
	ID               string                       `json:"id"`
	Title            string                       `json:"title"`
	Assignee         string                       `json:"assignee"`
	Authorisation    string                       `json:"authorisation"`
	Children         int                          `json:"children"`
	Delegated        bool                         `json:"delegated"`
	Depth            int                          `json:"depth"`
	Differs          bool                         `json:"differs_from_trunk"`
	Authorities      []string                     `json:"authorities"`
	JunctionDefaults map[string]map[string]string `json:"junction_defaults"`
	By               string                       `json:"by"`
	Way              string                       `json:"way"`
	Commit           *commit                      `json:"commit"`
}

type commit struct {
	Hash      string `json:"hash"`
	Author    string `json:"author"`
	Committer string `json:"committer"`
	Date      string `json:"date"`
	Subject   string `json:"subject"`
}

type person struct {
	Email          string          `json:"email"`
	Assigned       json.RawMessage `json:"assigned"`
	ContributesNxt int             `json:"contributes_next"`
	ReviewsNxt     int             `json:"reviews_next"`
	Models         []string        `json:"models"`
	AssignedBy     map[string]int  `json:"assigned_by_gate"`
	Authority      []subtree       `json:"authority_over"`
	Contributes    []junction      `json:"contributes"`
	Reviews        []junction      `json:"reviews"`
}

type subtree struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Descendants int    `json:"descendants"`
}

type junction struct {
	Task        string            `json:"task"`
	Gate        string            `json:"gate"`
	Next        bool              `json:"next"`
	Contributor string            `json:"contributor"`
	Reviewer    string            `json:"reviewer"`
	Model       string            `json:"model"`
	Source      string            `json:"source"`
	Sources     map[string]string `json:"sources"`
}

type assigned struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status struct {
		Gate, State, Reason string
	} `json:"status"`
}

func main() {
	level := flag.String("level", "provenance", "glance, detail or provenance")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: c48a [-level glance|detail|provenance] [file.json]\nReads a view from the file or standard input and writes Markdown.")
	}
	flag.Parse()
	lv := map[string]int{"glance": glance, "detail": detail, "provenance": provenance}
	l, ok := lv[*level]
	if !ok {
		fmt.Fprintln(os.Stderr, "unknown level", *level)
		os.Exit(2)
	}
	in := io.Reader(os.Stdin)
	if flag.NArg() > 0 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()
		in = f
	}
	out, err := Render(in, l)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(out)
}

// Render reads one view's JSON and returns its Markdown at the level.
func Render(r io.Reader, level int) (string, error) {
	var v view
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		return "", err
	}
	switch v.View {
	case "authority":
		return authority(&v, level), nil
	case "assignment":
		return assignment(&v, level)
	}
	return "", fmt.Errorf("view %q is not authority or assignment", v.View)
}

// table writes a Markdown table with padded columns; right-aligns numeric columns.
func table(b *strings.Builder, head []string, rows [][]string, right ...int) {
	w := make([]int, len(head))
	for i, h := range head {
		w[i] = len([]rune(h))
	}
	for _, r := range rows {
		for i, c := range r {
			if n := len([]rune(c)); n > w[i] {
				w[i] = n
			}
		}
	}
	isRight := map[int]bool{}
	for _, i := range right {
		isRight[i] = true
	}
	line := func(cells []string) {
		b.WriteString("|")
		for i, c := range cells {
			pad := strings.Repeat(" ", w[i]-len([]rune(c)))
			if isRight[i] {
				b.WriteString(" " + pad + c + " |")
			} else {
				b.WriteString(" " + c + pad + " |")
			}
		}
		b.WriteString("\n")
	}
	line(head)
	b.WriteString("|")
	for i := range head {
		d := strings.Repeat("-", w[i])
		if isRight[i] {
			b.WriteString(" " + d[1:] + ": |")
		} else {
			b.WriteString(" " + d + " |")
		}
	}
	b.WriteString("\n")
	for _, r := range rows {
		line(r)
	}
	b.WriteString("\n")
}

func dash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func code(s string) string { return "`" + s + "`" }

func header(b *strings.Builder, title string, v *view) {
	fmt.Fprintf(b, "# %s\n\n", title)
	var p []string
	ref := v.Ref
	if len(ref) > 7 && !strings.ContainsAny(ref, "/") && !strings.EqualFold(ref, "main") {
		ref = ref[:7]
	}
	p = append(p, "ref "+code(ref))
	if !v.OnTrunk {
		p = append(p, "off the trunk: every task reads as proposed")
	}
	if v.Params.Task != "" {
		p = append(p, "task "+code(v.Params.Task))
	}
	if v.Params.Person != "" {
		p = append(p, "person "+v.Params.Person)
	}
	if v.Params.Proposed {
		p = append(p, "proposed tasks only")
	}
	b.WriteString(strings.Join(p, " · ") + "\n\n")
}

func authority(v *view, level int) string {
	var b strings.Builder
	header(&b, "Authority delegation", v)
	// The tree: the assignee shows where it changes from the parent's; a dot
	// marks the parent's assignee. The root always shows its own.
	b.WriteString("```\n")
	idw, tw, aw := 0, 0, 0
	for _, r := range v.Glance.Rows {
		if n := len(r.Title) + 2*r.Depth; n > tw {
			tw = n
		}
		if len(r.ID) > idw {
			idw = len(r.ID)
		}
		if n := len(r.Assignee); n > aw {
			aw = n
		}
	}
	for _, r := range v.Glance.Rows {
		who := "·"
		if r.Delegated || r.Depth == 0 {
			who = r.Assignee
		}
		mark := ""
		if r.Authorisation == "proposed" {
			mark = "proposed"
			if r.Differs {
				mark += " (differs from trunk)"
			}
		}
		left := strings.Repeat("  ", r.Depth) + r.ID + " " + r.Title
		line := fmt.Sprintf("%-*s  %-*s  %s", tw+idw+1, left, aw, who, mark)
		b.WriteString(strings.TrimRight(line, " ") + "\n")
	}
	b.WriteString("```\n\n")
	if level < detail {
		return b.String()
	}
	b.WriteString("## Chains and defaults\n\n")
	titles := map[string]string{}
	for _, r := range v.Glance.Rows {
		titles[r.ID] = r.Title
	}
	var rows [][]string
	for _, r := range v.Detail.Rows {
		rows = append(rows, []string{code(r.ID), titles[r.ID], chain(r.Authorities), defaults(r.JunctionDefaults)})
	}
	table(&b, []string{"Task", "Title", "Authority chain (nearest first)", "Junction defaults it states"}, rows)
	if level < provenance {
		return b.String()
	}
	b.WriteString("## Deciding commits\n\n")
	rows = nil
	for _, r := range v.Provenance.Rows {
		c := r.Commit
		if c == nil {
			rows = append(rows, []string{code(r.ID), "—", "—", "—", "—", "—", "—"})
			continue
		}
		rows = append(rows, []string{code(r.ID), code(c.Hash[:7]), c.Date, dash(r.Way), dash(r.By), c.Author, c.Committer})
	}
	table(&b, []string{"Task", "Commit", "Date", "Accepted as", "Authority is", "Author", "Committer"}, rows)
	return b.String()
}

func chain(a []string) string {
	if len(a) == 0 {
		return "—"
	}
	return strings.Join(a, ", ")
}

func defaults(d map[string]map[string]string) string {
	if len(d) == 0 {
		return "—"
	}
	var gates []string
	for g := range d {
		gates = append(gates, g)
	}
	sort.Strings(gates)
	var parts []string
	for _, g := range gates {
		var fields []string
		for f := range d[g] {
			fields = append(fields, f)
		}
		sort.Strings(fields)
		for _, f := range fields {
			parts = append(parts, fmt.Sprintf("%s %s: %s", g, f, d[g][f]))
		}
	}
	return strings.Join(parts, "; ")
}

func assignment(v *view, level int) (string, error) {
	var b strings.Builder
	header(&b, "Task assignment", v)
	var rows [][]string
	for _, p := range v.Glance.People {
		n := countAssigned(p.Assigned)
		models := "—"
		if len(p.Models) > 0 {
			var m []string
			for _, x := range p.Models {
				m = append(m, code(x))
			}
			models = strings.Join(m, ", ")
		}
		rows = append(rows, []string{p.Email, fmt.Sprint(n), fmt.Sprint(p.ContributesNxt), fmt.Sprint(p.ReviewsNxt), models})
	}
	table(&b, []string{"Email", "Assigned", "Contributes next", "Reviews next", "Models"}, rows, 1, 2, 3)
	if level < detail {
		return b.String(), nil
	}
	prov := map[string]junction{}
	for _, p := range v.Provenance.People {
		for _, j := range append(append([]junction{}, p.Contributes...), p.Reviews...) {
			prov[p.Email+"/"+j.Task+"/"+j.Gate] = j
		}
	}
	for _, p := range v.Detail.People {
		fmt.Fprintf(&b, "## %s\n\n", p.Email)
		var as []assigned
		if err := json.Unmarshal(p.Assigned, &as); err != nil && string(p.Assigned) != "null" && len(p.Assigned) > 0 {
			return "", err
		}
		b.WriteString("Assigned tasks\n\n")
		if len(as) == 0 {
			b.WriteString("none\n\n")
		} else {
			var r [][]string
			for _, a := range as {
				s := a.Status.Gate + ", " + a.Status.State
				if a.Status.Reason != "" {
					s += ", " + a.Status.Reason
				}
				r = append(r, []string{code(a.ID), a.Title, s})
			}
			table(&b, []string{"Task", "Title", "Status"}, r)
		}
		if len(p.Authority) > 0 {
			b.WriteString("Authority over\n\n")
			var r [][]string
			for _, s := range p.Authority {
				r = append(r, []string{code(s.ID), s.Title, fmt.Sprint(s.Descendants)})
			}
			table(&b, []string{"Subtree", "Title", "Descendants"}, r, 2)
		}
		for _, kind := range []struct {
			name string
			js   []junction
		}{{"Contributes", p.Contributes}, {"Reviews", p.Reviews}} {
			if len(kind.js) == 0 {
				continue
			}
			b.WriteString(kind.name + "\n\n")
			head := []string{"Task", "Gate", "Next", "Contributor", "Reviewer", "Model"}
			if level >= provenance {
				head = append(head, "Stated by")
			}
			var r [][]string
			for _, j := range kind.js {
				next := ""
				if j.Next {
					next = "next"
				}
				row := []string{code(j.Task), j.Gate, next, dash(j.Contributor), dash(j.Reviewer), dash(j.Model)}
				if level >= provenance {
					row = append(row, sources(prov[p.Email+"/"+j.Task+"/"+j.Gate].Sources))
				}
				r = append(r, row)
			}
			table(&b, head, r)
		}
	}
	return b.String(), nil
}

func countAssigned(raw json.RawMessage) int {
	var n int
	if json.Unmarshal(raw, &n) == nil {
		return n
	}
	var a []json.RawMessage
	_ = json.Unmarshal(raw, &a)
	return len(a)
}

func sources(m map[string]string) string {
	if len(m) == 0 {
		return "—"
	}
	var k []string
	for f := range m {
		k = append(k, f)
	}
	sort.Strings(k)
	var parts []string
	for _, f := range k {
		s := m[f]
		if s != "default" {
			s = code(s)
		}
		parts = append(parts, f+" "+s)
	}
	return strings.Join(parts, ", ")
}
