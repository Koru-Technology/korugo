package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

// TODO: convert to embedded static .yml files
// When Go 1.16 supports it
var yml_gqlgen = `
schema:
  - ../gql/schema/query.gql
  - ../gql/schema/mutation.gql

exec:
  filename: graph/server/server.go
  package: gqlserver

model:
  filename: graph/model/model.go
  package: gqlmodel

resolver:
  layout: follow-schema
  dir: ../gql/resolvers
  package: gql

autobind:
  - "%s/internal/korugo/private/db"

omit_slice_element_pointers: true
`
var yml_sqlc = `
version: "1"
packages:
  - name: "gendb"
    path: "./db"
    queries: "../db/queries/"
    schema: "../db/migrations/"
    engine: "postgresql"
    emit_json_tags: true
    emit_prepared_queries: true
    emit_interface: false
    emit_exact_table_names: false
overrides:
  - db_type: "uuid"
    go_type: "github.com/gofrs/uuid.UUID"
`

type Config struct {
	Generate map[string]interface{}
}

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

					configFile, err := os.Open("./korugo.yml")
					if err != nil && err != os.ErrNotExist {
						log.Fatalf("Error opening config: %v", err)
					}

					if configFile != nil {
						configContent, err := ioutil.ReadAll(configFile)
						if err != nil {
							log.Fatal(err)
						}
						result := &Config{}
						yaml.Unmarshal(configContent, result)
						if val, ok := result.Generate["gql"]; ok {
							customGqlConfig, err := yaml.Marshal(val)
							if err != nil {
								log.Fatal(err)
							}
							yml_gqlgen = fmt.Sprintf("%s\n%s\n", yml_gqlgen, string(customGqlConfig))
						}
					}

					dir, _ := os.Getwd()
					dir = fmt.Sprintf("%s/internal/korugo", dir)
					os.MkdirAll(dir, os.ModePerm)
					os.MkdirAll(fmt.Sprintf("%s/private", dir), os.ModePerm)
					os.MkdirAll(fmt.Sprintf("%s/db", dir), os.ModePerm)
					os.MkdirAll(fmt.Sprintf("%s/gql", dir), os.ModePerm)
					os.MkdirAll(fmt.Sprintf("%s/dataloaders", dir), os.ModePerm)

					// Prepare sql models via sqlc
					f, err := os.Create(fmt.Sprintf("%s/private/sqlc.yaml", dir))
					f.Write([]byte(yml_sqlc))
					f.Close()
					var b bytes.Buffer
					var berr bytes.Buffer
					cmdout := bufio.NewWriter(&b)
					cmderr := bufio.NewWriter(&berr)
					cmd := exec.Command("sqlc", "generate")
					cmd.Stdout = cmdout
					cmd.Stderr = cmderr
					cmd.Dir = fmt.Sprintf("%s/private", dir)
					cmd.Run()
					read, err := ioutil.ReadAll(&b)
					fmt.Println(string(read))
					read, err = ioutil.ReadAll(&berr)
					fmt.Println(string(read))

					// Prepare gql resolvers via gqlgen
					file, err := os.Open("./go.mod")
					if err != nil {
						log.Fatal(err)
					}
					defer file.Close()

					scanner := bufio.NewScanner(file)
					scanner.Scan()
					modline := scanner.Text()
					if err := scanner.Err(); err != nil {
						log.Fatal(err)
					}
					if modline == "" {
						log.Fatal("Could not find module")
					}
					parts := strings.Split(modline, " ")
					if len(parts) != 2 {
						log.Fatal("Could not find module")
					}
					modname := parts[1]

					f, err = os.Create(fmt.Sprintf("%s/private/gqlgen.yml", dir))
					f.Write([]byte(fmt.Sprintf(yml_gqlgen, modname)))
					f.Close()

					cmd = exec.Command("gqlgen")
					cmd.Stdout = cmdout
					cmd.Stderr = cmderr
					cmd.Dir = fmt.Sprintf("%s/private", dir)
					cmd.Run()
					read, err = ioutil.ReadAll(&b)
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
