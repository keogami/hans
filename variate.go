// the algo for this command would be moved into another package
// because it is too ambitious in terms of complexity.
// moving it into another package will give it more room to grow.
package main

import (
	"fmt"
	"regexp"

	"github.com/urfave/cli/v2"
)

// Matcher is a func that indicates whether the input word is applicable for
// a corresponding variation
type Match func(word string) bool

// Regex creates a Match func using a regular expression
// panics if the expr doesn't compiles
func Regex(r *regexp.Regexp) Match {
	return func(word string) bool {
		return r.Match([]byte(word))
	}
}

// Vary is a pure func that returns a single, reproducible (deterministic) variation
// of its input word
type Vary func(word string) string

func ReplaceLast(r *regexp.Regexp, rep string) Vary {
	return func(word string) string {
		l := r.FindAllStringSubmatchIndex(word, -1)
		lastMatchStart := l[len(l)-1][0]
		rem := word[lastMatchStart:]
		rem = r.ReplaceAllString(rem, rep)
		res := word[:lastMatchStart] + rem
		return res
	}
}

// Variator encodes, whether a variation is applicable, and the variation routine
type Variator struct {
	Match Match
	Vary  Vary
}

// Variators collects together all possible variators
type Variators []Variator

// For generates a list of possible variations for the input word
func (v Variators) For(word string) (variations []string) {
	for _, variator := range v {
		if variator.Match(word) {
			variations = append(variations, variator.Vary(word))
		}
	}
	return
}

func unique(list []string) (u []string) {
	cache := make(map[string]bool)
	for _, item := range list {
		if _, ok := cache[item]; !ok {
			u = append(u, item)
			cache[item] = true
		}
	}
	return
}

// variate recursively generates variations for the given word using the given variators
func variate(v Variators, word string) []string {
	variations := v.For(word) // create variations for this level
	if len(variations) == 0 { // this means the word can't be varied further
		return nil // in which case just return [a list of] nothing
	}

	// subVariants is an aux slice to store variants of each variant
	// it exits cuz using `variations` slice itself would cause problems
	// in the loop below. (writing while iterating problem)
	subVariants := make([]string, 0, len(variations))

	for _, variation := range variations { // generate variants for each variation
		subVariants = append(subVariants, variate(v, variation)...)
	}

	return unique(append(variations, subVariants...)) // return all the unique variations
}

func RemoveRedundantA(context *regexp.Regexp, repl *regexp.Regexp) Vary {
	return func(word string) string {
		spl := repl.FindAllStringSubmatchIndex(word, -1)
		for i := len(spl) - 1; i >= 0; i-- {
			mat := spl[i][0]
			subword := word[mat:]
			if !context.MatchString(subword) {
				continue
			}
			res := context.ReplaceAllString(subword, "$1$2")
			return word[:mat] + res
		}

		panic("replaceA: no redundant `a` found in " + word)
	}
}

func createVariators() Variators {
	aa := regexp.MustCompile(`aa`)
	aContext := regexp.MustCompile(`([^aeiou])a([^aeiou])`)
	aRepl := regexp.MustCompile(`[^aeiou]a`)
	ai := regexp.MustCompile(`ai`)
	hi := regexp.MustCompile(`hi$`)
	ee := regexp.MustCompile(`ee`)
	oo := regexp.MustCompile(`oo`)
	return Variators{
		{
			Match: Regex(aContext),
			Vary:  RemoveRedundantA(aContext, aRepl),
		},
		{
			Match: Regex(aa),
			Vary:  ReplaceLast(aa, "a"),
		},
		{
			Match: Regex(ai),
			Vary:  ReplaceLast(ai, "e"),
		},
		{
			Match: Regex(hi),
			Vary:  ReplaceLast(hi, "i"),
		},
		{
			Match: Regex(ee),
			Vary:  ReplaceLast(ee, "i"),
		},
		{
			Match: Regex(oo),
			Vary:  ReplaceLast(oo, "u"),
		},
	}
}

func variateMain(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {
		return ErrNotEnoughArgs
	}

	word := ctx.Args().First()
	variations := variate(createVariators(), word)
	res := fmt.Sprintf("%v", variations)

	fmt.Printf("[%d]%s\n", len(variations), res)

	return nil
}
