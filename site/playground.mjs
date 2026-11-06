// SPDX-License-Identifier: MIT
// Provenance: gopuml site/playground.mjs

// Playground logic of the landing page, without the DOM, so that Node.js
// can test it: loading the module once, rendering sources as SVG, the
// examples, the warnings and errors, and the download of a result. page.mjs
// connects it to the page.

// examples are the graphs the visitor can pick; none of them gives a
// warning. The first is the example of the README, docs/example.gv.
export const examples = [
  {
    name: 'The stages of gogv',
    source: `digraph gogv {
	rankdir=LR
	node [shape=box, style=rounded, fontname=Helvetica, fontsize=12]
	edge [fontname=Helvetica, fontsize=10, color="#555555"]

	input [label="DOT text", shape=plaintext]

	subgraph cluster_language {
		label="dotlang"
		fontname=Helvetica
		color="#4a7ebb"
		lexer -> parser [label="tokens"]
	}

	subgraph cluster_layout {
		label="layout"
		fontname=Helvetica
		color="#5a9e4b"
		rank -> order -> position -> route
	}

	subgraph cluster_output {
		label="output"
		fontname=Helvetica
		color="#b36b2e"
		node [shape=ellipse]
		svg; xdot; json; canon
	}

	input -> lexer
	parser -> rank [label="graph"]
	parser -> canon [label="no layout"]
	route -> svg
	route -> xdot
	route -> json
}
`,
  },
  {
    name: 'Clusters',
    source: `digraph {
  node [shape=box]
  subgraph cluster_client {
    label="client"
    labeljust=l
    browser
    app
  }
  subgraph cluster_server {
    label="server"
    labeljust=l
    api
    worker
    db [shape=ellipse]
  }
  browser -> api
  app -> api
  api -> worker [label="jobs"]
  api -> db
  worker -> db
}
`,
  },
  {
    name: 'Left to right, with edge labels',
    source: `digraph {
  rankdir=LR
  node [shape=circle, width=0.9]
  start [shape=point, width=0.15]
  start -> idle
  idle -> running [label="start"]
  running -> done [label="finish"]
  done -> idle [label="reset"]
}
`,
  },
  {
    name: 'Table label with ports',
    source: `digraph {
  node [shape=plaintext]
  row [label=<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0">
    <TR>
      <TD PORT="a" WIDTH="40" HEIGHT="30" BGCOLOR="#cfe2f3"></TD>
      <TD PORT="b" WIDTH="40" HEIGHT="30" BGCOLOR="#d9ead3"></TD>
      <TD PORT="c" WIDTH="40" HEIGHT="30" BGCOLOR="#fce5cd"></TD>
    </TR>
  </TABLE>>]
  node [shape=ellipse]
  row:a -> x
  row:b -> y
  row:c -> z
}
`,
  },
  {
    name: 'Call graph of go tool pprof',
    source: `digraph "cpu" {
  node [style=filled fillcolor="#f8f8f8"]
  N1 [label="main\\n0 of 120ms (100%)" id="node1" shape=box fontsize=8 color="#b20000" fillcolor="#edd5d5" tooltip="main (120ms)"]
  N2 [label="hash\\n80ms (66.67%)" id="node2" shape=box fontsize=20 color="#b21600" fillcolor="#edd6d5" tooltip="hash (80ms)"]
  N3 [label="sort\\n40ms (33.33%)" id="node3" shape=box fontsize=14 color="#b23d00" fillcolor="#edddd5" tooltip="sort (40ms)"]
  N2_0 [label="work:hash" id="N2_0" shape=box3d fontsize=8 tooltip="80ms"]
  N1 -> N2 [label=" 80ms" weight=67 penwidth=4 color="#b21600" tooltip="main -> hash (80ms)" labeltooltip="main -> hash (80ms)"]
  N1 -> N3 [label=" 40ms" weight=34 penwidth=2 color="#b23d00" tooltip="main -> sort (40ms)" labeltooltip="main -> sort (40ms)"]
  N2 -> N2_0 [label=" 80ms" weight=100 style="dotted" tooltip="80ms"]
}
`,
  },
];

// describe returns the text of a render error: its message, with its line
// when the message does not name it.
export function describe(error) {
  if (error.line && !error.message.includes(`line ${error.line}`)) {
    return `Line ${error.line}: ${error.message}`;
  }
  return error.message;
}

// download returns the file of a successful result, or null for an error.
export function download(result) {
  if (!result || result.error || !result.svg) {
    return null;
  }
  return { filename: 'graph.svg', blob: result.svg };
}

// createPlayground returns a playground whose module comes from load, a
// function that resolves to an object with version and render(source), as
// load() of gogv.mjs does. Nothing is loaded before activate or render.
export function createPlayground(load) {
  let module;
  let version = null;
  let status = 'idle';
  let latest = 0;
  return {
    // status is idle, loading, ready or failed.
    get status() {
      return status;
    },

    // version is the release of the module once it is loaded, else null.
    get version() {
      return version;
    },

    // activate loads the module once. It resolves when it is loaded and
    // rejects when it cannot be.
    activate() {
      if (!module) {
        status = 'loading';
        module = Promise.resolve()
          .then(load)
          .then(
            (m) => {
              status = 'ready';
              version = m.version;
              return m;
            },
            (err) => {
              status = 'failed';
              throw err;
            },
          );
      }
      return module;
    },

    // render renders source as SVG. It resolves to null when a newer render
    // started in the meantime, else to { svg, warnings, error }: svg is a
    // Blob of the SVG, or null after an error; warnings are the texts of the
    // warnings, as the command line writes them; error is null or
    // { message, line, text }.
    async render(source) {
      const n = ++latest;
      const gogv = await this.activate();
      const r = await gogv.render(source);
      if (n !== latest) {
        return null;
      }
      return {
        svg: r.svg ? new Blob([r.svg], { type: 'image/svg+xml' }) : null,
        warnings: r.warnings.map((w) => `Warning: ${w}`),
        error: r.error ? { ...r.error, text: describe(r.error) } : null,
      };
    },
  };
}
