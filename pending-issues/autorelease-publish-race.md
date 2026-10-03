# The autorelease step fails a build when two runs publish at once

**Where the fix belongs:** `wow-look-at-my/go-toolchain`, the autorelease step. Not this repo -- this repo only observes it.

## What happened

Two commits were pushed seconds apart. Both CI runs went green through build, tests and both dats suites, then both reached the publish step and resolved the same next version. One published it. The other died:

```
Found 6 per-platform artifact(s) and 0 portable artifact(s) across 1 project(s): xml-validator
##[error]Release exists for xml-validator, no explicit version
```

- failing run: 31902939879, head `5e0d07f`
- run that won the race: 31902959805, head `4db0582`, published the version seconds earlier and reported success

The failing run's own work was entirely correct. The red check says nothing about the commit it is attached to. That is the whole problem. A red build that does not mean the code is broken teaches everyone to skim past red builds.

## Why it is a real defect and not bad luck

Any repo that takes two pushes in quick succession hits it. Quick pushes are normal: a fix and its follow-up, a rebase, a batch of review edits. The version is resolved by reading what exists and adding one. Nothing holds between that read and the write.

## Candidate fixes, best first

1. **Resolve and publish atomically, server-side.** The publisher asks for "the next version" in one operation. A colliding request either receives the version after it, or, when the artifact digest matches what is already published, is a no-op that succeeds.
2. **Derive an explicit version from the commit SHA.** Two runs then never contend, because they are never publishing the same version. Costs the monotonic numbering.
3. **A concurrency group on the publish step.** Serializes the two runs, which hides the race rather than removing it. A genuine simultaneous arrival still ties.

One rule holds whichever lands. **A publish that loses the race must not fail the build when its artifact is byte-identical to the published one.** Nothing is wrong in that case. Nothing needs a human.

## How to reproduce

Push two commits to the same branch about ten seconds apart, so the second run's publish step starts while the first one's is in flight.
