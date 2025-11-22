// Package language provides a convenient interface for retrieving the
// corresponding tree-sitter language object for a given file extension.
package language

import (
	"fmt"
	"path/filepath"
	"strings"

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

type Info struct {
	Name       string
	SitterLang *sitter.Language
}

var extToLang = map[string]Info{
	".js":     {Name: "JavaScript", SitterLang: sitter.NewLanguage(javascript.GetLanguage())},
	".mjs":    {Name: "JavaScript", SitterLang: sitter.NewLanguage(javascript.GetLanguage())},
	".ts":     {Name: "TypeScript", SitterLang: sitter.NewLanguage(typescript.GetLanguage())},
	".tsx":    {Name: "TypeScript", SitterLang: sitter.NewLanguage(typescript.GetLanguage())},
	".py":     {Name: "Python", SitterLang: sitter.NewLanguage(python.GetLanguage())},
	".go":     {Name: "Go", SitterLang: sitter.NewLanguage(golang.GetLanguage())},
	".rs":     {Name: "Rust", SitterLang: sitter.NewLanguage(rust.GetLanguage())},
	".java":   {Name: "Java", SitterLang: sitter.NewLanguage(java.GetLanguage())},
	".c":      {Name: "C", SitterLang: sitter.NewLanguage(c.GetLanguage())},
	".cpp":    {Name: "C++", SitterLang: sitter.NewLanguage(cpp.GetLanguage())},
	".hpp":    {Name: "C++", SitterLang: sitter.NewLanguage(cpp.GetLanguage())},
	".hxx":    {Name: "C++", SitterLang: sitter.NewLanguage(cpp.GetLanguage())},
	".hh":     {Name: "C++", SitterLang: sitter.NewLanguage(cpp.GetLanguage())},
	".cc":     {Name: "C++", SitterLang: sitter.NewLanguage(cpp.GetLanguage())},
	".cxx":    {Name: "C++", SitterLang: sitter.NewLanguage(cpp.GetLanguage())},
	".cs":     {Name: "C#", SitterLang: sitter.NewLanguage(c_sharp.GetLanguage())},
	".sh":     {Name: "Bash", SitterLang: sitter.NewLanguage(bash.GetLanguage())},
	".html":   {Name: "HTML", SitterLang: sitter.NewLanguage(html.GetLanguage())},
	".htm":    {Name: "HTML", SitterLang: sitter.NewLanguage(html.GetLanguage())},
	".css":    {Name: "CSS", SitterLang: sitter.NewLanguage(css.GetLanguage())},
	".rb":     {Name: "Ruby", SitterLang: sitter.NewLanguage(ruby.GetLanguage())},
	".php":    {Name: "PHP", SitterLang: sitter.NewLanguage(php.GetLanguage())},
	".php3":   {Name: "PHP", SitterLang: sitter.NewLanguage(php.GetLanguage())},
	".phtml":  {Name: "PHP", SitterLang: sitter.NewLanguage(php.GetLanguage())},
	".swift":  {Name: "Swift", SitterLang: sitter.NewLanguage(swift.GetLanguage())},
	".kt":     {Name: "Kotlin", SitterLang: sitter.NewLanguage(kotlin.GetLanguage())},
	".kts":    {Name: "Kotlin", SitterLang: sitter.NewLanguage(kotlin.GetLanguage())},
	".scala":  {Name: "Scala", SitterLang: sitter.NewLanguage(scala.GetLanguage())},
	".sql":    {Name: "SQL", SitterLang: sitter.NewLanguage(sql.GetLanguage())},
	".lua":    {Name: "Lua", SitterLang: sitter.NewLanguage(lua.GetLanguage())},
	".pl":     {Name: "Perl", SitterLang: sitter.NewLanguage(perl.GetLanguage())},
	".ps1":    {Name: "PowerShell", SitterLang: sitter.NewLanguage(powershell.GetLanguage())},
	".dart":   {Name: "Dart", SitterLang: sitter.NewLanguage(dart.GetLanguage())},
	".r":      {Name: "R", SitterLang: sitter.NewLanguage(r.GetLanguage())},
	".zig":    {Name: "Zig", SitterLang: sitter.NewLanguage(zig.GetLanguage())},
	".adb":    {Name: "Ada", SitterLang: sitter.NewLanguage(ada.GetLanguage())},
	".ads":    {Name: "Ada", SitterLang: sitter.NewLanguage(ada.GetLanguage())},
	".asm":    {Name: "Assembly", SitterLang: sitter.NewLanguage(asm.GetLanguage())},
	".s":      {Name: "Assembly", SitterLang: sitter.NewLanguage(asm.GetLanguage())},
	".cbl":    {Name: "COBOL", SitterLang: sitter.NewLanguage(cobol.GetLanguage())},
	".cob":    {Name: "COBOL", SitterLang: sitter.NewLanguage(cobol.GetLanguage())},
	".lisp":   {Name: "Common Lisp", SitterLang: sitter.NewLanguage(commonlisp.GetLanguage())},
	".cl":     {Name: "Common Lisp", SitterLang: sitter.NewLanguage(commonlisp.GetLanguage())},
	".ex":     {Name: "Elixir", SitterLang: sitter.NewLanguage(elixir.GetLanguage())},
	".exs":    {Name: "Elixir", SitterLang: sitter.NewLanguage(elixir.GetLanguage())},
	".erl":    {Name: "Erlang", SitterLang: sitter.NewLanguage(erlang.GetLanguage())},
	".hrl":    {Name: "Erlang", SitterLang: sitter.NewLanguage(erlang.GetLanguage())},
	".f":      {Name: "Fortran", SitterLang: sitter.NewLanguage(fortran.GetLanguage())},
	".for":    {Name: "Fortran", SitterLang: sitter.NewLanguage(fortran.GetLanguage())},
	".f90":    {Name: "Fortran", SitterLang: sitter.NewLanguage(fortran.GetLanguage())},
	".f95":    {Name: "Fortran", SitterLang: sitter.NewLanguage(fortran.GetLanguage())},
	".f03":    {Name: "Fortran", SitterLang: sitter.NewLanguage(fortran.GetLanguage())},
	".fs":     {Name: "F#", SitterLang: sitter.NewLanguage(fsharp.GetLanguage())},
	".fsi":    {Name: "F#", SitterLang: sitter.NewLanguage(fsharp.GetLanguage())},
	".gd":     {Name: "GDScript", SitterLang: sitter.NewLanguage(gdscript.GetLanguage())},
	".gleam":  {Name: "Gleam", SitterLang: sitter.NewLanguage(gleam.GetLanguage())},
	".groovy": {Name: "Groovy", SitterLang: sitter.NewLanguage(groovy.GetLanguage())},
	".m":      {Name: "MATLAB", SitterLang: sitter.NewLanguage(matlab.GetLanguage())},
	".ml":     {Name: "OCaml", SitterLang: sitter.NewLanguage(ocaml.GetLanguage())},
	".mli":    {Name: "OCaml", SitterLang: sitter.NewLanguage(ocaml.GetLanguage())},
	".pas":    {Name: "Pascal", SitterLang: sitter.NewLanguage(pascal.GetLanguage())},
	".pp":     {Name: "Pascal", SitterLang: sitter.NewLanguage(pascal.GetLanguage())},
	".pro":    {Name: "Prolog", SitterLang: sitter.NewLanguage(prolog.GetLanguage())},
}

func FromFilename(filename string) (*Info, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if info, ok := extToLang[ext]; ok {
		return &info, nil
	}
	return nil, fmt.Errorf("no language found for file extension %s", ext)
}
