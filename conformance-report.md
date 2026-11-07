# Conformance report

- Reference: Graphviz 16.1.0 built without Pango
- Cases: 607
- Corpora: curated 297, gopuml 260 (gopuml v0.3.0), gopuml-clusters 36 (gopuml v0.3.0), cli 14

## By capability

| Name | Cases | Passed | Known differences | Not supported yet | Failed |
|---|---|---|---|---|---|
| chen-eer-diagram | 24 | 23 | 1 | 0 | 0 |
| class-diagram | 9 | 9 | 0 | 0 | 0 |
| cli | 14 | 12 | 0 | 2 | 0 |
| description-diagram | 11 | 11 | 0 | 0 | 0 |
| diagram-source | 4 | 4 | 0 | 0 | 0 |
| dot-language | 107 | 68 | 3 | 36 | 0 |
| dot-layout | 159 | 130 | 28 | 1 | 0 |
| gopuml-clusters | 36 | 35 | 1 | 0 | 0 |
| graph-layout | 1 | 1 | 0 | 0 | 0 |
| legacy-activity-diagram | 13 | 13 | 0 | 0 | 0 |
| output-formats | 25 | 24 | 1 | 0 | 0 |
| pprof | 7 | 3 | 4 | 0 | 0 |
| preprocessor | 2 | 2 | 0 | 0 | 0 |
| state-diagram | 25 | 23 | 2 | 0 | 0 |
| styling | 3 | 3 | 0 | 0 | 0 |
| vega-graphml | 1 | 1 | 0 | 0 | 0 |
| vega-migration | 7 | 4 | 3 | 0 | 0 |
| vega-nonreg | 67 | 63 | 4 | 0 | 0 |
| vega-scxml | 8 | 6 | 2 | 0 | 0 |
| vega-state | 64 | 64 | 0 | 0 | 0 |
| vega-svg | 14 | 12 | 2 | 0 | 0 |
| vega-tikz | 2 | 2 | 0 | 0 | 0 |
| vega-xmi | 4 | 4 | 0 | 0 | 0 |

## By format

| Name | Cases | Passed | Known differences | Not supported yet | Failed |
|---|---|---|---|---|---|
| canon | 593 | 556 | 0 | 37 | 0 |
| plain | 593 | 539 | 17 | 37 | 0 |
| plain-ext | 24 | 24 | 0 | 0 | 0 |
| svg | 593 | 507 | 49 | 37 | 0 |
| svg_inline | 24 | 23 | 1 | 0 | 0 |
| dot | 297 | 251 | 9 | 37 | 0 |
| gv | 24 | 24 | 0 | 0 | 0 |
| xdot | 297 | 251 | 9 | 37 | 0 |
| xdot1.2 | 24 | 24 | 0 | 0 | 0 |
| xdot1.4 | 24 | 24 | 0 | 0 | 0 |
| json | 297 | 251 | 9 | 37 | 0 |
| json0 | 24 | 24 | 0 | 0 | 0 |
| dot_json | 24 | 24 | 0 | 0 | 0 |
| xdot_json | 24 | 24 | 0 | 0 | 0 |

## gopuml contract proxy

273 of 296 cases of the gopuml corpora pass the proxy of gopuml's contract: a `plain` size within 2%, every node and edge label center within 2% of the graph size, and an SVG at SSIM 0.90 or more. The proxy does not change the outcomes.

