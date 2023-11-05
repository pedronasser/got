package transform

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build/constraint"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/ast/astutil"
)

type gotTransformer struct {
	baseDir     string
	currentFile string

	methods    map[string]ExtractedMethod
	decorators map[string]ExtractedDecorator
}

type ExtractedMethod = func(...interface{}) interface{}
type ExtractedDecorator = func(c *TransformContext) (err error)

func GotTransform(baseDir string) *gotTransformer {
	return &gotTransformer{
		baseDir:     baseDir,
		currentFile: "",

		methods:    map[string]ExtractedMethod{},
		decorators: map[string]ExtractedDecorator{},
	}
}

func (t *gotTransformer) Execute() error {
	targetFiles := LookupGoFiles(t.baseDir)
	for _, path := range targetFiles {
		err := t.transformFile(path)
		if err != nil {
			return fmt.Errorf("Failed to transform file `%s`: \n\t%v", path, err)
		}
	}

	return nil
}

func (t *gotTransformer) transformFile(path string) error {
	t.currentFile = path

	f, err := os.Open(path)
	if err != nil {
		println(fmt.Errorf("Failed to open file: %s", err))
		os.Exit(1)
	}
	defer f.Close()

	srcBytes, err := io.ReadAll(f)
	if err != nil {
		println(fmt.Errorf("Failed to read file: %s", err))
		os.Exit(1)
	}

	exportedDecorators = []string{}
	exportedMethods = []string{}

	src := bytes.NewBuffer(srcBytes)

	usages, err := extractAttributeUsages(src)
	if err != nil {
		return err
	}

	var isModified bool
	if len(usages) > 0 {
		t.log("Applying builtin only attributes...")

		var isMod bool
		if isMod, err = t.processAttributeTransforms(src, &usages, true); err != nil {
			return err
		}

		if isMod {
			isModified = true
		}

		err = t.loadExtractedFunctions()
		if err != nil {
			return err
		}

		t.log("Applying all remaining attributes...")
		if isMod, err = t.processAttributeTransforms(src, &usages, false); err != nil {
			return err
		}

		if isMod {
			isModified = true
		}
	}

	if !isModified {
		return nil
	}

	t.log("Cleaning up...")
	err = cleanupSource(src)
	if err != nil {
		return err
	}

	if bytes.Equal(src.Bytes(), srcBytes) {
		t.log("No changes detected. Skipping...")
		return nil
	}

	goFile := strings.Replace(path, GO_FILE_EXTENSION, "_generated.go", 1)
	t.log("Writing to file:", goFile)
	err = writeGoFile(goFile, src)
	if err != nil {
		return err
	}

	t.log("Executing goimports on file")
	err = executeGoImports(goFile)
	if err != nil {
		return err
	}

	return nil
}

func (t *gotTransformer) applyTemplate(src *bytes.Buffer) error {
	result := bytes.NewBuffer([]byte{})
	fns := template.FuncMap{}
	for name, method := range t.methods {
		fns[name] = method
	}

	tpl, err := template.New("").Funcs(fns).Parse(src.String())
	if err != nil {
		return err
	}

	err = tpl.Execute(result, nil)
	if err != nil {
		return err
	}

	src.Reset()
	_, _ = src.Write(result.Bytes())

	return nil
}

func (t *gotTransformer) loadExtractedFunctions() error {
	for _, methodName := range exportedMethods {
		fn, err := loadExtractedFunction[ExtractedMethod](filepath.Join(t.baseDir, GOT_BUILD_DIR, GOT_METHODS_DIR, fmt.Sprintf("%s.so", methodName)))
		if err != nil {
			log(fmt.Sprintf("Failed to load method `%s`: %s", methodName, err))
		}
		t.log("Extracted method:", methodName)
		t.methods[methodName] = fn
	}

	for _, decoratorName := range exportedDecorators {
		fn, err := loadExtractedFunction[ExtractedDecorator](filepath.Join(t.baseDir, GOT_BUILD_DIR, GOT_DECORATORS_DIR, fmt.Sprintf("%s.so", decoratorName)))
		if err != nil {
			log(fmt.Sprintf("Failed to load decorator `%s`: %s", decoratorName, err))
			return err
		}
		t.log("Extracted decorator:", decoratorName)
		t.decorators[decoratorName] = fn
	}

	return nil
}

