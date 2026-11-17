This is the Go implementation of FLOE (Fast Lightweight Online Encryption) from
<https://github.com/Snowflake-Labs/floe-specification/tree/main/go>.
It's vendored here because the upstream `snowflake.com/floe/go` module is
unresolvable.

The following modifications have been made:
* Disabled tests that require data files.
* Fixed error string casing.
