# go-maml

[MAML](https://maml.dev) data format implementation for Go.

- Spec-accurate v0.1 parser
- Zero dependencies
- Full AST with comments and source positions
- Struct marshaling/unmarshaling with `maml` tags

## Installation

```bash
go get github.com/maml-dev/go-maml
```

## Usage

### Marshal / Unmarshal

```go
package main

import (
	"fmt"

	"github.com/maml-dev/go-maml"
)

type Config struct {
	Name  string   `maml:"name"`
	Port  int      `maml:"port"`
	Debug bool     `maml:"debug,omitempty"`
	Tags  []string `maml:"tags"`
}

func main() {
	// Unmarshal
	input := []byte(`{
  name: "my-app"
  port: 8080
  tags: ["web", "api"]
}`)

	var config Config
	if err := maml.Unmarshal(input, &config); err != nil {
		panic(err)
	}
	fmt.Println(config.Name) // my-app

	// Marshal
	data, err := maml.Marshal(config)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
```

### Parse to Value

```go
val, err := maml.Parse(`{name: "hello", items: [1, 2, 3]}`)
if err != nil {
    panic(err)
}

name, _ := val.Get("name")
fmt.Println(name.AsString()) // hello
```

### AST with Comments

```go
package main

import (
	"fmt"

	"github.com/maml-dev/go-maml/ast"
)

func main() {
	doc, err := ast.Parse(`
# Server config
{
  host: "localhost"
  port: 8080 # default port
}
`)
	if err != nil {
		panic(err)
	}

	// Comments are preserved
	fmt.Println(len(doc.LeadingComments)) // 1

	// Print back with comments
	fmt.Println(ast.Print(doc))
}
```

## MAML Format

```maml
{
  # Comments start with hash
  name: "MAML"
  version: 1
  score: 9.5
  enabled: true
  nothing: null

  tags: [
    "minimal"
    "readable"
  ]

  nested: {
    key: "value"
  }

  poem: """
Roses are red,
Violets are blue.
"""
}
```

## License

[MIT](LICENSE)
