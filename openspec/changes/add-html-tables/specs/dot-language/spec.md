## ADDED Requirements

### Requirement: Node ports

An edge end written `node:port`, or set by the edge's `tailport` or `headport`, SHALL name a port of the node: the `PORT` of a cell, or of the `TABLE`, of the node's table label. A name that the node defines as a port SHALL be that port, even when it is also the name of a compass point. When the node has no port of that name and the name is not a compass point, the system SHALL warn as the reference does, `node a, port P unrecognized`, once for each edge end that names it: the edges of one tail node after those of the previous one, in the order of the nodes, and the tail of an edge before its head. The edge SHALL then attach to the node itself, and this warning SHALL NOT say that a feature is not supported yet.

#### Scenario: Port of a cell

- **WHEN** a node `a` has `shape=plaintext, label=<<TABLE><TR><TD PORT="l"></TD><TD PORT="r"></TD></TR></TABLE>>` and an edge is written `a:r -> b`
- **THEN** the edge attaches to the cell `r` of `a`, without a warning

#### Scenario: Unknown port

- **WHEN** a node `a` has `shape=rect` and an edge is written `a:P -> b`
- **THEN** the system warns `node a, port P unrecognized`, and the edge attaches to `a` itself

#### Scenario: Port named like a compass point

- **WHEN** a node `a` has `shape=plaintext, label=<<TABLE><TR><TD PORT="n"></TD><TD WIDTH="40"></TD></TR></TABLE>>` and an edge is written `a:n -> b`
- **THEN** the edge attaches to the cell `n` of `a`, without a warning

## RENAMED Requirements

- FROM: `### Requirement: Single-cell table labels`
- TO: `### Requirement: Table labels`

## MODIFIED Requirements

### Requirement: Table labels

An HTML label SHALL be a table label when it is a `TABLE` element that holds one or more `TR` elements, each of which holds one or more empty `TD` elements, with these attributes only, all optional:

- on the `TABLE`: `BORDER`, `CELLBORDER`, `CELLSPACING` and `CELLPADDING` (whole points), `BGCOLOR` and `COLOR` (`#rrggbb` colors), `FIXEDSIZE` (`TRUE` or `FALSE`), `WIDTH` and `HEIGHT` (points), `PORT`
- on a `TD`: `BORDER`, `CELLSPACING` and `CELLPADDING` (whole points), `BGCOLOR` and `COLOR`, `COLSPAN` (a whole number of columns), `FIXEDSIZE`, `WIDTH` and `HEIGHT`, `PORT`
- on a `TR`: none

Element and attribute names SHALL be compared without regard to case, and white space between the elements SHALL be allowed. A table label SHALL be accepted as the `label` and the `xlabel` of a node, as the `label`, `xlabel`, `headlabel` and `taillabel` of an edge, and as the `label` of a cluster; its ports SHALL count only on the `label` of a node.

#### Scenario: Table that gopuml writes

- **WHEN** an edge has `label=<<TABLE BGCOLOR="#000009" FIXEDSIZE="TRUE" WIDTH="57" HEIGHT="17"><TR><TD></TD></TR></TABLE>>`
- **THEN** the system lays out the graph without a warning, with a table label on the edge

#### Scenario: Cell with content

- **WHEN** a node `a` has `label=<<TABLE><TR><TD>x</TD></TR></TABLE>>`
- **THEN** the system warns that HTML labels of node `a` are not supported yet and gives `a` its default label

#### Scenario: Cell with a port

- **WHEN** a node `a` has `label=<<TABLE><TR><TD PORT="p"></TD></TR></TABLE>>`
- **THEN** the system lays out the graph without a warning, with a table label on `a` whose cell has the port `p`

#### Scenario: Grid that gopuml writes

- **WHEN** a node has `shape=plaintext` and a label of three rows of three empty cells, with `BORDER="0" CELLBORDER="0" CELLSPACING="0" CELLPADDING="0"` on the `TABLE` and `FIXEDSIZE`, `WIDTH`, `HEIGHT`, `BGCOLOR` and `PORT` on its cells
- **THEN** the system lays out the graph without a warning

#### Scenario: Row span

- **WHEN** a node `a` has `label=<<TABLE><TR><TD ROWSPAN="2"></TD></TR><TR><TD></TD></TR></TABLE>>`
- **THEN** the system warns that HTML labels of node `a` are not supported yet and gives `a` its default label

