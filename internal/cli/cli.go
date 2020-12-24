package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

// TODO: convert to embedded static .yml files
// When Go 1.16 supports it
var yml_gqlgen = `
schema:
  - ../../gql/schema/query.gql
  - ../../gql/schema/mutation.gql

exec:
  filename: private/graph/server/server.go
  package: gqlserver

model:
  filename: private/graph/model/model.go
  package: gqlmodel

resolver:
  layout: follow-schema
  dir: resolvers
  package: gql

autobind:
  - "%s/internal/korugo/db"

omit_slice_element_pointers: true
`
var yml_sqlc = `
version: "1"
packages:
  - name: "gendb"
    path: "."
    queries: "./db/queries/"
    schema: "./db/migrations/"
    engine: "postgresql"
    emit_json_tags: true
    emit_prepared_queries: true
    emit_interface: false
    emit_exact_table_names: false
overrides:
  - db_type: "uuid"
    go_type: "github.com/gofrs/uuid.UUID"
`

func main() {
	app := &cli.App{
		Name:  "korugo",
		Usage: "the source of your graphql api",
		Commands: []*cli.Command{
			{
				Name:    "generate",
				Aliases: []string{"gen"},
				Usage:   "(Re)generate your graphql resolvers",
				Action: func(c *cli.Context) error {
					dir, _ := os.Getwd()
					dir = fmt.Sprintf("%s/internal/korugo", dir)
					os.MkdirAll(dir, os.ModePerm)

					f, err := os.Create(fmt.Sprintf("%s/sqlc.yaml", dir))
					f.Write([]byte(yml_sqlc))
					f.Close()

					f, err = os.Create(fmt.Sprintf("%s/gqlgen.yml", dir))
					f.Write([]byte(yml_gqlgen))
					f.Close()

					var b bytes.Buffer
					var berr bytes.Buffer
					cmdout := bufio.NewWriter(&b)
					cmderr := bufio.NewWriter(&berr)

					cmd := exec.Command("sqlc", "generate")
					cmd.Stdout = cmdout
					cmd.Stderr = cmderr
					cmd.Dir = dir
					cmd.Run()
					read, err := ioutil.ReadAll(&b)
					fmt.Println(string(read))
					read, err = ioutil.ReadAll(&berr)
					fmt.Println(string(read))
					return err
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
