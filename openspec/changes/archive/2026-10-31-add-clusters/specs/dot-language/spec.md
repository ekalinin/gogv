## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: Attribute defaults and scope

`graph`, `node` and `edge` attribute statements SHALL set defaults for the objects created after them in the same graph or subgraph and in its subgraphs, as documented by Graphviz. Attributes given on an object SHALL override those defaults. A subgraph SHALL receive the graph attributes of its parent at the time of its definition, as documented by Graphviz.

#### Scenario: Node default inside a subgraph

- **WHEN** the input is `digraph { node [shape=rect]; a; subgraph s { node [shape=circle]; b } c }`
- **THEN** `a` and `c` have `shape=rect` and `b` has `shape=circle`

#### Scenario: Graph label received by a cluster

- **WHEN** the input is `digraph { label="G"; subgraph cluster_a { a } }`
- **THEN** `cluster_a` has the label `G`

### Requirement: Supported attributes

The system SHALL act on the following attributes and values:

- graphs and clusters: `label`, `fontname`, `fontsize`, `fontcolor`, `labeljust` (`c`, `l`, `r`)
- graphs: `nodesep`, `ranksep`, `rankdir` (`TB`, `LR`), `splines` (`spline`, `polyline`), `remincross`, `searchsize`
- clusters: `style` (`solid`), `color`
- subgraphs: `rank` (`same`, `min`, `max`, `source`, `sink`)
- nodes: `label`, `fontname`, `fontsize`, `fontcolor`, `shape` (`rect` and its synonyms `box` and `rectangle`, `plaintext`, `point`, `ellipse` and its synonym `oval`, `circle`, `diamond`, `hexagon`, `octagon`), `width`, `height`, `color`, `style` (`solid`, and `rounded` except on `diamond`, `hexagon` and `octagon`)
- edges: `minlen`, `color`, `style` (`solid`, `invis`), `constraint`, `arrowhead` and `arrowtail` (`none`, `empty`, `normal`), `dir` (`forward`, `back`, `both`, `none`), `arrowsize`

Colors SHALL be accepted in the form `#rrggbb`, in either letter case.

#### Scenario: Supported attributes give no warning

- **WHEN** the input is `digraph { rankdir=LR; a [shape=diamond, color="#00FF00"]; a -> b [arrowhead=empty, minlen=2] }`
- **THEN** the system lays out the graph without a warning

#### Scenario: Cluster attributes give no warning

- **WHEN** the input is `digraph { subgraph cluster_a { label="A"; labeljust=l; style=solid; color="#FF0000"; a; { rank=source; b } } }`
- **THEN** the system lays out the graph without a warning

### Requirement: Graphviz features that are not supported yet

For a feature that Graphviz documents but this system does not support yet, the system SHALL emit a warning and continue as if the feature were absent, using the documented default:

- an attribute of the Graphviz documentation that is not in the supported list, or a supported attribute with a value that is not supported: the attribute takes its default value
- a port or compass point on a node ID: the edge attaches to the node itself
- an HTML label: the object takes its default label
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
