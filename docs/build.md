# Build guide

## Local build

Use the provided script:

```bash
bash scripts/build-local.sh
```

This produces a Unix binary in dist/ch and a Windows portable executable in dist/ch.exe.

## Docker build

```bash
docker build -t clip-history .
```
