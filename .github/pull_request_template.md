<!-- Thanks for the PR. A few things that make review faster: -->

## What this does

<!-- One-line summary + 2-3 sentences on the why. If this fixes an issue,
say "Fixes #123". -->

## How to test

<!-- Steps a reviewer can follow to verify this. "Ran go test" is not
enough on its own — tell me what a user would do and what they'd see. -->

## Checklist

- [ ] `go test ./...` passes
- [ ] `cd web && npm test` passes
- [ ] `cd web && npm run build` succeeds
- [ ] `cd web && npm run lint` is clean
- [ ] `gofmt -s` is clean on any Go file I touched
- [ ] If I edited `components/*.md`, I regenerated the catalog
      (`npm run build:catalog`) and committed the output
- [ ] If this changes behavior, I updated the README / docs / `.env.example`

## Breaking changes

<!-- Anything that forces users to change env vars, migrate the DB, or
update their own code? Describe it. If "no", say "none". -->
