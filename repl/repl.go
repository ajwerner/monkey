// Package repl implements a simple repl.
package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/ajwerner/monkey/evaluator"
	"github.com/ajwerner/monkey/lexer"
	"github.com/ajwerner/monkey/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		evaluated := evaluator.Eval(program)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func printParserErrors(out io.Writer, errors []error) {
	for _, err := range errors {
		io.WriteString(out, "\t"+err.Error()+"\n")
	}
}
