package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

const (
	COMMENT  = "#"
	RECSEP   = "%%"
	EMPTY    = "EMPTY"
	MACBEGIN = "${"
	MACPIPE  = "|"
	MACEND   = "}"
)

// expansion encodes a macro expansion.
type expansion struct {
	base string
	pipe string
}

func (e expansion) String() string {
	if e.pipe == "" {
		return e.base
	}
	return e.base + "|" + e.pipe
}

// production encodes a macro production.
// Each macro production is an interleaving of nonterminals (further macros)
// and terminals (raw strings).
// Each starts and finishes with a terminal (which may be empty).
type production struct {
	terminals    []string
	nonterminals []expansion
}

func (p production) String() string {
	if len(p.nonterminals)+1 != len(p.terminals) {
		return "(invalid production)"
	}

	werr := func(e error) string {
		return fmt.Sprintf("(error: %s", e.Error())
	}

	bbuf := bytes.NewBufferString(p.terminals[0])
	for i, nt := range p.nonterminals {
		var err error
		if _, err = bbuf.WriteString(MACBEGIN); err != nil {
			return werr(err)
		}
		if _, err = bbuf.WriteString(nt.String()); err != nil {
			return werr(err)
		}
		if _, err = bbuf.WriteString(MACEND); err != nil {
			return werr(err)
		}
		if _, err = bbuf.WriteString(p.terminals[i+1]); err != nil {
			return werr(err)
		}
	}

	return fmt.Sprintf("(%s)", bbuf.String())
}

var ntimes = flag.Int("n", 1, "number of times to expand the base macro")

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [-n int] [FILE]\n", os.Args[0])
	flag.PrintDefaults()
}

func printErr(e error) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], e.Error())
}

// getFile gets the file this program has been asked to open.
// If there were no file arguments, it opens stdin.
// If there was one file argument, it opens that file.
// Otherwise, it errors---use cat to open multiple files!
func getFile() (*os.File, error) {
	switch flag.NArg() {
	case 0:
		return os.Stdin, nil
	case 1:
		return os.Open(flag.Arg(0))
	default:
		return nil, fmt.Errorf("too many arguments: %d", flag.NArg())
	}
}

// parseCookieJar reads f to completion as a cookie-jar format file.
// (See http://www.catb.org/esr/writings/taoup/html/ch05s02.html)
func parseCookieJar(f *os.File) (records [][]string, err error) {
	var crecord []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ln := scanner.Text()
		lnNoComment := strings.SplitN(ln, COMMENT, 2)[0]
		if strings.HasPrefix(lnNoComment, RECSEP) {
			records = append(records, crecord)
			crecord = []string{}
		} else if lnNoComment != "" {
			crecord = append(crecord, lnNoComment)
		}
	}

	// Finish off last record
	records = append(records, crecord)

	err = scanner.Err()
	return
}

func parseExpansion(raw string) (e expansion) {
	psplit := strings.SplitN(raw, MACPIPE, 2)
	e.base = psplit[0]
	if len(psplit) == 2 {
		e.pipe = psplit[1]
	} else {
		e.pipe = ""
	}
	return
}

func parseProduction(raw string) (p production, err error) {
	rest := strings.TrimSpace(raw)

	for {
		bsplit := strings.SplitN(rest, MACBEGIN, 2)
		p.terminals = append(p.terminals, bsplit[0])
		if len(bsplit) == 1 {
			// No more macros to expand
			break
		}

		esplit := strings.SplitN(bsplit[1], MACEND, 2)
		if len(esplit) == 1 {
			err = fmt.Errorf("unclosed macro in line '%s'", raw)
			return
		}

		p.nonterminals = append(p.nonterminals, parseExpansion(esplit[0]))
		rest = esplit[1]
	}

	return
}

func cookieJarToMacros(records [][]string) (macros map[string][]production, lastmac string, err error) {
	macros = make(map[string][]production)

	for _, record := range records {
		// Ignore empty macros.
		// This is so we can start/end a macro list with a stray %%.
		if len(record) == 0 {
			continue
		}

		key := record[0]
		vals := []production{}
		for _, raw := range record[1:] {
			var val production
			if val, err = parseProduction(raw); err != nil {
				return
			}

			vals = append(vals, val)
		}
		macros[key] = vals
		lastmac = key
	}

	return
}