| Case | Largest deviation | Object | SSIM |
|---|---|---|---|
| chen-eer-diagram/entities-relationship-and-attributes-1 | 53.9% | label of sh0010->sh0006#0 | 0.7404 |
| vega-migration/ChenMovie_Test-1 | 48.0% | node sh0007 | 0.8090 |
| gopuml-clusters/vega-scxml-state3-1 | 37.1% | label of sh0025->sh0017#0 | 0.4358 |
| vega-scxml/state3-1 | 37.1% | label of sh0025->sh0017#0 | 0.5920 |
| vega-migration/ChenMovieExtended_Test-1 | 30.8% | node sh0013 | 0.8670 |
| vega-migration/ChenMovieAlias_Test-1 | 28.5% | node sh0013 | 0.8680 |
| state-diagram/choice-fork-join-1 | 24.7% | node sh0006 | 0.7133 |
| state-diagram/junction-1 | 21.2% | label of sh0006->sh0008#0 | 0.9190 |
| gopuml-clusters/state-diagram-history-state-1 | 18.2% | label of sh0014->sh0013#0 | 0.9738 |
| state-diagram/history-state-1 | 18.2% | label of sh0014->sh0013#0 | 0.9753 |
| vega-svg/interactive-SVG0006_Svek-1 | 17.4% | label of sh0006->sh0007#0 | 0.8028 |
| vega-scxml/junction-pseudostate-1 | 12.3% | label of sh0006->sh0008#0 | 0.9084 |
| vega-state/junction-pseudostate-1 | 12.3% | label of sh0006->sh0008#0 | 0.9084 |
| gopuml-clusters/vega-scxml-state4-1 | 10.6% | node sh0010 | 0.9786 |
| vega-scxml/state4-1 | 10.6% | node sh0010 | 0.9786 |
| vega-svg/id-component_01-1 | 9.1% | graph height | 0.5905 |
| vega-nonreg/simple-A0004-1 | 8.0% | node sh0012 | 0.8610 |
| state-diagram/history-style-1 | 6.8% | node sh0006 | 0.7064 |
| vega-nonreg/group2681-history_style_all_forms-1 | 6.8% | node sh0006 | 0.7064 |
| vega-nonreg/simple-ComponentExtraArrows_0001-1 | 4.1% | node sh0007 | 0.8178 |
| vega-scxml/state2-1 | 4.1% | graph width | 0.8736 |
| chen-eer-diagram/bold-only-from-the-left-1 | 3.9% | label of sh0006->sh0008#0 | 0.9194 |
| vega-nonreg/group2917-smetana_hide_state_plain_edges-1 | 1.0% | label of sh0007->sh0008#0 | 0.8669 |

## Cases not passed

