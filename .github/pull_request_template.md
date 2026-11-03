## Summary

<!-- What changes and why. Link the issue, and the OpenSpec change when the specs change. -->

## Clean-room

See the clean-room rules in [CONTRIBUTING.md](https://github.com/ekalinin/gogv/blob/main/CONTRIBUTING.md#clean-room-rules).

- [ ] This change was written without reading or copying the source code of Graphviz or of its ports and bindings (such as Smetana in PlantUML, graphviz-pure, and goccy/go-graphviz and its forks), by me or by an AI assistant I used.
- [ ] Every new or changed Go file names its sources in `Provenance:` lines.

## Checks

- [ ] `make lint test provenance` passes.
- [ ] For a change of the behavior that `openspec/specs` describes: the OpenSpec change is part of this pull request.