### Requirement: Supported attributes

The system SHALL act on the following attributes and values:

- graphs and clusters: `label`, `fontname`, `fontsize`, `fontcolor`, `labeljust` (`c`, `l`, `r`)
- graphs: `nodesep`, `ranksep`, `rankdir` (`TB`, `LR`), `splines` (`spline`, `polyline`), `remincross`, `searchsize`
- clusters: `style` (`solid`), `color`
- subgraphs: `rank` (`same`, `min`, `max`, `source`, `sink`)
- nodes: `label`, `xlabel`, `fontname`, `fontsize`, `fontcolor`, `shape` (`rect` and its synonyms `box` and `rectangle`, `plaintext`, `point`, `ellipse` and its synonym `oval`, `circle`, `diamond`, `hexagon`, `octagon`), `width`, `height`, `color`, `style` (`solid`, and `rounded` except on `diamond`, `hexagon` and `octagon`)
- edges: `label`, `xlabel`, `headlabel`, `taillabel`, `fontname`, `fontsize`, `fontcolor`, `sametail`, `tailport` and `headport` (the `PORT` of a table cell of the node), `minlen`, `color`, `style` (`solid`, `invis`), `constraint`, `arrowhead` and `arrowtail` (`none`, `empty`, `normal`), `dir` (`forward`, `back`, `both`, `none`), `arrowsize`

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

### Requirement: Graphviz features that are not supported yet

For a feature that Graphviz documents but this system does not support yet, the system SHALL emit a warning and continue as if the feature were absent, using the documented default:

- an attribute of the Graphviz documentation that is not in the supported list, or a supported attribute with a value that is not supported: the attribute takes its default value
- a compass point on a node ID, alone or after a port: the edge attaches to the node itself, or to the port
- a port on a loop: the loop is drawn on the node itself
- an HTML label that is not a table label, or an HTML label of the root graph: the object takes its default label
- an undirected graph: its edges take the documented default `dir=none`
- `style=rounded` on a `diamond`, `hexagon` or `octagon` node: the node is drawn without rounded corners
- a `fontname` for which the reference has metrics that this system does not have yet (`DejaVu Sans`, `Verdana`, `Georgia`, `Consolas`, `Calibri`, `Trebuchet MS`, `Open Sans`, compared without regard to case): the text takes the Times metrics

The warning SHALL start with `Warning:`, SHALL name the object and the feature, SHALL say `not supported yet`, and SHALL name the default used when there is one. Attributes that Graphviz does not document SHALL be kept without a warning, as Graphviz keeps them.

#### Scenario: Unsupported shape

- **WHEN** a node `a` has `shape=record`
- **THEN** the system warns that `shape=record` of node `a` is not supported yet and draws `a` as an `ellipse`

#### Scenario: Cluster before its slice

- **WHEN** the graph has `subgraph cluster_x { label=<<TABLE><TR><TD>x</TD></TR></TABLE>>; a }`
- **THEN** the system warns that HTML labels are not supported yet and draws `cluster_x` without a label

#### Scenario: Unsupported cluster style

- **WHEN** `subgraph cluster_x { style=filled; a }` is in the graph
- **THEN** the system warns that `style=filled` of `cluster_x` is not supported yet and draws `cluster_x` with `style=solid`

#### Scenario: Font without metrics yet

- **WHEN** a node `a` has `fontname="Verdana"`
- **THEN** the system warns that `fontname=Verdana` of node `a` is not supported yet and measures its label with the Times metrics

#### Scenario: Attribute unknown to Graphviz

- **WHEN** a node has `foo=bar`
- **THEN** the system emits no warning, and `canon` keeps `foo=bar`

#### Scenario: Head label angle

- **WHEN** an edge `a -> b` has `headlabel="H"` and `labelangle=30`
- **THEN** the system warns that `labelangle=30` of edge `a->b` is not supported yet, and places `H` as without `labelangle`

#### Scenario: Compass point

- **WHEN** an edge is written `a:n -> b`
- **THEN** the system warns that `tailport=n` of edge `a->b` is not supported yet, and the edge attaches to `a` itself

#### Scenario: Port on a loop

- **WHEN** a node `a` has `shape=plaintext, label=<<TABLE><TR><TD PORT="p"></TD></TR></TABLE>>` and an edge is written `a:p -> a`
- **THEN** the system warns that `tailport=p` of edge `a->a` is not supported yet, and draws the loop on `a`
