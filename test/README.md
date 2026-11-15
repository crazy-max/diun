# Integration tests

Run from the repository root with Docker Compose:

```console
docker compose -f test/compose.yaml run --build --rm tests
```

Clean up afterward:

```console
docker compose -f test/compose.yaml down --volumes --remove-orphans
```
