# Conformance report

- Reference: Graphviz 16.1.0 built without Pango
- Cases: 398

## By capability

| Name | Cases | Passed | Known differences | Not supported yet | Failed |
|---|---|---|---|---|---|
| chen-eer-diagram | 24 | 12 | 0 | 12 | 0 |
| class-diagram | 9 | 5 | 0 | 4 | 0 |
| cli | 13 | 10 | 0 | 3 | 0 |
| description-diagram | 11 | 6 | 0 | 5 | 0 |
| diagram-source | 4 | 2 | 0 | 2 | 0 |
| dot-language | 78 | 48 | 1 | 29 | 0 |
| dot-layout | 47 | 45 | 1 | 1 | 0 |
| graph-layout | 1 | 1 | 0 | 0 | 0 |
| legacy-activity-diagram | 13 | 1 | 0 | 12 | 0 |
| output-formats | 1 | 1 | 0 | 0 | 0 |
| preprocessor | 2 | 1 | 0 | 1 | 0 |
| state-diagram | 25 | 0 | 0 | 25 | 0 |
| styling | 3 | 3 | 0 | 0 | 0 |
| vega-graphml | 1 | 1 | 0 | 0 | 0 |
| vega-migration | 7 | 1 | 0 | 6 | 0 |
| vega-nonreg | 67 | 19 | 0 | 48 | 0 |
| vega-scxml | 8 | 0 | 0 | 8 | 0 |
| vega-state | 64 | 1 | 0 | 63 | 0 |
| vega-svg | 14 | 1 | 0 | 13 | 0 |
| vega-tikz | 2 | 2 | 0 | 0 | 0 |
| vega-xmi | 4 | 4 | 0 | 0 | 0 |

## By format

| Name | Cases | Passed | Known differences | Not supported yet | Failed |
|---|---|---|---|---|---|
| canon | 398 | 156 | 0 | 229 | 0 |
| plain | 398 | 156 | 0 | 229 | 0 |
| svg | 398 | 154 | 2 | 229 | 0 |

## Cases not passed

