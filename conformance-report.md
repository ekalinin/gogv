# Conformance report

- Reference: Graphviz 16.1.0 built without Pango
- Cases: 511

## By capability

| Name | Cases | Passed | Known differences | Not supported yet | Failed |
|---|---|---|---|---|---|
| chen-eer-diagram | 24 | 23 | 1 | 0 | 0 |
| class-diagram | 9 | 8 | 0 | 1 | 0 |
| cli | 13 | 10 | 0 | 3 | 0 |
| description-diagram | 11 | 8 | 0 | 3 | 0 |
| diagram-source | 4 | 4 | 0 | 0 | 0 |
| dot-language | 94 | 61 | 2 | 31 | 0 |
| dot-layout | 108 | 99 | 8 | 1 | 0 |
| gopuml-clusters | 36 | 32 | 1 | 3 | 0 |
| graph-layout | 1 | 1 | 0 | 0 | 0 |
| legacy-activity-diagram | 13 | 13 | 0 | 0 | 0 |
| output-formats | 1 | 1 | 0 | 0 | 0 |
| preprocessor | 2 | 2 | 0 | 0 | 0 |
| state-diagram | 25 | 21 | 2 | 2 | 0 |
| styling | 3 | 3 | 0 | 0 | 0 |
| vega-graphml | 1 | 1 | 0 | 0 | 0 |
| vega-migration | 7 | 4 | 3 | 0 | 0 |
| vega-nonreg | 67 | 54 | 4 | 9 | 0 |
| vega-scxml | 8 | 6 | 2 | 0 | 0 |
| vega-state | 64 | 64 | 0 | 0 | 0 |
| vega-svg | 14 | 11 | 1 | 2 | 0 |
| vega-tikz | 2 | 2 | 0 | 0 | 0 |
| vega-xmi | 4 | 4 | 0 | 0 | 0 |

## By format

| Name | Cases | Passed | Known differences | Not supported yet | Failed |
|---|---|---|---|---|---|
| canon | 511 | 446 | 0 | 52 | 0 |
| plain | 511 | 437 | 9 | 52 | 0 |
| svg | 511 | 422 | 24 | 52 | 0 |

## Cases not passed

