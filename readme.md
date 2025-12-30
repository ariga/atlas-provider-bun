# atlas-provider-bun

Use [Atlas](https://atlasgo.io/) with [Bun](https://bun.uptrace.dev/) to manage your database schema as code. By connecting your Bun models to Atlas,
you can define and edit your schema directly in Go, and Atlas will automatically plan and apply database schema migrations for you, eliminating
the need to write migrations manually.

Atlas brings automated CI/CD workflows to your database, along with built-in support for [testing](https://atlasgo.io/testing/schema), [linting](https://atlasgo.io/versioned/lint),
schema [drift detection](https://atlasgo.io/monitoring/drift-detection), and [schema monitoring](https://atlasgo.io/monitoring). It also allows you to extend Bun with advanced database
objects such as triggers, row-level security, and custom functions that are not supported natively.

### Use-cases
1. [**Declarative migrations**](https://atlasgo.io/declarative/apply) - Use the Terraform-like `atlas schema apply --env bun` command to apply your Bun schema to the database.
2. [**Automatic migration planning**](https://atlasgo.io/versioned/diff) - Use `atlas migrate diff --env bun` to automatically plan database schema changes and generate
   a migration from the current database version to the desired version defined by your Bun schema.


### Installation

Install Atlas from macOS or Linux by running:
```bash
curl -sSf https://atlasgo.sh | sh
```
See [atlasgo.io](https://atlasgo.io/getting-started#installation) for more installation options.

Install the provider by running:
```bash
go get -u ariga.io/atlas-provider-bun
``` 

#### Standalone 

If all of your Bun models exist in a single package, you can use the provider directly to load your Bun schema into Atlas.

In your project directory, create a new file named `atlas.hcl` with the following contents:

```hcl
data "external_schema" "bun" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-bun",
    "load",
    "--path", "./path/to/models",
    "--dialect", "mysql" // | postgres | sqlite | sqlserver
  ]
}

env "bun" {
  src = data.external_schema.bun.url
  dev = "docker://mysql/8/dev"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
```

##### Pinning Go dependencies

Next, to prevent the Go Modules system from dropping this dependency from our `go.mod` file, let's
follow its [official recommendation](https://go.dev/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module)
for tracking dependencies of tools and add a file named `tools.go` with the following contents:

```go title="tools.go"
//go:build tools
package main

import _ "ariga.io/atlas-provider-bun/bunschema"
```
Alternatively, you can simply add a blank import to the `models.go` file we created above.

Finally, to tidy things up, run:

```text
go mod tidy
```

#### As Go File

If you want to use the provider as a Go file, you can use the provider as follows:

Create a new program named `loader/main.go` with the following contents:

```go
package main

import (
  "io"
  "os"

  "ariga.io/atlas-provider-bun/bunschema"
  _ "ariga.io/atlas/sdk/recordriver"
  "github.com/<yourorg>/<yourrepo>/path/to/models"
)

func main() {
  stmts, err := bunschema.New(bunschema.DialectMySQL).Load(
    &models.User{},
    &models.Post{},
  )
  if err != nil {
    fmt.Fprintf(os.Stderr, "failed to load bun schema: %v\n", err)
    os.Exit(1)
  }
  io.WriteString(os.Stdout, stmts)
}
```

In your project directory, create a new file named `atlas.hcl` with the following contents:

```hcl
data "external_schema" "bun" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "./loader",
  ]
}

env "bun" {
  src = data.external_schema.bun.url
  dev = "docker://mysql/8/dev"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
```

### Usage

Once you have the provider installed, you can use it to apply your Bun schema to the database:

#### Apply

You can use the `atlas schema apply` command to plan and apply a migration of your database to
your current Bun schema. This works by inspecting the target database and comparing it to the
Bun schema and creating a migration plan. Atlas will prompt you to confirm the migration plan
before applying it to the database.

```bash
atlas schema apply --env bun -u "mysql://root:password@localhost:3306/mydb"
```
Where the `-u` flag accepts the [URL](https://atlasgo.io/concepts/url) to the
target database.

#### Diff

Atlas supports a [versioned migration](https://atlasgo.io/concepts/declarative-vs-versioned#versioned-migrations) 
workflow, where each change to the database is versioned and recorded in a migration file. You can use the
`atlas migrate diff` command to automatically generate a migration file that will migrate the database
from its latest revision to the current Bun schema.

```bash 
atlas migrate diff --env bun 
```

### Supported Databases

The provider supports the following databases:
* MySQL
* PostgreSQL
* SQLite
* SQL Server

### Frequently Asked Questions

- **Many-to-many relationships support** - The provider fully supports many-to-many relations in both [Standalone mode](#standalone) and [Script mode](#as-go-file). The provider automatically detects join tables based on struct tags and registers them in the correct order.

For example, given the following models:

```go
// Order has many Items via OrderToItem join table
type Order struct {
    ID    int64  `bun:",pk,autoincrement"`
    Items []Item `bun:"m2m:order_to_items,join:Order=Item"`
}

type Item struct {
    ID     int64   `bun:",pk,autoincrement"`
    Orders []Order `bun:"m2m:order_to_items,join:Item=Order"`
}

// OrderToItem is the join table (auto-detected via rel:belongs-to tags)
type OrderToItem struct {
    OrderID int64  `bun:",pk"`
    Order   *Order `bun:"rel:belongs-to,join:order_id=id"`
    ItemID  int64  `bun:",pk"`
    Item    *Item  `bun:"rel:belongs-to,join:item_id=id"`
}
```

**Standalone mode:**
```bash
atlas-provider-bun load --path ./path/to/models --dialect mysql
```

**Script mode** (see `internal/testdata/m2m/loader.go` for a complete program):
```go
stmts, err := bunschema.New(bunschema.DialectMySQL).Load(
    &models.Order{},
    &models.Item{},
    &models.OrderToItem{}, // join table - auto-detected
)
if err != nil {
    log.Fatal(err)
}
fmt.Println(stmts)
```

To understand how to declare the many-to-many relation on your Bun models, consult the official Bun documentation: <https://bun.uptrace.dev/guide/relations.html#many-to-many-relation>

### Issues

Please report any issues or feature requests in the [ariga/atlas](https://github.com/ariga/atlas/issues) repository.

### License

This project is licensed under the [Apache License 2.0](LICENSE).