// getProduction gets a random production for the macro named m.
func getProduction(m string, macros map[string][]production) (p production, err error) {
	ps, ok := macros[m]
	if !ok {
		err = fmt.Errorf("no such macro: %s", m)
		return
	}

	p = ps[rand.Intn(len(ps))]
	return
}

// expandPipeOnce expands out the pipe in a piped macro expansion.
func expandPipe(e expansion, macros map[string][]production) (p production, err error) {
	if e.pipe == "" {
		p.nonterminals = []expansion{e}
		p.terminals = []string{}
		return
	}

	var fp production
	if fp, err = getProduction("|"+e.pipe, macros); err != nil {
		return
	}

	p.terminals = fp.terminals
	for _, fpe := range fp.nonterminals {
		var rbase string
		if fpe.base == "" {
			rbase = e.base
		} else {
			rbase = fpe.base
		}
		p.nonterminals = append(p.nonterminals, expansion{rbase, fpe.pipe})
	}

	return
}

// expandBaseOnce expands a non-piped macro invocation into a macro production.
func expandBaseOnce(s string, macros map[string][]production) (p production, err error) {
	if strings.HasPrefix(s, "|") {
		err = fmt.Errorf("tried to expand function macro: %s", s)
		return
	}

	return getProduction(s, macros)
}

// empty creates an empty expansion.
func empty() expansion {
	return expansion{EMPTY, ""}
}

// expandProdOnce does one round of expansion on a production.
// It will return a production with empty strings.
func expandProdOnce(p production, macros map[string][]production) (np production, err error) {
	if len(p.nonterminals)+1 != len(p.terminals) {
		err = fmt.Errorf("invalid production: %d nonterminals, %d terminals", p.nonterminals, p.terminals)
		return
	}

	np.terminals = []string{p.terminals[0]}

	for i, nt := range p.nonterminals {
		var ntp production
		if nt.pipe == "" {
			if ntp, err = expandBaseOnce(nt.base, macros); err != nil {
				return
			}
		} else if nt.base != EMPTY {
			if ntp, err = expandPipe(nt, macros); err != nil {
				return
			}
		}

		np.terminals = append(np.terminals, ntp.terminals...)
		// The empty()s help us glue together the terminals
		// on either side of the expansion.
		np.nonterminals = append(np.nonterminals, empty())
		np.nonterminals = append(np.nonterminals, ntp.nonterminals...)
		np.nonterminals = append(np.nonterminals, empty())

		np.terminals = append(np.terminals, p.terminals[i+1])
	}

	return
}

func removeEmpty(p production) (np production, err error) {
	if len(p.nonterminals)+1 != len(p.terminals) {
		err = fmt.Errorf("invalid production: %d nonterminals, %d terminals", p.nonterminals, p.terminals)
		return
	}

	np.terminals = []string{p.terminals[0]}

	for i, nt := range p.nonterminals {
		if nt.base == EMPTY {
			np.terminals[len(np.terminals)-1] += p.terminals[i+1]
		} else {
			np.nonterminals = append(np.nonterminals, nt)
			np.terminals = append(np.terminals, p.terminals[i+1])
		}
	}

	return
}

func expandProd(p production, macros map[string][]production) (out string, err error) {
	for 1 < len(p.terminals) {
		if p, err = expandProdOnce(p, macros); err != nil {
			return
		}
		if p, err = removeEmpty(p); err != nil {
			return
		}
	}

	out = p.terminals[0]
	return
}

func expand(s string, macros map[string][]production) (out string, err error) {
	var prod production
	if prod, err = expandBaseOnce(s, macros); err != nil {
		return
	}
	return expandProd(prod, macros)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	file, err := getFile()
	if err != nil {
		printErr(err)
		flag.Usage()
		return
	}
	if file != os.Stdin {
		defer file.Close()
	}

	var records [][]string
	if records, err = parseCookieJar(file); err != nil {
		printErr(err)
		return
	}

	var macros map[string][]production
	var lastmac string
	if macros, lastmac, err = cookieJarToMacros(records); err != nil {
		printErr(err)
		return
	}

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < *ntimes; i++ {
		var out string
		if out, err = expand(lastmac, macros); err != nil {
			printErr(err)
			return
		}
		fmt.Println(out)
	}
}
