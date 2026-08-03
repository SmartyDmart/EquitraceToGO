# TypescriptToGolang

A small Go port of a CSV parser package with a minimal runnable entry point.

## Run the parser

From the project root:

```bash
go run . "a,b,c\n1,2,3"
```

Expected output:

```text
["a" "b" "c"]
["1" "2" "3"]
```

## Run tests

```bash
go test ./...
```

## Package use

The parser logic lives in the package:

```go
parser := csv.NewParser(nil)
rows, err := parser.Parse("a,b,c\n1,2,3")
```
