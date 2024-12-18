# Sarfya Service

This was created to break away all the logic from the library to make it easier to integrate into fwew.

## License

The GPL license is no longer in effect since the GPL dependency is in the other package.

The text used in `data/` is the property of their original authors,
and are attributed under the `source` property of the YAML documents.

The remaining code and the annotations in `data/` falls under the ISC license (LICENSE-ISC.txt).


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
