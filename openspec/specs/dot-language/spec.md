# dot-language Specification

## Purpose

Reads graphs written in the Graphviz DOT language, so that the layout, the outputs and `canon` see the same graph structure, attributes and labels as the reference `dot`.

## Requirements

### Requirement: DOT grammar

The system SHALL parse the whole DOT grammar as documented by Graphviz ("The DOT Language"):

- `graph`, `digraph` and `strict` graphs
- node, edge, attribute and `ID = ID` statements
- named and anonymous subgraphs
- edge chains whose operands are nodes or subgraphs
- node IDs with a port name and a compass point
- IDs written as identifiers, numerals, double-quoted strings and HTML strings
- `+` concatenation of double-quoted strings and backslash-newline continuation inside them
- C and C++ comments, and `#` with the rest of its line

Keywords SHALL be case-insensitive. An input with several graphs SHALL yield those graphs in input order. As the reference does, the rest of the line after the closing brace of a graph SHALL be ignored, and line numbers SHALL start again from 1 on the next line.

#### Scenario: Edge chain with a subgraph operand

- **WHEN** the input is `digraph { a -> { b c } -> d }`
- **THEN** the graph has the edges `a -> b`, `a -> c`, `b -> d` and `c -> d`

#### Scenario: Concatenated string

- **WHEN** a node has `label="Hel" + "lo"`
- **THEN** its label is `Hello`

#### Scenario: HTML string

- **WHEN** a node has `label=<<TABLE><TR><TD>x</TD></TR></TABLE>>`
- **THEN** the label is an HTML string, and `canon` writes it between `<` and `>`

#### Scenario: Two graphs in one input

- **WHEN** the input is `digraph A { a }` on one line and `digraph B { b }` on the next
- **THEN** the system yields graph `A`, then graph `B`

#### Scenario: Text after a graph on its line

- **WHEN** the input is `digraph A { a } digraph B { b }` on one line
- **THEN** the system yields graph `A` only

### Requirement: Attribute defaults and scope

`graph`, `node` and `edge` attribute statements SHALL set defaults for the objects created after them in the same graph or subgraph and in its subgraphs, as documented by Graphviz. Attributes given on an object SHALL override those defaults. A subgraph SHALL receive the graph attributes of its parent at the time of its definition, as documented by Graphviz.

#### Scenario: Node default inside a subgraph

- **WHEN** the input is `digraph { node [shape=rect]; a; subgraph s { node [shape=circle]; b } c }`
- **THEN** `a` and `c` have `shape=rect` and `b` has `shape=circle`

#### Scenario: Graph label received by a cluster

- **WHEN** the input is `digraph { label="G"; subgraph cluster_a { a } }`
- **THEN** `cluster_a` has the label `G`

### Requirement: Strict graphs

In a `strict` graph, the system SHALL keep at most one edge with a given tail and head (one edge per pair of nodes in an undirected graph). A later edge statement with the same two nodes SHALL apply its attributes to the existing edge, as documented by Graphviz.

#### Scenario: Repeated edge in a strict graph

- **WHEN** the input is `strict digraph { a -> b [color="#ff0000"]; a -> b [minlen=2] }`
- **THEN** the graph has one edge `a -> b`, with `color="#ff0000"` and `minlen=2`

### Requirement: Syntax errors

On a syntax error, the system SHALL report it in the reference's wording: the line and the token near which the error was found (`syntax error in line N near 'TOKEN'`), the line alone when the input ends too early (`syntax error in line N`), and the start of an unterminated quoted or HTML string. It SHALL produce no output for the graph that contains the error.

#### Scenario: Edge without a head

- **WHEN** the input is `digraph { a -> }`
- **THEN** the system reports `syntax error in line 1 near '}'` and produces no output

### Requirement: Plain-text labels

The `label` of nodes, edges, graphs and clusters, the `xlabel` of nodes and edges, and the `headlabel` and `taillabel` of edges SHALL be plain text unless they are HTML strings. In a plain-text label:

- `\N` SHALL be replaced by the node name, `\G` by the graph name, `\E` by the edge name, `\T` by the tail name, `\H` by the head name and `\L` by the object's label, as documented by Graphviz for escape strings
- `\n`, `\l` and `\r` SHALL end a line that is centered, left-justified and right-justified respectively, and the text after the last of them SHALL form a centered line when it is not empty

The default label of a node SHALL be `\N`.

#### Scenario: Default node label

- **WHEN** node `a` has no `label`
- **THEN** its label is `a`

#### Scenario: Escapes

- **WHEN** graph `G` has a node `a` with `label="\N in \G"`
- **THEN** the label of `a` is `a in G`

#### Scenario: Line breaks

- **WHEN** a node has `label="left\lright\rcenter\n"`
- **THEN** the label has three lines: `left` left-justified, `right` right-justified and `center` centered

#### Scenario: Escapes in an edge label

- **WHEN** an edge `a -> b` has `label="\T to \H"`
- **THEN** its label is `a to b`

### Requirement: Supported attributes

The system SHALL act on the following attributes and values:

