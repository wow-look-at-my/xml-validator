# Nothing checks that a module's require of a sibling module resolves

**Where the fix belongs:** this repo. The version bump that unbroke consumers
has landed; the reason nobody saw the break has not.

## What happened

`validator/go.mod`, `cli/go.mod` and `writer/go.mod` each required
`github.com/wow-look-at-my/xml-validator/reader` at
`v0.0.0-20260815191334-3096d3b2da43`. That commit is real, and it is older than
`9e50633`, the commit that added `reader/go.mod`. So the version names a point
in history where the reader module did not exist yet.

Every build inside this repo passed anyway, because each of those three
`go.mod` files also carries:

```
replace github.com/wow-look-at-my/xml-validator/reader => ../reader
```

A `replace` applies only in the main module. Inside this repo the sibling
directory answers and the require's version is never read. Outside it, the
version is the only thing there is, and Go stops with:

```
github.com/wow-look-at-my/xml-validator/reader@v0.0.0-20260815191334-3096d3b2da43:
invalid version: missing github.com/wow-look-at-my/xml-validator/reader/go.mod
at revision 3096d3b2da43
```

`wow-look-at-my/tml` hit exactly that and could not build at all: its own
`go mod tidy` failed on a dependency two levels down. The bad version had been
sitting in three files, through green CI on every one of them.

## Why the checks that exist cannot catch it

CI runs `go-toolchain` once per module, in that module's directory. That is the
one place the `replace` wins, so the require is dead text there. There is no job
that consumes a module of this repo the way another repository does, and that is
the only vantage point the defect is visible from.

## Two ways to fix it, and what each costs

1. **Delete the `replace` directives and add a `go.work` for local
   development.** The requires become load-bearing everywhere, so a version that
   does not resolve fails the module's own build immediately. This is the fix
   that removes the class rather than detecting it. It needs a decision about
   whether `go-toolchain` runs correctly under a workspace, which is why it is
   written here rather than applied.

2. **Add a job that builds this repo's modules as an outside consumer would**:
   a scratch module requiring `.../validator` at the pushed commit, with no
   `replace`, built from the proxy. It catches the same defect one push later
   than option 1, and it needs the commit to be published before it can run.

Either one turns "the requires are decorative" into something a build can say
out loud. Until one lands, a sibling require is only as correct as whoever last
typed it.
