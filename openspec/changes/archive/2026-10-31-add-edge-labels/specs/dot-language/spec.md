## ADDED Requirements

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

## MODIFIED Requirements

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
