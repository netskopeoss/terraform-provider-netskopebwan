package main

import (
	"fmt"
	"html"
	"path"
	"strings"
)

// render wraps one page's body in the site template: a sidebar listing every page,
// and enough styling to read a schema without squinting.
func render(current page, pages []page, title string) string {
	var out strings.Builder

	fmt.Fprintf(&out, `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s · %s</title>
<style>%s</style>
</head>
<body>
<nav class="sidebar">
<a class="brand" href="%s">%s</a>
<input type="search" id="filter" placeholder="Filter…" autocomplete="off" aria-label="Filter pages">
`, html.EscapeString(current.Title), html.EscapeString(title), stylesheet, relative(current.Path, "index.html"), html.EscapeString(title))

	var section string

	for _, entry := range pages {
		if entry.Section != section {
			section = entry.Section
			fmt.Fprintf(&out, "<h2>%s</h2>\n<ul>\n", html.EscapeString(section))
		}

		class := ""
		if entry.Path == current.Path {
			class = ` class="current"`
		}

		fmt.Fprintf(&out, `<li><a%s href="%s">%s</a></li>`+"\n",
			class, relative(current.Path, entry.Path), html.EscapeString(entry.Title))
	}

	fmt.Fprintf(&out, `</ul>
</nav>
<main>
%s
</main>
<script>%s</script>
</body>
</html>
`, current.Body, filterScript)

	return out.String()
}

// relative renders a link from one page to another. Every path is relative to the
// site root, so the depth of the page doing the linking is all that matters.
func relative(from, to string) string {
	depth := strings.Count(path.Dir(from), "/") + 1
	if path.Dir(from) == "." {
		depth = 0
	}

	return strings.Repeat("../", depth) + to
}

// filterScript narrows the sidebar as you type. With one page per resource there
// are too many to scan, and a static site has nothing else to search with.
const filterScript = `
const filter = document.getElementById("filter");
filter.addEventListener("input", () => {
  const term = filter.value.toLowerCase();
  for (const item of document.querySelectorAll(".sidebar li")) {
    item.hidden = !item.textContent.toLowerCase().includes(term);
  }
  for (const heading of document.querySelectorAll(".sidebar h2")) {
    const list = heading.nextElementSibling;
    heading.hidden = !list || [...list.children].every((item) => item.hidden);
  }
});
`

const stylesheet = `
:root {
  --bg: #ffffff;
  --fg: #1c1e21;
  --muted: #5c6370;
  --line: #e3e5e8;
  --accent: #7b4bd8;
  --code-bg: #f5f6f7;
  --sidebar-bg: #fafbfc;
}
@media (prefers-color-scheme: dark) {
  :root {
    --bg: #16181c;
    --fg: #dfe2e6;
    --muted: #9aa3ae;
    --line: #2b2f36;
    --accent: #b28cf5;
    --code-bg: #1e2127;
    --sidebar-bg: #12141a;
  }
}
* { box-sizing: border-box; }
body {
  margin: 0;
  display: flex;
  align-items: flex-start;
  background: var(--bg);
  color: var(--fg);
  font: 16px/1.65 system-ui, -apple-system, "Segoe UI", sans-serif;
}
.sidebar {
  position: sticky;
  top: 0;
  flex: 0 0 17rem;
  height: 100vh;
  overflow-y: auto;
  padding: 1.25rem 1rem 3rem;
  border-right: 1px solid var(--line);
  background: var(--sidebar-bg);
  font-size: 14px;
}
.sidebar .brand {
  display: block;
  font-weight: 600;
  font-size: 15px;
  color: var(--fg);
  text-decoration: none;
  margin-bottom: 0.75rem;
}
.sidebar input {
  width: 100%;
  padding: 0.35rem 0.5rem;
  margin-bottom: 1rem;
  border: 1px solid var(--line);
  border-radius: 5px;
  background: var(--bg);
  color: var(--fg);
  font: inherit;
}
.sidebar h2 {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--muted);
  margin: 1.25rem 0 0.4rem;
}
.sidebar ul { list-style: none; margin: 0; padding: 0; }
.sidebar li { margin: 0; }
.sidebar a {
  display: block;
  padding: 0.2rem 0.4rem;
  border-radius: 4px;
  color: var(--fg);
  text-decoration: none;
  overflow-wrap: anywhere;
}
.sidebar a:hover { background: var(--code-bg); }
.sidebar a.current { background: var(--accent); color: #fff; }
main {
  flex: 1 1 auto;
  min-width: 0;
  max-width: 52rem;
  padding: 2rem 2.5rem 6rem;
}
main h1 { font-size: 1.8rem; margin: 0 0 1rem; }
main h2 { font-size: 1.3rem; margin: 2.2rem 0 0.6rem; padding-bottom: 0.3rem; border-bottom: 1px solid var(--line); }
main h3 { font-size: 1.05rem; margin: 1.8rem 0 0.4rem; }
main a { color: var(--accent); }
main ul { padding-left: 1.25rem; }
main li { margin: 0.3rem 0; }
main li > code:first-child { font-weight: 600; }
code {
  background: var(--code-bg);
  padding: 0.1em 0.35em;
  border-radius: 4px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.88em;
}
pre {
  background: var(--code-bg);
  padding: 0.9rem 1rem;
  border-radius: 6px;
  overflow-x: auto;
}
pre code { background: none; padding: 0; }
table { border-collapse: collapse; display: block; overflow-x: auto; max-width: 100%; }
th, td { border: 1px solid var(--line); padding: 0.4rem 0.6rem; text-align: left; }
hr { border: 0; border-top: 1px solid var(--line); margin: 2rem 0; }
@media (max-width: 60rem) {
  body { flex-direction: column; }
  .sidebar { position: static; height: auto; flex-basis: auto; width: 100%; border-right: 0; border-bottom: 1px solid var(--line); }
  main { padding: 1.5rem 1.25rem 4rem; }
}
`
