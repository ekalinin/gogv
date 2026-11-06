## MODIFIED Requirements

### Requirement: Supported attributes

The system SHALL act on the following attributes and values:

- graphs and clusters: `label`, `fontname`, `fontsize`, `fontcolor`, `labeljust` (`c`, `l`, `r`)
- graphs: `nodesep`, `ranksep`, `rankdir` (`TB`, `LR`), `splines` (`spline`, `polyline`), `remincross`, `searchsize`
- clusters: `style` (`solid`), `color`
- subgraphs: `rank` (`same`, `min`, `max`, `source`, `sink`)
- nodes: `label`, `xlabel`, `fontname`, `fontsize`, `fontcolor`, `shape` (`rect` and its synonyms `box` and `rectangle`, `box3d`, `plaintext`, `point`, `ellipse` and its synonym `oval`, `circle`, `diamond`, `hexagon`, `octagon`), `width`, `height`, `color`, `fillcolor`, `style` (`solid`, `filled`, and `rounded` except on `diamond`, `hexagon` and `octagon`), `tooltip`, `id`
- edges: `label`, `xlabel`, `headlabel`, `taillabel`, `fontname`, `fontsize`, `fontcolor`, `sametail`, `tailport` and `headport` (the `PORT` of a table cell of the node), `minlen`, `color`, `style` (`solid`, `invis`, `dotted`), `penwidth` (a number of points, at least 0), `weight` (a whole number, at least 0), `constraint`, `arrowhead` and `arrowtail` (`none`, `empty`, `normal`), `dir` (`forward`, `back`, `both`, `none`), `arrowsize`, `tooltip`, `labeltooltip`

`style` SHALL also be accepted as a comma-separated list of the values above, such as `filled,rounded`; a list with a value that is not supported is not supported.

Colors SHALL be accepted in the form `#rrggbb`, in either letter case.

#### Scenario: Supported attributes give no warning

- **WHEN** the input is `digraph { rankdir=LR; a [shape=diamond, color="#00FF00"]; a -> b [arrowhead=empty, minlen=2] }`
- **THEN** the system lays out the graph without a warning

#### Scenario: Cluster attributes give no warning

- **WHEN** the input is `digraph { subgraph cluster_a { label="A"; labeljust=l; style=solid; color="#FF0000"; a; { rank=source; b } } }`
- **THEN** the system lays out the graph without a warning

#### Scenario: Label attributes give no warning

- **WHEN** the input is `digraph { a [xlabel="X"]; a -> b [label="L", xlabel="X", headlabel="H", taillabel="T", fontname="Helvetica", fontsize=10, fontcolor="#0000FF", sametail=t] }`
- **THEN** the system lays out the graph without a warning

#### Scenario: Port attributes give no warning

- **WHEN** the input is `digraph { a [shape=plaintext, label=<<TABLE><TR><TD PORT="p"></TD></TR></TABLE>>]; a -> b [tailport=p] }`
- **THEN** the system lays out the graph without a warning

#### Scenario: Attributes that pprof writes give no warning

- **WHEN** the input is `digraph { node [style=filled fillcolor="#f8f8f8"]; N1 [label="main" id="node1" shape=box tooltip="main (10ms)" color="#b20e00" fillcolor="#edd7d5"]; N1_0 [label="work:hash" id="N1_0" shape=box3d tooltip="10ms"]; N1 -> N1_0 [label=" 10ms" weight=100 penwidth=2 tooltip="10ms" labeltooltip="10ms"]; N1_0 -> N1 [weight=2 style="dotted"] }`
- **THEN** the system lays out the graph without a warning

#### Scenario: List of styles

- **WHEN** a node has `shape=box, style="filled,rounded"`
- **THEN** the system lays out the graph without a warning, and the node is filled and rounded

#### Scenario: Style that is not supported in a list

- **WHEN** a node `a` has `style="filled,dashed"`
- **THEN** the system warns that `style=filled,dashed` of node `a` is not supported yet and draws `a` with `style=solid`