func (t *gotTransformer) processAttributeTransforms(
	src *bytes.Buffer,
	usages *[]*attributesUsage,
	builtinOnly bool,
) (isModified bool, err error) {
	if len(*usages) == 0 {
		return false, nil
	}

	fset := token.NewFileSet()
	pfile, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return false, fmt.Errorf("Failed to parse file: %v", err)
	}

	buf := make([]byte, src.Len())
	copy(buf, src.Bytes())
	updatedSrc := bytes.NewBuffer(buf)

	srcOffset := 0
	process := func(c *astutil.Cursor, usage *attributesUsage) error {
		originalLength := c.Node().End() - c.Node().Pos()
		pos := int(c.Node().Pos()) - 1 - srcOffset

		context := &TransformContext{
			Cursor:      c,
			currentNode: c.Node(),
			File:        pfile,
			fileSrc:     src.Bytes(),
		}

		for _, attribute := range usage.attributes {
			context.args = attribute.Arguments
			attributeName := attribute.Name

			if handler, ok := BuiltinAttributes[attributeName]; ok {
				t.log(fmt.Sprintf("Executing builtin attribute: `%s` on position %d", attributeName, pos))
				err := handler(context)
				if err != nil {
					return fmt.Errorf("Failed to execute decorator `%s`: %v", attributeName, err)
				}
				usage.isApplied = true
			}

			if !builtinOnly {
				if handler, ok := t.decorators[attributeName]; ok {
					t.log(fmt.Sprintf("Executing decorator: `%s` on position %d", attributeName, pos))
					err := handler(context)
					if err != nil {
						return fmt.Errorf("Failed to execute decorator `%s`: %v", attributeName, err)
					}
					usage.isApplied = true
				}
			}
		}

		if context.modified {
			t.log(fmt.Sprintf("Attribute `%s` modified source", usage.attributes[0].Name), "")
			updatedNode := bytes.NewBuffer([]byte{})
			printer.Fprint(updatedNode, fset, context.currentNode)
			srcOffset += updatedNode.Len() - int(originalLength)

			updatedSrc.Reset()
			err := printer.Fprint(updatedSrc, fset, pfile)
			if err != nil {
				return fmt.Errorf("Failed to update source: %v", err)
			}

			isModified = true
		}

		return nil
	}

	var processErr error

	for _, usage := range *usages {
		if usage.isApplied {
			continue
		}

		astutil.Apply(pfile, nil, func(c *astutil.Cursor) bool {
			n := c.Node()
			if n == nil {
				return true
			}

			node := c.Node()

			nodePos := int(node.Pos()) - 1 - srcOffset
			endPos := int(node.End()) - 1 - srcOffset

			var name string
			switch v := node.(type) {
			case *ast.GenDecl:
				switch v.Tok {
				case token.TYPE:
					name = v.Specs[0].(*ast.TypeSpec).Name.Name
				default:
				}

			case *ast.FuncDecl:
				name = v.Name.Name

				// Change endPos to the start of the function body
				endPos = int(v.Body.Lbrace) - 1 - srcOffset

			case *ast.AssignStmt, *ast.IfStmt:

			default:
				return true
			}

			_ = name

			if nodePos > usage.commentPos {
				separation := string(src.Bytes()[usage.commentPos:nodePos])
				separation = strings.Replace(separation, "\n", "", 1)
				separation = strings.TrimSpace(separation)
				if separation != "" {
					return true
				}
			}

			if usage.commentPos > endPos {
				return true
			}

			processErr = process(c, usage)
			if processErr != nil {
				return false
			}

			return false
		})
		if processErr != nil {
			return false, err
		}
	}

	src.Reset()
	_, _ = src.Write(updatedSrc.Bytes())

	return isModified, nil
}

type DeletedNode struct{}

func (e *DeletedNode) Pos() token.Pos {
	return 0
}

func (e *DeletedNode) End() token.Pos {
	return 0
}

type TransformContext struct {
	*astutil.Cursor
	*ast.File
	fileSrc []byte
	args    []string

	modified    bool
	currentNode ast.Node
}

func (t *TransformContext) Args() []string {
	return t.args
}

func (t *TransformContext) ASTFile() *ast.File {
	return t.File
}

func (t *TransformContext) FileSrc() []byte {
	return t.fileSrc
}

func (t *TransformContext) Node() ast.Node {
	return t.currentNode
}

func (t *TransformContext) Replace(node ast.Node) {
	t.Cursor.Replace(node)
	t.modified = true
	t.currentNode = node
}

func (t *TransformContext) Delete() {
	t.Cursor.Delete()
	t.modified = true
	t.currentNode = &DeletedNode{}
}

