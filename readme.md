# Sarfya Service

This was created to break away all the logic from the library to make it easier to integrate into fwew.

## License

This part of Sarfya, including the annotations to the data are licensed under the GPL license.

The text used in `data/` is the property of their original authors,
who are attributed under the `source` property of the YAML documents.

## Project structure.

This package contains all the parts needed to rig this up to run with fwew.

For a minimal use of the library, see `./cmd/sarfya-example`

### Adapters

The adapters are exchangable pieces of the logic.

#### `fwewdictionary`

This hooks the `sarfya.Dictionary` interface up with `fwew`.

#### `webapi`

An (aspirationally) REST API for the frontend.

#### `templfrontend`

A WIP templ-based frontend that's faster to host.
It's a bit coupled with the `webapi` package, however.

#### `sourcestorage`

The storage backend for `service` ran locally and for the compilation step.
It modifies the relevant files in `./data`.
