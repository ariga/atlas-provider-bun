variable "dialect" {
  type = string
}

locals {
  dev_url = {
    mysql = "docker://mysql/8/dev"
    postgres = "docker://postgres/15"
    mssql = "docker://sqlserver/2022-latest"
    sqlite = "sqlite://file::memory:?cache=shared"
    oracle = "docker://oracle/free:latest-lite"
  }[var.dialect]
}

data "external_schema" "bun" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-bun",
    "load",
    "--path", "./models",
    "--dialect", var.dialect,
  ]
}

env "bun" {
  src = data.external_schema.bun.url
  dev = local.dev_url
  migration {
    dir = "file://migrations/${var.dialect}"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
  diff {
    skip {
      # See this FAQ: https://atlasgo.io/faq/skip-constraint-rename
      rename_constraint = true
    }
  }
}