func (t *TransformContext) InsertBefore(node ast.Node) {
	t.Cursor.InsertBefore(node)
	t.modified = true
}

func (t *TransformContext) InsertAfter(node ast.Node) {
	t.Cursor.InsertAfter(node)
	t.modified = true
}

func cleanupSource(src *bytes.Buffer) error {
	fset := token.NewFileSet()

	pfile, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("Failed to parse file: %v", err)
	}

	handleComment := func(comment *ast.Comment) string {
		if constraint.IsGoBuild(comment.Text) {
			exp, err := constraint.Parse(comment.Text)
			if err != nil {
				return ""
			}

			if exp.Eval(hasGeneratedTag) {
				return comment.Text
			} else if strings.Contains(comment.Text, "generated") {
				result := removeTag(exp, func(tag string) bool {
					return strings.EqualFold(tag, "generated")
				})
				exp = result

				if result != nil {
					return "//go:build " + result.String()
				}
			}

			if exp == nil {
				exp = &constraint.TagExpr{Tag: "generated"}
			} else {
				exp = &constraint.AndExpr{
					X: exp,
					Y: &constraint.TagExpr{Tag: "generated"},
				}
			}

			return "//go:build " + exp.String()
		}

		return ""
	}

	astutil.Apply(pfile, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		if n == nil {
			return true
		}

		if comment, ok := n.(*ast.Comment); ok {
			commentText := handleComment(comment)

			if commentText == "" {
				c.Delete()
			} else {
				comment.Text = commentText
			}

			return true
		}

		return true
	})

	updatedComments := []*ast.CommentGroup{}
	for _, comments := range pfile.Comments {
		updatedGroup := []*ast.Comment{}
		for _, c := range comments.List {
			commentText := handleComment(c)
			if commentText != "" {
				c.Text = commentText
				updatedGroup = append(updatedGroup, c)
			}
		}

		comments.List = updatedGroup
		if len(comments.List) > 0 {
			updatedComments = append(updatedComments, comments)
		}
	}
	pfile.Comments = updatedComments

	src.Reset()
	printer.Fprint(src, fset, pfile)

	return nil
}

func (t *gotTransformer) log(args ...interface{}) {
	log(append([]interface{}{t.currentFile + ":"}, args...)...)
}

type attributesUsage struct {
	commentPos int
	attributes []AttributeInstruction
	isApplied  bool
}

func extractAttributeUsages(src io.Reader) ([]*attributesUsage, error) {
	var usages []*attributesUsage

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse file: %v", err)
	}

	for _, comments := range file.Comments {
		for _, comment := range comments.List {
			usage := extractComment(comment)
			if usage != nil {
				usages = append(usages, usage)
			}
		}
	}

	return usages, nil
}

func extractComment(comment *ast.Comment) *attributesUsage {
	if constraint.IsGoBuild(comment.Text) {
		return nil
	}

	line := strings.TrimSpace(comment.Text)
	if !IsLineGotPrefixed(line) {
		return nil
	}

	parser := NewInstructionParser(strings.NewReader(line), AllInstructions)
	_, err := parser.Parse()
	if err != nil {
		return nil
	}
	var usage *attributesUsage

	for _, instruction := range parser.result {
		if instruction.Type() == AttributeInstructionType {
			attr := instruction.(AttributeInstruction)

			if usage == nil {
				pos := int(comment.End())
				usage = &attributesUsage{
					commentPos: pos - 1,
					attributes: []AttributeInstruction{},
				}
			}

			usage.attributes = append(usage.attributes, attr)
		} else {
			return nil
		}
	}

	return usage
}

var hasGeneratedTag = func(tag string) bool {
	return strings.EqualFold(tag, "generated")
}

func removeTag(expr constraint.Expr, shouldRemove func(string) bool) constraint.Expr {
	switch v := expr.(type) {
	case *constraint.AndExpr:
		v.X = removeTag(v.X, shouldRemove)
		v.Y = removeTag(v.Y, shouldRemove)
		if v.X == nil {
			return v.Y
		}
		if v.Y == nil {
			return v.X
		}
	case *constraint.OrExpr:
		v.X = removeTag(v.X, shouldRemove)
		v.Y = removeTag(v.Y, shouldRemove)
		if v.X == nil {
			return v.Y
		}
		if v.Y == nil {
			return v.X
		}
	case *constraint.NotExpr:
		v.X = removeTag(v.X, shouldRemove)
		if v.X == nil {
			return nil
		}
	case *constraint.TagExpr:
		if shouldRemove(v.Tag) {
			return nil
		}
	}

	return expr
}