- graphs and clusters: `label`, `fontname`, `fontsize`, `fontcolor`, `labeljust` (`c`, `l`, `r`)
- graphs: `nodesep`, `ranksep`, `rankdir` (`TB`, `LR`), `splines` (`spline`, `polyline`), `remincross`, `searchsize`
- clusters: `style` (`solid`), `color`
- subgraphs: `rank` (`same`, `min`, `max`, `source`, `sink`)
- nodes: `label`, `xlabel`, `fontname`, `fontsize`, `fontcolor`, `shape` (`rect` and its synonyms `box` and `rectangle`, `plaintext`, `point`, `ellipse` and its synonym `oval`, `circle`, `diamond`, `hexagon`, `octagon`), `width`, `height`, `color`, `style` (`solid`, and `rounded` except on `diamond`, `hexagon` and `octagon`)
- edges: `label`, `xlabel`, `headlabel`, `taillabel`, `fontname`, `fontsize`, `fontcolor`, `sametail`, `minlen`, `color`, `style` (`solid`, `invis`), `constraint`, `arrowhead` and `arrowtail` (`none`, `empty`, `normal`), `dir` (`forward`, `back`, `both`, `none`), `arrowsize`

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

### Requirement: Graphviz features that are not supported yet

For a feature that Graphviz documents but this system does not support yet, the system SHALL emit a warning and continue as if the feature were absent, using the documented default:

- an attribute of the Graphviz documentation that is not in the supported list, or a supported attribute with a value that is not supported: the attribute takes its default value
- a port or compass point on a node ID: the edge attaches to the node itself
- an HTML label that is not a single-cell table label, or an HTML label of the root graph: the object takes its default label
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

### Requirement: Cluster subgraphs

A subgraph whose name starts with `cluster`, compared without regard to case, SHALL be a cluster. A cluster SHALL be nested in the nearest enclosing cluster, through any subgraphs that are not clusters.

A node SHALL belong to at most one cluster and to the clusters that enclose it. When several clusters contain a node, it SHALL belong to the innermost of them; when two of them do not contain each other, the one that comes first in the input SHALL keep the node. A cluster SHALL be drawn only when it has a node.

When a node of a cluster is also in a subgraph with `rank` that is not inside that cluster, the system SHALL warn and take the node out of the cluster, as the reference does.

A node taken out of a cluster SHALL also be taken out of the subgraphs inside that cluster, and `canon` SHALL write the graph without it there, as the reference does.

#### Scenario: Cluster name in another case

- **WHEN** the graph has `subgraph Cluster_x { a }`
- **THEN** `Cluster_x` is a cluster

#### Scenario: Node in two sibling clusters

- **WHEN** the graph is `digraph { subgraph cluster_a { a } subgraph cluster_b { a b } }`
- **THEN** `a` belongs to `cluster_a` only, `cluster_b` contains only `b`, and `canon` writes `a` in `cluster_a` only

#### Scenario: Node in nested clusters

- **WHEN** the graph is `digraph { subgraph cluster_a { subgraph cluster_b { a } a c } }`
- **THEN** `a` belongs to `cluster_b` and `cluster_a`, and `c` belongs to `cluster_a` only

#### Scenario: Cluster without nodes

- **WHEN** the graph is `digraph { subgraph cluster_e { } a }`
- **THEN** `cluster_e` is not drawn

#### Scenario: Rank set across a cluster

- **WHEN** the graph is `digraph { subgraph cluster_x { a; b } c; { rank=same; b; c } }`
- **THEN** the system warns that `b` is taken out of `cluster_x`, `cluster_x` contains only `a`, and `canon` writes `b` outside `cluster_x`
### Requirement: Single-cell table labels

An HTML label SHALL be a table label when it is a `TABLE` element that holds one `TR` element that holds one empty `TD` element, and when the `TABLE` has no attributes other than `BGCOLOR` (a `#rrggbb` color), `FIXEDSIZE` (`TRUE` or `FALSE`), `WIDTH` and `HEIGHT` (in points) and the `TD` has none. Element and attribute names SHALL be compared without regard to case, and white space between the elements SHALL be allowed. A table label SHALL be accepted as the `label` and the `xlabel` of a node, as the `label`, `xlabel`, `headlabel` and `taillabel` of an edge, and as the `label` of a cluster.

#### Scenario: Table that gopuml writes

- **WHEN** an edge has `label=<<TABLE BGCOLOR="#000009" FIXEDSIZE="TRUE" WIDTH="57" HEIGHT="17"><TR><TD></TD></TR></TABLE>>`
- **THEN** the system lays out the graph without a warning, with a table label on the edge

#### Scenario: Cell with content

- **WHEN** a node `a` has `label=<<TABLE><TR><TD>x</TD></TR></TABLE>>`
- **THEN** the system warns that HTML labels of node `a` are not supported yet and gives `a` its default label

#### Scenario: Cell with a port

- **WHEN** a node `a` has `label=<<TABLE><TR><TD PORT="p"></TD></TR></TABLE>>`
- **THEN** the system warns that HTML labels of node `a` are not supported yet and gives `a` its default label
