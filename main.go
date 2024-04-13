package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/twolodzko/goer/core"
	"github.com/twolodzko/goer/core/pids"
	"github.com/twolodzko/goer/parser/reader"
)

func main() {
	if len(os.Args) == 1 {
		repl()
		return
	}

	env := core.NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	for _, arg := range os.Args[1:] {
		_, err := core.EvalFile(arg, env, pid)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func repl() {
	env := core.NewEnv()
	pid := pids.NewPid()
	defer pid.Close()

	reader := reader.NewReader(os.Stdin)

	fmt.Println("Press ^C to exit.")
	fmt.Println()

	for {
		fmt.Print("> ")
		code, err := reader.Next()
		if err != nil {
			printError(err)
		}
		expr, err := core.ParseEval(code, env, pid)
		if err != nil {
			printError(err)
			continue
		}
		print(expr)
	}
}

func printError(msg error) {
	print(fmt.Sprintf("ERROR: %s", msg))
}

func print(msg any) {
	_, err := io.WriteString(os.Stdout, fmt.Sprintf("%v\n", msg))
	if err != nil {
		log.Fatal(err)
	}
}
