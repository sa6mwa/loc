# count

`count` is a simple code line counter that counts

```
go build -trimpath -ldflags="-s -w" .
sudo install count /usr/local/bin/

# or...
go install github.com/sa6mwa/count@latest
```

### Usage

`count` takes variadic arguments of file suffixes, excluding *common* non-source
file directory (like `.git`, `node_modules`, `bin`, `out`, etc). Returns total
number of lines.

```
cd src
count .go .c
18899
```
