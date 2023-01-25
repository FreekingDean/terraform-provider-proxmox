package main

import (
	"fmt"
	"io"
	"strings"
)

type Generator struct {
	o      io.Writer
	indent int
}

func (g Generator) in(i int) Generator {
	return Generator{
		o:      g.o,
		indent: g.indent + i,
	}
}

func (g Generator) p(m string, v ...any) {
	indent := strings.Repeat("	", g.indent)
	fmt.Fprintf(g.o, indent+m+"\n", v...)
}