| Case | Outcome | Problems |
|---|---|---|
| dot-language/canon-canon-02 | not supported yet |  |
| dot-language/canon-canon-03 | known difference | plain: plain height 1.2889 is not within 5% of the reference 1.5467; svg: svg height 101 is not within 5% of the reference 119 SSIM 0.5358 is below 0.90; dot: bb height 92.8 is not within 5% of the reference 111.36; xdot: bb height 92.8 is not within 5% of the reference 111.36; json: bb height 92.8 is not within 5% of the reference 111.36 |
| dot-language/canon-canon-04 | not supported yet |  |
| dot-language/canon-canon-05 | not supported yet |  |
| dot-language/canon-canon-06 | not supported yet |  |
| dot-language/canon-canon-08 | not supported yet |  |
| dot-language/canon-canon-10 | not supported yet |  |
| dot-language/canon-canon-11 | not supported yet |  |
| dot-language/canon-canon2-17 | not supported yet |  |
| dot-language/canon-canon2-18 | not supported yet |  |
| dot-language/canon-canon2-20 | not supported yet |  |
| dot-language/canon-canon3-B2 | not supported yet |  |
| dot-language/canon-canon3-F | not supported yet |  |
| dot-language/canon-canon4-M | not supported yet |  |
| dot-language/canon-canon5-N10 | not supported yet |  |
| dot-language/canon-canon5-N11 | not supported yet |  |
| dot-language/canon-canon5-N12 | not supported yet |  |
| dot-language/canon-canon5-N3 | not supported yet |  |
| dot-language/canon-canon5-N6 | not supported yet |  |
| dot-language/canon-p-P1 | not supported yet |  |
| dot-language/canon-p-P10 | not supported yet |  |
| dot-language/canon-p-P11 | not supported yet |  |
| dot-language/canon-p-P12 | not supported yet |  |
| dot-language/canon-p-P4 | not supported yet |  |
| dot-language/canon-p-P5 | not supported yet |  |
| dot-language/canon-p-P7 | not supported yet |  |
| dot-language/cluster-html-label | not supported yet |  |
| dot-language/cluster-style-filled | not supported yet |  |
| dot-language/compass-point | not supported yet |  |
| dot-language/grammar-all | not supported yet |  |
| dot-language/grammar-several-graphs | not supported yet |  |
| dot-language/head-label-angle | not supported yet |  |
| dot-language/node-lists | known difference | svg: SSIM 0.7385 is below 0.90 |
| dot-language/port-loop | not supported yet |  |
| dot-language/pprof-attributes | known difference | plain: plain width 1.1838 is not within 5% of the reference 1.0323; svg: svg width 93 is not within 5% of the reference 82 SSIM 0.5281 is below 0.90; dot: bb width 85.236 is not within 5% of the reference 74.324; xdot: bb width 85.236 is not within 5% of the reference 74.324; json: bb width 85.236 is not within 5% of the reference 74.324 |
| dot-language/style-list-unsupported | not supported yet |  |
| dot-language/table-bad-values | not supported yet |  |
| dot-language/table-label-forms | not supported yet |  |
| dot-language/table-row-span | not supported yet |  |
| dot-layout/cluster-edges-between | known difference | svg: SSIM 0.4339 is below 0.90 |
| dot-layout/cluster-two-chains | known difference | svg: SSIM 0.8553 is below 0.90 |
| dot-layout/edge-label-cycle | known difference | plain: plain width 0.8358 is not within 5% of the reference 0.75; svg: svg width 68 is not within 5% of the reference 62 SSIM 0.5052 is below 0.90; dot: bb width 60.178 is not within 5% of the reference 54; xdot: bb width 60.178 is not within 5% of the reference 54; json: bb width 60.178 is not within 5% of the reference 54 |
| dot-layout/edge-label-long | known difference | svg: SSIM 0.8779 is below 0.90 |
| dot-layout/edge-label-minlen | known difference | svg: SSIM 0.8283 is below 0.90 |
| dot-layout/edge-label-parallel | known difference | svg: SSIM 0.8169 is below 0.90 |
| dot-layout/fontname-verdana | not supported yet |  |
| dot-layout/long-edges | known difference | svg: SSIM 0.4851 is below 0.90 |
| dot-layout/port-flat-arc | known difference |  |
| dot-layout/reduced-ChenMovieAlias_Test-1 | known difference | svg: SSIM 0.8223 is below 0.90 |
| dot-layout/reduced-ChenMovieAlias_Test-1-worst | known difference | svg: SSIM 0.6226 is below 0.90 |
| dot-layout/reduced-ChenMovie_Test-1 | known difference | svg: SSIM 0.8738 is below 0.90 |
| dot-layout/reduced-ChenMovie_Test-1-worst | known difference | svg: SSIM 0.7758 is below 0.90 |
| dot-layout/reduced-choice-fork-join-1 | known difference |  |
| dot-layout/reduced-choice-fork-join-1-worst | known difference | svg: SSIM 0.5538 is below 0.90 |
| dot-layout/reduced-entities-relationship-and-attributes-1 | known difference | svg: SSIM 0.8195 is below 0.90 |
| dot-layout/reduced-entities-relationship-and-attributes-1-worst | known difference | svg: SSIM 0.8213 is below 0.90 |
| dot-layout/reduced-history-style-1 | known difference | svg: SSIM 0.8630 is below 0.90 |
| dot-layout/reduced-history-style-1-worst | known difference | svg: SSIM 0.8623 is below 0.90 |
| dot-layout/reduced-interactive-SVG0006_Svek-1 | known difference | svg: SSIM 0.8008 is below 0.90 |
| dot-layout/reduced-interactive-SVG0006_Svek-1-worst | known difference | svg: SSIM 0.8907 is below 0.90 |
| dot-layout/reduced-simple-A0004-1 | known difference | svg: SSIM 0.5258 is below 0.90 |
| dot-layout/reduced-simple-A0004-1-worst | known difference | plain: plain width 2.4592 is not within 5% of the reference 2.6176; svg: svg width 185 is not within 5% of the reference 196 SSIM 0.7143 is below 0.90; dot: bb width 177.06 is not within 5% of the reference 188.47; xdot: bb width 177.06 is not within 5% of the reference 188.47; json: bb width 177.06 is not within 5% of the reference 188.47 |
| dot-layout/reduced-simple-ComponentExtraArrows_0001-1 | known difference | plain: plain width 2.0739 is not within 5% of the reference 1.9353; svg: svg width 157 is not within 5% of the reference 147 SSIM 0.6155 is below 0.90; dot: bb width 149.32 is not within 5% of the reference 139.34; xdot: bb width 149.32 is not within 5% of the reference 139.34; json: bb width 149.32 is not within 5% of the reference 139.34 |
| dot-layout/reduced-simple-ComponentExtraArrows_0001-1-worst | known difference | svg: SSIM 0.7624 is below 0.90 |
| dot-layout/reduced-state2-1 | known difference | svg: SSIM 0.6513 is below 0.90 |
| dot-layout/reduced-state2-1-worst | known difference | svg: SSIM 0.7130 is below 0.90 |
| dot-layout/reduced-state3-1-worst | known difference | plain: plain width 2.4149 is not within 5% of the reference 2.2344; svg: svg width 182 is not within 5% of the reference 169 SSIM 0.6828 is below 0.90; dot: bb width 173.87 is not within 5% of the reference 160.87; xdot: bb width 173.87 is not within 5% of the reference 160.87; json: bb width 173.87 is not within 5% of the reference 160.87 |
| dot-layout/table-edge-not-fixed | known difference | svg: SSIM 0.8848 is below 0.90 |
| output-formats/tooltips-ids | known difference | svg: SSIM 0.8994 is below 0.90; svg_inline: SSIM 0.8994 is below 0.90 |
| pprof/cpu | known difference | plain: plain width 27.414 is not within 5% of the reference 24.544; svg: svg width 1982 is not within 5% of the reference 1775 SSIM 0.8179 is below 0.90; dot: bb width 1973.8 is not within 5% of the reference 1767.1; xdot: bb width 1973.8 is not within 5% of the reference 1767.1; json: bb width 1973.8 is not within 5% of the reference 1767.1 |
| pprof/cpu-diff | known difference | plain: plain width 27.083 is not within 5% of the reference 19.167; svg: svg width 1958 is not within 5% of the reference 1388 SSIM 0.7666 is below 0.90; dot: bb width 1950 is not within 5% of the reference 1380; xdot: bb width 1950 is not within 5% of the reference 1380; json: bb width 1950 is not within 5% of the reference 1380 |
| pprof/cpu-focus | known difference | plain: plain width 11.397 is not within 5% of the reference 12.022; svg: svg width 829 is not within 5% of the reference 874 SSIM 0.8835 is below 0.90; dot: bb width 820.6 is not within 5% of the reference 865.6; xdot: bb width 820.6 is not within 5% of the reference 865.6; json: bb width 820.6 is not within 5% of the reference 865.6 |
| pprof/heap-nodecount | known difference | svg: SSIM 0.8458 is below 0.90 |
| chen-eer-diagram/entities-relationship-and-attributes-1 | known difference | plain: plain width 4.6543 is not within 5% of the reference 3.3385; svg: svg width 343 is not within 5% of the reference 248 SSIM 0.7404 is below 0.90 |
| state-diagram/choice-fork-join-1 | known difference | plain: plain width 4.1586 is not within 5% of the reference 3.5128; svg: svg width 307 is not within 5% of the reference 261 SSIM 0.7133 is below 0.90 |
| state-diagram/history-style-1 | known difference | svg: SSIM 0.7064 is below 0.90 |
| vega-migration/ChenMovieAlias_Test-1 | known difference | plain: plain width 22.968 is not within 5% of the reference 20.408; svg: svg width 1662 is not within 5% of the reference 1477 SSIM 0.8680 is below 0.90 |
| vega-migration/ChenMovieExtended_Test-1 | known difference | plain: plain width 23.607 is not within 5% of the reference 20.951; svg: svg width 1708 is not within 5% of the reference 1517 SSIM 0.8670 is below 0.90 |
| vega-migration/ChenMovie_Test-1 | known difference | svg: SSIM 0.8090 is below 0.90 |
| vega-nonreg/group2681-history_style_all_forms-1 | known difference | svg: SSIM 0.7064 is below 0.90 |
| vega-nonreg/group2917-smetana_hide_state_plain_edges-1 | known difference | svg: SSIM 0.8669 is below 0.90 |
| vega-nonreg/simple-A0004-1 | known difference | svg: SSIM 0.8610 is below 0.90 |
| vega-nonreg/simple-ComponentExtraArrows_0001-1 | known difference | svg: SSIM 0.8178 is below 0.90 |
| vega-scxml/state2-1 | known difference | svg: SSIM 0.8736 is below 0.90 |
| vega-scxml/state3-1 | known difference | plain: plain height 5.7346 is not within 5% of the reference 6.0401; svg: SSIM 0.5920 is below 0.90 |
| vega-svg/id-component_01-1 | known difference | plain: plain height 0.42083 is not within 5% of the reference 0.46307; svg: svg height 38 is not within 5% of the reference 41 SSIM 0.5905 is below 0.90 |
| vega-svg/interactive-SVG0006_Svek-1 | known difference | plain: plain height 2.6806 is not within 5% of the reference 3.0929; svg: svg height 201 is not within 5% of the reference 231 SSIM 0.8028 is below 0.90 |
| gopuml-clusters/vega-scxml-state3-1 | known difference | plain: plain height 4.947 is not within 5% of the reference 4.1137; svg: svg height 364 is not within 5% of the reference 304 SSIM 0.4358 is below 0.90 |
| cli/other-engine | not supported yet |  |
| cli/default-attribute-flag | not supported yet |  |
