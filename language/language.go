// Package language provides a convenient interface for retrieving the
// corresponding tree-sitter language object for a given file extension.
package language

import (
	"fmt"
	"path/filepath"
	"strings"
	"unsafe"

	// this language list is based on the top most popular programming, scripting, and markup languages
	// according to stack overflow survey: https://survey.stackoverflow.co/2025/technology
	"github.com/alexaandru/go-sitter-forest/ada"
	"github.com/alexaandru/go-sitter-forest/asm"
	"github.com/alexaandru/go-sitter-forest/bash"
	"github.com/alexaandru/go-sitter-forest/c"
	"github.com/alexaandru/go-sitter-forest/c_sharp"
	"github.com/alexaandru/go-sitter-forest/cobol"
	"github.com/alexaandru/go-sitter-forest/commonlisp"
	"github.com/alexaandru/go-sitter-forest/cpp"
	"github.com/alexaandru/go-sitter-forest/css"
	"github.com/alexaandru/go-sitter-forest/dart"
	"github.com/alexaandru/go-sitter-forest/elixir"
	"github.com/alexaandru/go-sitter-forest/erlang"
	"github.com/alexaandru/go-sitter-forest/fortran"
	"github.com/alexaandru/go-sitter-forest/fsharp"
	"github.com/alexaandru/go-sitter-forest/gdscript"
	"github.com/alexaandru/go-sitter-forest/gleam"
	golang "github.com/alexaandru/go-sitter-forest/go"
	"github.com/alexaandru/go-sitter-forest/groovy"
	"github.com/alexaandru/go-sitter-forest/html"
	"github.com/alexaandru/go-sitter-forest/java"
	"github.com/alexaandru/go-sitter-forest/javascript"
	"github.com/alexaandru/go-sitter-forest/kotlin"
	"github.com/alexaandru/go-sitter-forest/lua"
	"github.com/alexaandru/go-sitter-forest/matlab"
	"github.com/alexaandru/go-sitter-forest/ocaml"
	"github.com/alexaandru/go-sitter-forest/pascal"
	"github.com/alexaandru/go-sitter-forest/perl"
	"github.com/alexaandru/go-sitter-forest/php"
	"github.com/alexaandru/go-sitter-forest/powershell"
	"github.com/alexaandru/go-sitter-forest/prolog"
	"github.com/alexaandru/go-sitter-forest/python"
	"github.com/alexaandru/go-sitter-forest/r"
	"github.com/alexaandru/go-sitter-forest/ruby"
	"github.com/alexaandru/go-sitter-forest/rust"
	"github.com/alexaandru/go-sitter-forest/scala"
	"github.com/alexaandru/go-sitter-forest/sql"
	"github.com/alexaandru/go-sitter-forest/swift"
	"github.com/alexaandru/go-sitter-forest/typescript"
	"github.com/alexaandru/go-sitter-forest/zig"
	sitter "github.com/smacker/go-tree-sitter"
)

var langToFactory = make(map[string]func() unsafe.Pointer)
var langCache = make(map[string]*sitter.Language)

func init() {
	supportedLangs := []struct{
		factory func() unsafe.Pointer
		extensions []string
	}{
		{javascript.GetLanguage, []string{".js", ".mjs"}},
		{typescript.GetLanguage, []string{".ts", ".tsx"}},
		{python.GetLanguage, []string{".py"}},
		{golang.GetLanguage, []string{".go"}},
		{rust.GetLanguage, []string{".rs"}},
		{java.GetLanguage, []string{".java"}},
		{c.GetLanguage, []string{".c"}},
		{cpp.GetLanguage, []string{".cpp", ".hpp", ".hxx", ".hh", ".cc", ".cxx"}},
		{c_sharp.GetLanguage, []string{".cs"}},
		{bash.GetLanguage, []string{".sh"}},
		{html.GetLanguage, []string{".html", ".htm"}},
		{css.GetLanguage, []string{".css"}},
		{ruby.GetLanguage, []string{".rb"}},
		{php.GetLanguage, []string{".php", ".php3", ".phtml"}},
		{swift.GetLanguage, []string{".swift"}},
		{kotlin.GetLanguage, []string{".kt", ".kts"}},
		{scala.GetLanguage, []string{".scala"}},
		{sql.GetLanguage, []string{".sql"}},
		{lua.GetLanguage, []string{".lua"}},
		{perl.GetLanguage, []string{".pl"}},
		{powershell.GetLanguage, []string{".ps1"}},
		{dart.GetLanguage, []string{".dart"}},
		{r.GetLanguage, []string{".r"}},
		{zig.GetLanguage, []string{".zig"}},
		{ada.GetLanguage, []string{".adb", ".ads"}},
		{asm.GetLanguage, []string{".asm", ".s"}},
		{cobol.GetLanguage, []string{".cbl", ".cob"}},
		{commonlisp.GetLanguage, []string{".lisp", ".cl"}},
		{elixir.GetLanguage, []string{".ex", ".exs"}},
		{erlang.GetLanguage, []string{".erl", ".hrl"}},
		{fortran.GetLanguage, []string{".f", ".for", ".f90", ".f95", ".f03"}},
		{fsharp.GetLanguage, []string{".fs", ".fsi"}},
		{gdscript.GetLanguage, []string{".gd"}},
		{gleam.GetLanguage, []string{".gleam"}},
		{groovy.GetLanguage, []string{".groovy"}},
		{matlab.GetLanguage, []string{".m"}},
		{ocaml.GetLanguage, []string{".ml", ".mli"}},
		{pascal.GetLanguage, []string{".pas", ".pp"}},
		{prolog.GetLanguage, []string{".pro"}},
	}

	for _, lang := range supportedLangs {
		for _, ext := range lang.extensions {
			langToFactory[ext] = lang.factory
		}
	}
}

func FromFilename(filename string) (*sitter.Language, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	if lang, exists := langCache[ext]; exists {
		return lang, nil
	}

	factory, exists := langToFactory[ext]
	if !exists {
		return nil, fmt.Errorf("no language found for file extension %s", ext)
	}

	lang := sitter.NewLanguage(factory())
	langCache[ext] = lang

	return lang, nil
}
