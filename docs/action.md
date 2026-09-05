# The GitHub Action

`action.yml` is a composite action. It obtains an `xml-validator` binary, then
runs it once per file. Consumers use it as
`wow-look-at-my/xml-validator@master`; this repository has no tags.

## Getting the binary

1. **buildhost.** `buildhost-download` resolves the release for the runner's own
   os/arch. `required: false` turns a missing artifact into an output rather
   than a failed step, which is what leaves the build below reachable. The
   destination is named because Windows wants the `.exe` suffix to run a file.
2. **The `--help` check.** A download proves bytes arrived, never that they run
   here: a fat APE exits 121 before `main` on a runner with no loader for it.
   One `--help` separates the two.
3. **The source build.** Only a failure above reaches it, and it warns as it
   goes, so a broken publish shows up in the log rather than as a slower green
   run. It sets up Go from `cli/go.mod`, caches the binary on a hash of the Go
   sources, and builds `cli/cmd/xml-validator`. The binary is the `cli` module
   and this repository has no root `go.mod`, so both steps name `cli/`
   explicitly.

## Choosing files

`files` takes names and glob patterns. With no `files`, discovery walks the
workspace, skipping `.git` and `node_modules`, and takes every `*.xml` and
`*.xsd`. An XSD is an XML document: a schema this tool cannot parse is a
finding, not a file to walk past.

`schema` adds `--schema` to every invocation. `args` is passed through.

## Driving the binary yourself

`install-only: 'true'` puts xml-validator on `PATH` and validates nothing.

It exists for a suite that asserts a document is REJECTED. The validate mode
fails the job on exactly those documents. Such a suite runs the binary itself.
`api-cli-spec` is the consumer.

The `path` output names the binary this run resolved. It is set either way.

## What CI proves

The `action` job in `.github/workflows/ci.yml` runs the action from the
checkout, so a change is executed here before any consumer sees it. It covers a
valid document, a document validated against a schema, and an XML 1.0 document
that must be rejected. That last one is a negative control: without it the
other cases also pass for an action that reports success whatever the validator
answered. Fixtures live in `testdata/action/`.

## Writing comments in this file

`yaml-comment-block` fails the build on two adjacent `#` lines in a workflow or
an action, and the TypeScript action fails on two adjacent `//` lines in a
script. One line, or this doc.