| Case | Outcome | Problems |
|---|---|---|
| dot-language/canon-canon-02 | not supported yet |  |
| dot-language/canon-canon-03 | known difference | plain: plain height 1.2889 is not within 5% of the reference 1.5467; svg: svg height 101 is not within 5% of the reference 119 SSIM 0.5505 is below 0.90 |
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
| dot-language/grammar-all | not supported yet |  |
| dot-language/grammar-several-graphs | not supported yet |  |
| dot-language/head-label-angle | not supported yet |  |
| dot-language/node-lists | known difference | svg: SSIM 0.7385 is below 0.90 |
| dot-language/table-label-forms | not supported yet |  |
| dot-layout/cluster-edges-between | known difference | svg: SSIM 0.4339 is below 0.90 |
| dot-layout/cluster-two-chains | known difference | svg: SSIM 0.8553 is below 0.90 |
| dot-layout/edge-label-cycle | known difference | plain: plain width 0.8358 is not within 5% of the reference 0.75; svg: svg width 68 is not within 5% of the reference 62 SSIM 0.5052 is below 0.90 |
| dot-layout/edge-label-long | known difference | svg: SSIM 0.8779 is below 0.90 |
| dot-layout/edge-label-minlen | known difference | svg: SSIM 0.8283 is below 0.90 |
| dot-layout/edge-label-parallel | known difference | svg: SSIM 0.8169 is below 0.90 |
| dot-layout/fontname-verdana | not supported yet |  |
| dot-layout/long-edges | known difference | svg: SSIM 0.4851 is below 0.90 |
| dot-layout/table-edge-not-fixed | known difference | svg: SSIM 0.8848 is below 0.90 |
| chen-eer-diagram/entities-relationship-and-attributes-1 | known difference | plain: plain width 4.6543 is not within 5% of the reference 3.3385; svg: svg width 343 is not within 5% of the reference 248 SSIM 0.7404 is below 0.90 |
| class-diagram/role-visibility-1 | not supported yet |  |
| description-diagram/linked-ports-1 | not supported yet |  |
| description-diagram/map-element-1 | not supported yet |  |
| description-diagram/short-forms-1 | not supported yet |  |
| state-diagram/choice-fork-join-1 | known difference | plain: plain width 4.1586 is not within 5% of the reference 3.5128; svg: svg width 307 is not within 5% of the reference 261 SSIM 0.7133 is below 0.90 |
| state-diagram/entry-exit-points-1 | not supported yet |  |
| state-diagram/history-style-1 | known difference | svg: SSIM 0.7064 is below 0.90 |
| state-diagram/map-element-1 | not supported yet |  |
| vega-migration/ChenMovieAlias_Test-1 | known difference | plain: plain width 22.968 is not within 5% of the reference 20.408; svg: svg width 1662 is not within 5% of the reference 1477 SSIM 0.8680 is below 0.90 |
| vega-migration/ChenMovieExtended_Test-1 | known difference | plain: plain width 23.607 is not within 5% of the reference 20.951; svg: svg width 1708 is not within 5% of the reference 1517 SSIM 0.8670 is below 0.90 |
| vega-migration/ChenMovie_Test-1 | known difference | svg: SSIM 0.8090 is below 0.90 |
| vega-nonreg/group2681-history_style_all_forms-1 | known difference | svg: SSIM 0.7064 is below 0.90 |
| vega-nonreg/group2904-all_visibilities-1 | not supported yet |  |
| vega-nonreg/group2904-diamond_at_tail-1 | not supported yet |  |
| vega-nonreg/group2904-icon_size_zero-1 | not supported yet |  |
| vega-nonreg/group2904-left_to_right-1 | not supported yet |  |
| vega-nonreg/group2904-reported_case-1 | not supported yet |  |
| vega-nonreg/group2904-role_tail_only-1 | not supported yet |  |
| vega-nonreg/group2904-role_without_visibility-1 | not supported yet |  |
| vega-nonreg/group2917-smetana_hide_state_plain_edges-1 | known difference | svg: SSIM 0.8669 is below 0.90 |
| vega-nonreg/simple-A0004-1 | known difference | svg: SSIM 0.8610 is below 0.90 |
| vega-nonreg/simple-ComponentExtraArrows_0001-1 | known difference | svg: SSIM 0.8178 is below 0.90 |
| vega-nonreg/simple-QualifiedAssoc001-1 | not supported yet |  |
| vega-nonreg/simple-QualifiedAssoc002-1 | not supported yet |  |
| vega-scxml/state2-1 | known difference | svg: SSIM 0.8736 is below 0.90 |
| vega-scxml/state3-1 | known difference | plain: plain height 5.7346 is not within 5% of the reference 6.0401; svg: SSIM 0.5923 is below 0.90 |
| vega-svg/id-component_01-1 | known difference | plain: plain height 0.42083 is not within 5% of the reference 0.46307; svg: svg height 38 is not within 5% of the reference 41 SSIM 0.5905 is below 0.90 |
| vega-svg/interactive-SVG0005_Svek-1 | not supported yet |  |
| vega-svg/interactive-SVG0006_Svek-1 | not supported yet |  |
| gopuml-clusters/description-diagram-linked-ports-1 | not supported yet |  |
| gopuml-clusters/state-diagram-entry-exit-points-1 | not supported yet |  |
| gopuml-clusters/vega-scxml-state3-1 | known difference | plain: plain height 4.947 is not within 5% of the reference 4.1137; svg: svg height 364 is not within 5% of the reference 304 SSIM 0.4361 is below 0.90 |
| gopuml-clusters/vega-svg-interactive-SVG0005_Svek-1 | not supported yet |  |
| cli/later-format | not supported yet |  |
| cli/other-engine | not supported yet |  |
| cli/default-attribute-flag | not supported yet |  |
