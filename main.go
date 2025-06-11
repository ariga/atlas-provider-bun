package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/types"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/alecthomas/kong"
	"golang.org/x/tools/go/packages"
)

var (
	//go:embed loader.tmpl
	loader     string
	loaderTmpl = template.Must(template.New("loader").Parse(loader))
)

func main() {
	var cli struct {
		Load LoadCmd `cmd:""`
	}
	ctx := kong.Parse(&cli)
	if err := ctx.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err) // nolint: errcheck
		os.Exit(1)
	}
}

// LoadCmd is the command to load models
type LoadCmd struct {
	Path    string   `help:"Path to the model files" required:""`
	Models  []string `help:"Models to load"`
	Dialect string   `help:"dialect to use" enum:"mysql,sqlite,postgres,mssql,oracle" required:""`
	out     io.Writer
}

func (c *LoadCmd) Run() error {
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedModule | packages.NeedDeps}
	pkgs, err := packages.Load(cfg, c.Path)
	if err != nil {
		return err
	}
	models := gatherModels(pkgs)
	p := Payload{
		Models:  models,
		Dialect: c.Dialect,
	}
	var buf bytes.Buffer
	if err := loaderTmpl.Execute(&buf, p); err != nil {
		return err
	}
	source, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	s, err := runprog(source)
	if err != nil {
		return err
	}
	if c.out == nil {
		c.out = os.Stdout
	}
	_, err = fmt.Fprintln(c.out, s)
	return err
}

func runprog(src []byte) (string, error) {
	if err := os.MkdirAll(".bunschema", os.ModePerm); err != nil {
		return "", err
	}
	target := fmt.Sprintf(".bunschema/%s.go", filename("bun"))
	if err := os.WriteFile(target, src, 0644); err != nil {
		return "", fmt.Errorf("bunschema: write file %s: %w", target, err)
	}
	defer os.RemoveAll(".bunschema")
	return gorun(target)
}

// run 'go run' command and return its output.
func gorun(target string) (string, error) {
	s, err := gocmd("run", target)
	if err != nil {
		return "", fmt.Errorf("bunschema: %s", err)
	}
	return s, nil
}

// goCmd runs a go command and returns its output.
func gocmd(command, target string) (string, error) {
	args := []string{command}
	args = append(args, target)
	cmd := exec.Command("go", args...)
	stderr := bytes.NewBuffer(nil)
	stdout := bytes.NewBuffer(nil)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		return "", errors.New(stderr.String())
	}
	return stdout.String(), nil
}

func filename(pkg string) string {
	name := strings.ReplaceAll(pkg, "/", "_")
	return fmt.Sprintf("atlasloader_%s_%d", name, time.Now().Unix())
}

type Payload struct {
	Models  []model
	Dialect string
}

func (p Payload) Imports() []string {
	imports := make(map[string]struct{})
	for _, m := range p.Models {
		imports[m.ImportPath] = struct{}{}
	}
	var result []string
	for k := range imports {
		result = append(result, k)
	}
	return result
}

type model struct {
	ImportPath string
	PkgName    string
	Name       string
}

func gatherModels(pkgs []*packages.Package) []model {
	var models []model
	for _, pkg := range pkgs {
		for k, v := range pkg.TypesInfo.Defs {
			_, ok := v.(*types.TypeName)
			if !ok || !k.IsExported() {
				continue
			}
			if isBunModel(k.Obj.Decl) {
				models = append(models, model{
					ImportPath: pkg.PkgPath,
					Name:       k.Name,
					PkgName:    pkg.Name,
				})
			}
		}
	}
	// Return models in deterministic order.
	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})
	return models
}

func isBunModel(decl any) bool {
	spec, ok := decl.(*ast.TypeSpec)
	if !ok {
		return false
	}
	// Any struct can be a Bun model
	_, ok = spec.Type.(*ast.StructType)
	return ok
}
