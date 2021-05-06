package rego

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fugue/regula/pkg/version"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/repl"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/storage/inmem"
)

type RunREPLOptions struct {
	Ctx      context.Context
	UserOnly bool
	Includes []string
}

func RunREPL(options *RunREPLOptions) error {
	RegisterBuiltins()
	store, err := initStore(options.Ctx, options.UserOnly, options.Includes)
	if err != nil {
		return err
	}
	r := repl.New(
		store,
		"./.regula-history",
		os.Stdout,
		"pretty",
		ast.CompileErrorLimitDefault,
		getBanner())
	r.Loop(options.Ctx)
	return nil
}

func initStore(ctx context.Context, userOnly bool, includes []string) (storage.Store, error) {
	store := inmem.New()
	txn, err := store.NewTransaction(ctx, storage.TransactionParams{
		Write: true,
	})
	if err != nil {
		return nil, err
	}
	cb := func(r RegoFile) error {
		return store.UpsertPolicy(ctx, txn, r.Path(), r.Raw())
	}
	if err := LoadRegula(userOnly, cb); err != nil {
		return nil, err
	}
	if err := LoadOSFiles(includes, cb); err != nil {
		return nil, err
	}
	if err := store.Commit(ctx, txn); err != nil {
		return nil, err
	}
	return store, nil
}

func getBanner() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Regula v%v - built on OPA v%v\n", version.Version, version.OPAVersion))
	sb.WriteString("Run 'help' to see a list of commands.")
	return sb.String()
}
