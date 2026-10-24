# CDK Roadmap

CDK TypeScript emission is deferred to v2.

v1 of the IaC quality uplift focuses on:

- AWS-only structured `ResourceGraph` generation
- deterministic Terraform rendering
- deterministic CloudFormation rendering
- five canonical stack fixtures with CI validation

Interim plan:

- expose `cdk-context.json` as a stub export derived from the `ResourceGraph`
- publish a thin importer package (`@clyra/cdk-import`) in a later milestone
- keep the renderer contracts stable so CDK emission can consume the same graph

This keeps the current partnership/demo path credible without over-promising
TypeScript AST generation work in the first release.
