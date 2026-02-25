package generator

import (
	"bytes"
	"embed"
	"strings"
	"text/template"

	"github.com/Zubimendi/erc20gen/internal/config"
)

//go:embed templates/*
var templatesFS embed.FS

// Generator holds config and renders templates.
type Generator struct {
	cfg *config.TokenConfig
}

// New creates a new Generator.
func New(cfg *config.TokenConfig) *Generator {
	return &Generator{cfg: cfg}
}

// GenerateContract renders the Solidity ERC-20 contract.
func (g *Generator) GenerateContract() (string, error) {
	tmpl, err := template.New("contract.sol.tmpl").Funcs(templateFuncs()).ParseFS(templatesFS, "templates/contract.sol.tmpl")
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "contract.sol.tmpl", g.cfg); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateDeployScript renders a Hardhat deploy script (JS).
func (g *Generator) GenerateDeployScript() (string, error) {
	tmpl, err := template.New("deploy.js.tmpl").Funcs(templateFuncs()).ParseFS(templatesFS, "templates/deploy.js.tmpl")
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "deploy.js.tmpl", g.cfg); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateTestSkeleton renders a Hardhat test skeleton (JS).
func (g *Generator) GenerateTestSkeleton() (string, error) {
	tmpl, err := template.New("test.js.tmpl").Funcs(templateFuncs()).ParseFS(templatesFS, "templates/test.js.tmpl")
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "test.js.tmpl", g.cfg); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"join":  strings.Join,
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"quote": func(s string) string { return "\"" + s + "\"" },
		"add":   func(a, b int) int { return a + b },
	}
}