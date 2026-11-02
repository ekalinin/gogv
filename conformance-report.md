# Conformance report

- Reference: Graphviz 16.1.0 built without Pango
- Cases: 565
- Corpora: curated 255, gopuml 260 (gopuml v0.3.0), gopuml-clusters 36 (gopuml v0.3.0), cli 14

## By capability

| Name | Cases | Passed | Known differences | Not supported yet | Failed |
|---|---|---|---|---|---|
| chen-eer-diagram | 24 | 23 | 1 | 0 | 0 |
| class-diagram | 9 | 9 | 0 | 0 | 0 |
| cli | 14 | 12 | 0 | 2 | 0 |
| description-diagram | 11 | 11 | 0 | 0 | 0 |
| diagram-source | 4 | 4 | 0 | 0 | 0 |
| dot-language | 104 | 67 | 2 | 35 | 0 |
| dot-layout | 132 | 122 | 9 | 1 | 0 |
| gopuml-clusters | 36 | 35 | 1 | 0 | 0 |
| graph-layout | 1 | 1 | 0 | 0 | 0 |
| legacy-activity-diagram | 13 | 13 | 0 | 0 | 0 |
| output-formats | 20 | 20 | 0 | 0 | 0 |
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
| canon | 551 | 515 | 0 | 36 | 0 |
| plain | 551 | 504 | 11 | 36 | 0 |
| plain-ext | 19 | 19 | 0 | 0 | 0 |
| svg | 551 | 489 | 26 | 36 | 0 |
| svg_inline | 19 | 19 | 0 | 0 | 0 |
| dot | 255 | 216 | 3 | 36 | 0 |
| gv | 19 | 19 | 0 | 0 | 0 |
| xdot | 255 | 216 | 3 | 36 | 0 |
| xdot1.2 | 19 | 19 | 0 | 0 | 0 |
| xdot1.4 | 19 | 19 | 0 | 0 | 0 |
| json | 255 | 216 | 3 | 36 | 0 |
| json0 | 19 | 19 | 0 | 0 | 0 |
| dot_json | 19 | 19 | 0 | 0 | 0 |
| xdot_json | 19 | 19 | 0 | 0 | 0 |

## Cases not passed

| Case | Outcome | Problems |
|---|---|---|
| dot-language/canon-canon-02 | not supported yet |  |
| dot-language/canon-canon-03 | known difference | plain: plain height 1.2889 is not within 5% of the reference 1.5467; svg: svg height 101 is not within 5% of the reference 119 SSIM 0.5505 is below 0.90; dot: bb height 92.8 is not within 5% of the reference 111.36; xdot: bb height 92.8 is not within 5% of the reference 111.36; json: bb height 92.8 is not within 5% of the reference 111.36 |
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
| dot-layout/port-flat-arc | known difference | plain: plain height 0.55556 is not within 5% of the reference 0.84095; svg: svg height 48 is not within 5% of the reference 69 SSIM 0.5692 is below 0.90; dot: bb height 40 is not within 5% of the reference 60.549; xdot: bb height 40 is not within 5% of the reference 60.549; json: bb height 40 is not within 5% of the reference 60.549 |
| dot-layout/table-edge-not-fixed | known difference | svg: SSIM 0.8848 is below 0.90 |
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
| vega-scxml/state3-1 | known difference | plain: plain height 5.7346 is not within 5% of the reference 6.0401; svg: SSIM 0.5923 is below 0.90 |
| vega-svg/id-component_01-1 | known difference | plain: plain height 0.42083 is not within 5% of the reference 0.46307; svg: svg height 38 is not within 5% of the reference 41 SSIM 0.5905 is below 0.90 |
| vega-svg/interactive-SVG0006_Svek-1 | known difference | plain: plain height 2.6806 is not within 5% of the reference 3.0929; svg: svg height 201 is not within 5% of the reference 231 SSIM 0.8028 is below 0.90 |
| gopuml-clusters/vega-scxml-state3-1 | known difference | plain: plain height 4.947 is not within 5% of the reference 4.1137; svg: svg height 364 is not within 5% of the reference 304 SSIM 0.4361 is below 0.90 |
| cli/other-engine | not supported yet |  |
| cli/default-attribute-flag | not supported yet |  |
