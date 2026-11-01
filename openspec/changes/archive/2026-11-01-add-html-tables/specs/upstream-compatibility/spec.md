## MODIFIED Requirements

### Requirement: Case outcomes and report

Each case SHALL have one outcome:

- passed
- known difference: listed in a known-differences file with a reason and the follow-up change that resolves it
- not supported yet: the run of the case emitted a "not supported yet" warning or error
- failed: any other case that does not pass

The conformance run SHALL fail when a case is failed, and SHALL NOT fail for known differences and cases not supported yet. It SHALL write a report with the number of cases of each outcome per capability and per format.

#### Scenario: Feature of a later slice

- **WHEN** a case uses a compass point on a node ID, as in `a:n -> b`, before compass points are supported
- **THEN** the case is reported as not supported yet, and the run does not fail because of it

#### Scenario: Unlisted failure

- **WHEN** a case fails and is not listed in the known-differences file
- **THEN** the conformance run fails and the report names the case
