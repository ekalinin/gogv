# Spec Delta

## Purpose

Reads graphs written in the Graphviz DOT language, so that the layout, the outputs and `canon` see the same graph structure, attributes and labels as the reference `dot`.

## ADDED Requirements

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

`graph`, `node` and `edge` attribute statements SHALL set defaults for the objects created after them in the same graph or subgraph and in its subgraphs, as documented by Graphviz. Attributes given on an object SHALL override those defaults.

#### Scenario: Node default inside a subgraph

- **WHEN** the input is `digraph { node [shape=rect]; a; subgraph s { node [shape=circle]; b } c }`
- **THEN** `a` and `c` have `shape=rect` and `b` has `shape=circle`

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

The `label` of nodes, edges, graphs and clusters SHALL be plain text unless it is an HTML string. In a plain-text label:

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

### Requirement: Supported attributes

The system SHALL act on the following attributes and values:

- graphs: `label`, `fontname`, `fontsize`, `fontcolor`, `nodesep`, `ranksep`, `rankdir` (`TB`, `LR`), `splines` (`spline`, `polyline`), `remincross`, `searchsize`
- nodes: `label`, `fontname`, `fontsize`, `fontcolor`, `shape` (`rect` and its synonyms `box` and `rectangle`, `plaintext`, `point`, `ellipse` and its synonym `oval`, `circle`, `diamond`, `hexagon`, `octagon`), `width`, `height`, `color`, `style` (`solid`, and `rounded` except on `diamond`, `hexagon` and `octagon`)
- edges: `minlen`, `color`, `style` (`solid`, `invis`), `constraint`, `arrowhead` and `arrowtail` (`none`, `empty`, `normal`), `dir` (`forward`, `back`, `both`, `none`), `arrowsize`

Colors SHALL be accepted in the form `#rrggbb`, in either letter case.

#### Scenario: Supported attributes give no warning

- **WHEN** the input is `digraph { rankdir=LR; a [shape=diamond, color="#00FF00"]; a -> b [arrowhead=empty, minlen=2] }`
- **THEN** the system lays out the graph without a warning

### Requirement: Graphviz features that are not supported yet

For a feature that Graphviz documents but this system does not support yet, the system SHALL emit a warning and continue as if the feature were absent, using the documented default:

- an attribute of the Graphviz documentation that is not in the supported list, or a supported attribute with a value that is not supported: the attribute takes its default value
- a cluster subgraph: it is laid out as a subgraph that is not a cluster
- a subgraph with `rank`: the rank constraint is ignored
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

- **WHEN** the graph has `subgraph cluster_x { a b }`
- **THEN** the system warns that clusters are not supported yet and lays out `a` and `b` without a cluster box

#### Scenario: Font without metrics yet

- **WHEN** a node `a` has `fontname="Verdana"`
- **THEN** the system warns that `fontname=Verdana` of node `a` is not supported yet and measures its label with the Times metrics

#### Scenario: Attribute unknown to Graphviz

- **WHEN** a node has `foo=bar`
- **THEN** the system emits no warning, and `canon` keeps `foo=bar`
