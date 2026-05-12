package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	io := commands.NewDefaultIO()
	cmd := newRootCmd(io)
	cmd.Execute(context.Background(), os.Args[1:])
}
