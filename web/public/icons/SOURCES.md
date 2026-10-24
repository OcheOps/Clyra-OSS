# Vendored architecture icons

The AWS and Google Cloud SVGs in this directory were fetched from the
IcePanel architecture-icon mirror by `scripts/vendor-icons.mjs`. The small
set under `brands/` comes from the Iconify open-source logos collection. The
manifest records canonical component IDs for the provider architecture icons.

- AWS product icons and trademarks remain the property of Amazon Web Services.
- Google Cloud product icons and trademarks remain the property of Google.
- Other brand marks retain the licenses and trademarks documented by the
  corresponding Iconify collection.

Run `node scripts/vendor-icons.mjs aws gcp brands` to refresh the pinned repository
copy, then review asset changes before committing them.
