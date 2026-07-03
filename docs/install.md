# Installation guide

clip-history is designed as a portable CLI application.

## Recommended install

- Download the latest portable binary from the GitHub releases page.
- Place it in a directory that is already on your PATH.
- Verify the install with:

```bash
ch --help
```

## Manual build

Build from source with:

```bash
go build ./cmd/ch
```

To produce a Windows portable executable:

```bash
GOOS=windows GOARCH=amd64 go build -o dist/ch.exe ./cmd/ch
```