| Case | Outcome | Problems |
|---|---|---|
| dot-language/canon-canon-02 | not supported yet |  |
| dot-language/canon-canon-03 | not supported yet |  |
| dot-language/canon-canon-04 | not supported yet |  |
| dot-language/canon-canon-05 | not supported yet |  |
| dot-language/canon-canon-06 | not supported yet |  |
| dot-language/canon-canon-08 | not supported yet |  |
| dot-language/canon-canon-10 | not supported yet |  |
| dot-language/canon-canon-11 | not supported yet |  |
| dot-language/canon-canon2-17 | not supported yet |  |
| dot-language/canon-canon2-18 | not supported yet |  |
| dot-language/canon-canon2-20 | not supported yet |  |
| dot-language/canon-canon2-23 | not supported yet |  |
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
| dot-language/grammar-all | not supported yet |  |
| dot-language/grammar-several-graphs | not supported yet |  |
| dot-language/node-lists | known difference | svg: SSIM 0.7385 is below 0.90 |
| dot-layout/fontname-verdana | not supported yet |  |
| dot-layout/long-edges | known difference | svg: SSIM 0.4851 is below 0.90 |
| chen-eer-diagram/association-color-1 | not supported yet |  |
| chen-eer-diagram/bold-only-from-the-left-1 | not supported yet |  |
| chen-eer-diagram/cardinalities-1 | not supported yet |  |
| chen-eer-diagram/display-names-1 | not supported yet |  |
| chen-eer-diagram/element-styles-1 | not supported yet |  |
| chen-eer-diagram/embedded-in-note-1 | not supported yet |  |
| chen-eer-diagram/entities-relationship-and-attributes-1 | not supported yet |  |
| chen-eer-diagram/entity-and-relationship-colors-1 | not supported yet |  |
| chen-eer-diagram/left-to-right-1 | not supported yet |  |
| chen-eer-diagram/png-output-1 | not supported yet |  |
| chen-eer-diagram/total-participation-1 | not supported yet |  |
| chen-eer-diagram/weak-entity-and-identifying-relationship-1 | not supported yet |  |
| class-diagram/composition-cardinalities-1 | not supported yet |  |
| class-diagram/nested-package-1 | not supported yet |  |
| class-diagram/package-code-as-display-1 | not supported yet |  |
| class-diagram/role-visibility-1 | not supported yet |  |
| description-diagram/component-inside-node-1 | not supported yet |  |
| description-diagram/dotted-link-direction-1 | not supported yet |  |
| description-diagram/linked-ports-1 | not supported yet |  |
| description-diagram/map-element-1 | not supported yet |  |
| description-diagram/short-forms-1 | not supported yet |  |
| diagram-source/legacy-activity-before-state-1 | not supported yet |  |
| diagram-source/state-detected-1 | not supported yet |  |
| legacy-activity-diagram/activity-colors-1 | not supported yet |  |
| legacy-activity-diagram/activity-in-partition-1 | not supported yet |  |
| legacy-activity-diagram/activity-several-lines-1 | not supported yet |  |
| legacy-activity-diagram/activity-with-code-1 | not supported yet |  |
| legacy-activity-diagram/branch-two-outcomes-1 | not supported yet |  |
| legacy-activity-diagram/fork-and-join-1 | not supported yet |  |
| legacy-activity-diagram/note-one-line-1 | not supported yet |  |
| legacy-activity-diagram/note-several-lines-1 | not supported yet |  |
| legacy-activity-diagram/partition-block-1 | not supported yet |  |
| legacy-activity-diagram/start-activities-end-1 | not supported yet |  |
| legacy-activity-diagram/transition-styles-1 | not supported yet |  |
| legacy-activity-diagram/transitions-from-last-1 | not supported yet |  |
| preprocessor/stdlib-include-1 | not supported yet |  |
| state-diagram/choice-fork-join-1 | not supported yet |  |
| state-diagram/composite-own-start-end-1 | not supported yet |  |
| state-diagram/composite-own-start-end-2 | not supported yet |  |
| state-diagram/direction-style-label-1 | not supported yet |  |
| state-diagram/display-names-and-descriptions-1 | not supported yet |  |
| state-diagram/dotted-state-names-1 | not supported yet |  |
| state-diagram/dotted-state-names-2 | not supported yet |  |
| state-diagram/entry-exit-points-1 | not supported yet |  |
| state-diagram/hide-empty-description-1 | not supported yet |  |
| state-diagram/history-state-1 | not supported yet |  |
| state-diagram/history-style-1 | not supported yet |  |
| state-diagram/junction-1 | not supported yet |  |
| state-diagram/left-to-right-1 | not supported yet |  |
| state-diagram/map-element-1 | not supported yet |  |
| state-diagram/note-on-state-1 | not supported yet |  |
| state-diagram/reversed-transition-1 | not supported yet |  |
| state-diagram/side-by-side-regions-1 | not supported yet |  |
| state-diagram/side-by-side-regions-2 | not supported yet |  |
| state-diagram/side-by-side-regions-3 | not supported yet |  |
| state-diagram/stacked-regions-1 | not supported yet |  |
| state-diagram/stacked-regions-2 | not supported yet |  |
| state-diagram/stacked-regions-3 | not supported yet |  |
| state-diagram/state-colors-1 | not supported yet |  |
| state-diagram/states-and-transitions-1 | not supported yet |  |
| state-diagram/synchronization-bar-1 | not supported yet |  |
| vega-migration/ChenMovieAlias_Test-1 | not supported yet |  |
| vega-migration/ChenMovieExtended_Test-1 | not supported yet |  |
| vega-migration/ChenMovie_Test-1 | not supported yet |  |
| vega-migration/ChenRankDir_Test-1 | not supported yet |  |
| vega-migration/HideShow002_Test-1 | not supported yet |  |
| vega-migration/HideShow003_Test-1 | not supported yet |  |
| vega-nonreg/group2681-history_deep_class-1 | not supported yet |  |
| vega-nonreg/group2681-history_in_composite-1 | not supported yet |  |
| vega-nonreg/group2681-history_in_composite-2 | not supported yet |  |
| vega-nonreg/group2681-history_inline_color-1 | not supported yet |  |
| vega-nonreg/group2681-history_style_all_forms-1 | not supported yet |  |
| vega-nonreg/group2681-inline_stereotype_style-1 | not supported yet |  |
| vega-nonreg/group2681-inline_stereotype_style-2 | not supported yet |  |
| vega-nonreg/group2681-name_description_body-1 | not supported yet |  |
| vega-nonreg/group2727-dump_memory_shown_with_verbose-1 | not supported yet |  |
| vega-nonreg/group2727-dump_memory_silent_by_default-1 | not supported yet |  |
| vega-nonreg/group2727-log_basic-1 | not supported yet |  |
| vega-nonreg/group2727-log_still_clean_with_verbose-1 | not supported yet |  |
| vega-nonreg/group2727-log_with_variable-1 | not supported yet |  |
| vega-nonreg/group2783-composite_with_usymbol_unaffected-1 | not supported yet |  |
| vega-nonreg/group2783-composite_with_usymbol_unaffected-2 | not supported yet |  |
| vega-nonreg/group2783-hide_stereotype-1 | not supported yet |  |
| vega-nonreg/group2783-nested_styled_composite-1 | not supported yet |  |
| vega-nonreg/group2783-nested_styled_composite-2 | not supported yet |  |
| vega-nonreg/group2783-reported_case-1 | not supported yet |  |
| vega-nonreg/group2846-bug-1 | not supported yet |  |
| vega-nonreg/group2846-correct_code-1 | not supported yet |  |
| vega-nonreg/group2846-dollar_sign_code-1 | not supported yet |  |
| vega-nonreg/group2903-reported_case-1 | not supported yet |  |
| vega-nonreg/group2904-all_visibilities-1 | not supported yet |  |
| vega-nonreg/group2904-diamond_at_tail-1 | not supported yet |  |
| vega-nonreg/group2904-icon_size_zero-1 | not supported yet |  |
| vega-nonreg/group2904-left_to_right-1 | not supported yet |  |
| vega-nonreg/group2904-reported_case-1 | not supported yet |  |
| vega-nonreg/group2904-role_tail_only-1 | not supported yet |  |
| vega-nonreg/group2904-role_without_visibility-1 | not supported yet |  |
| vega-nonreg/group2917-smetana_hide_state_node_label-1 | not supported yet |  |
| vega-nonreg/group2917-smetana_hide_state_plain_edges-1 | not supported yet |  |
| vega-nonreg/group2936-dark_theme_stereotype-1 | not supported yet |  |
| vega-nonreg/group2936-light_theme_nested_package-1 | not supported yet |  |
| vega-nonreg/group2938-dot_separator-1 | not supported yet |  |
| vega-nonreg/group2938-double_backslash_separator-1 | not supported yet |  |
| vega-nonreg/group2938-intermediate_packages-1 | not supported yet |  |
| vega-nonreg/group2938-issue_example-1 | not supported yet |  |
| vega-nonreg/group2938-keyword_namespaceseparator-1 | not supported yet |  |
| vega-nonreg/group2938-namespace_block-1 | not supported yet |  |
| vega-nonreg/group2938-nested_links-1 | not supported yet |  |
| vega-nonreg/group2938-object_diagram-1 | not supported yet |  |
| vega-nonreg/group2938-stereotype_and_note-1 | not supported yet |  |
| vega-nonreg/simple-A0004-1 | not supported yet |  |
| vega-nonreg/simple-ComponentExtraArrows_0001-1 | not supported yet |  |
| vega-nonreg/simple-NamespaceSeparatorColon001-1 | not supported yet |  |
| vega-nonreg/simple-QualifiedAssoc001-1 | not supported yet |  |
| vega-nonreg/simple-QualifiedAssoc002-1 | not supported yet |  |
| vega-scxml/junction-pseudostate-1 | not supported yet |  |
| vega-scxml/state1-1 | not supported yet |  |
| vega-scxml/state2-1 | not supported yet |  |
| vega-scxml/state2-2 | not supported yet |  |
| vega-scxml/state3-1 | not supported yet |  |
| vega-scxml/state4-1 | not supported yet |  |
| vega-scxml/state4-2 | not supported yet |  |
| vega-scxml/state5-1 | not supported yet |  |
| vega-state/composite-autarkic-simple-1 | not supported yet |  |
| vega-state/composite-autarkic-simple-2 | not supported yet |  |
| vega-state/composite-non-autarkic-outgoing-only-1 | not supported yet |  |
| vega-state/composite-non-autarkic-outgoing-only-2 | not supported yet |  |
| vega-state/composite-non-autarkic-start-end-1 | not supported yet |  |
| vega-state/composite-non-autarkic-start-end-2 | not supported yet |  |
| vega-state/composite-siblings-with-transition-1 | not supported yet |  |
| vega-state/composite-siblings-with-transition-2 | not supported yet |  |
| vega-state/composite-siblings-with-transition-3 | not supported yet |  |
| vega-state/composite-two-levels-nested-1 | not supported yet |  |
| vega-state/composite-two-levels-nested-2 | not supported yet |  |
| vega-state/composite-two-levels-nested-3 | not supported yet |  |
| vega-state/composite-two-levels-non-autarkic-1 | not supported yet |  |
| vega-state/concurrent-composite-non-autarkic-sibling-link-1 | not supported yet |  |
| vega-state/concurrent-composite-non-autarkic-sibling-link-2 | not supported yet |  |
| vega-state/concurrent-composite-non-autarkic-sibling-link-3 | not supported yet |  |
| vega-state/concurrent-composite-non-autarkic-sibling-link-4 | not supported yet |  |
| vega-state/concurrent-composite-non-autarkic-start-end-1 | not supported yet |  |
| vega-state/concurrent-composite-non-autarkic-start-end-2 | not supported yet |  |
| vega-state/concurrent-composite-non-autarkic-start-end-3 | not supported yet |  |
| vega-state/concurrent-dash-three-regions-1 | not supported yet |  |
| vega-state/concurrent-dash-three-regions-2 | not supported yet |  |
| vega-state/concurrent-dash-three-regions-3 | not supported yet |  |
| vega-state/concurrent-dash-three-regions-4 | not supported yet |  |
| vega-state/concurrent-dash-two-regions-1 | not supported yet |  |
| vega-state/concurrent-dash-two-regions-2 | not supported yet |  |
| vega-state/concurrent-dash-two-regions-3 | not supported yet |  |
| vega-state/concurrent-empty-first-region-1 | not supported yet |  |
| vega-state/concurrent-empty-first-region-2 | not supported yet |  |
| vega-state/concurrent-empty-first-region-4 | not supported yet |  |
| vega-state/concurrent-nested-inside-composite-1 | not supported yet |  |
| vega-state/concurrent-nested-inside-composite-2 | not supported yet |  |
| vega-state/concurrent-nested-inside-composite-3 | not supported yet |  |
| vega-state/concurrent-nested-inside-composite-4 | not supported yet |  |
| vega-state/concurrent-pipe-three-regions-1 | not supported yet |  |
| vega-state/concurrent-pipe-three-regions-2 | not supported yet |  |
| vega-state/concurrent-pipe-three-regions-3 | not supported yet |  |
| vega-state/concurrent-pipe-three-regions-4 | not supported yet |  |
| vega-state/concurrent-pipe-two-regions-1 | not supported yet |  |
| vega-state/concurrent-pipe-two-regions-2 | not supported yet |  |
| vega-state/concurrent-pipe-two-regions-3 | not supported yet |  |
| vega-state/concurrent-pipe-with-nested-composite-region-1 | not supported yet |  |
| vega-state/concurrent-pipe-with-nested-composite-region-2 | not supported yet |  |
| vega-state/concurrent-pipe-with-nested-composite-region-3 | not supported yet |  |
| vega-state/concurrent-pipe-with-nested-composite-region-4 | not supported yet |  |
| vega-state/concurrent-region-nested-composite-1 | not supported yet |  |
| vega-state/concurrent-region-nested-composite-2 | not supported yet |  |
| vega-state/concurrent-region-nested-composite-3 | not supported yet |  |
| vega-state/concurrent-region-nested-composite-4 | not supported yet |  |
| vega-state/concurrent-top-level-root-1 | not supported yet |  |
| vega-state/concurrent-top-level-root-2 | not supported yet |  |
| vega-state/concurrent-with-internal-transitions-1 | not supported yet |  |
| vega-state/concurrent-with-internal-transitions-2 | not supported yet |  |
| vega-state/concurrent-with-internal-transitions-3 | not supported yet |  |
| vega-state/junction-pseudostate-1 | not supported yet |  |
| vega-state/node-style-transition-label-1 | not supported yet |  |
| vega-state/node-style-transition-label-orphaned-by-hide-1 | not supported yet |  |
| vega-state/node-style-transition-label-orphaned-by-remove-1 | not supported yet |  |
| vega-state/node-style-transition-label-orphaned-by-remove-both-sides-1 | not supported yet |  |
| vega-state/node-style-transition-label-orphaned-by-remove-source-1 | not supported yet |  |
| vega-state/node-style-transition-label-remove-then-restore-1 | not supported yet |  |
| vega-state/simple-states-1 | not supported yet |  |
| vega-state/simple-states-2 | not supported yet |  |
| vega-svg/id-component_01-1 | not supported yet |  |
| vega-svg/id-state_01-1 | not supported yet |  |
| vega-svg/id-state_02-1 | not supported yet |  |
| vega-svg/id-state_02-2 | not supported yet |  |
| vega-svg/id-state_03-1 | not supported yet |  |
| vega-svg/id-state_04-1 | not supported yet |  |
| vega-svg/id-state_04-2 | not supported yet |  |
| vega-svg/id-state_05-1 | not supported yet |  |
| vega-svg/id-state_05-2 | not supported yet |  |
| vega-svg/id-state_05-3 | not supported yet |  |
| vega-svg/interactive-SVG0004_Svek-1 | not supported yet |  |
| vega-svg/interactive-SVG0005_Svek-1 | not supported yet |  |
| vega-svg/interactive-SVG0006_Svek-1 | not supported yet |  |
| cli/later-format | not supported yet |  |
| cli/other-engine | not supported yet |  |
| cli/default-attribute-flag | not supported yet |  |
