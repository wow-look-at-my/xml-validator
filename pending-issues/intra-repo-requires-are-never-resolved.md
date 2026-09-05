# No job builds a module of this repo the way an outside consumer does

**Where the fix belongs:** this repo.

## The break this came from, and why it cannot recur

`validator/go.mod`, `cli/go.mod` and `writer/go.mod` each required `github.com/wow-look-at-my/xml-validator/reader` at a pseudo-version older than `9e50633`, the commit that added `reader/go.mod`. The version named a point in history where the reader module did not exist.

Every build in this repo passed, because each file also carries `replace github.com/wow-look-at-my/xml-validator/reader => ../reader`. A replace applies only in the main module. Here the sibling directory answers and the require's version is dead text. Outside, the version is all there is. `wow-look-at-my/tml` stopped with `missing go.mod at revision 3096d3b2da43`.

Those requires now carry `// go-toolchain:auto-branch`, so go-toolchain re-resolves each one to the default branch's head on every run. The pin cannot go stale again. Any replace used to exempt a require from that marker, which is what let the pins sit still. That is fixed at the source.

## What is still missing

No job here consumes a module of this repo from outside. CI runs `go-toolchain` once per module, in the directory where the replace wins. A require is never load-bearing in this repository at all. The staleness that produced the outage is gone. The vantage point that shows such staleness still does not exist.

Two ways to build it:

1. **Delete the replace directives and add a `go.work` for local development.** The requires become load-bearing everywhere, so a version that does not resolve fails the module's own build. This removes the class rather than detecting it. It needs a decision about whether `go-toolchain` runs correctly under a workspace. That is why it is written here rather than applied.
2. **Add a job that builds these modules the way an outside consumer does.** That is a scratch module requiring `.../validator` at the pushed commit, with no replace, built from the proxy. It catches a defect one push later than option 1, and the commit has to be published before it can run.
