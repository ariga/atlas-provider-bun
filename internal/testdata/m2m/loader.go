package main

import (
	"fmt"
	"log"
	"os"

	"ariga.io/atlas-provider-bun/bunschema"
	"ariga.io/atlas-provider-bun/internal/testdata/m2m/models"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run . <dialect>")
	}
	stmt, err := bunschema.New(bunschema.Dialect(os.Args[1])).Load(
		&models.OrderToItem{},
		&models.Item{},
		&models.Order{},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(stmt)
}
