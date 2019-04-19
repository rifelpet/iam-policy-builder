package parser

import (
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)
// TODO:
// Use https://gist.github.com/0xdabbad00/581714de8f0957fce30efcb1634785a9
// for building condition keys

var awsImportPrefix = regexp.MustCompile("\"github.com/aws/aws-sdk-go/service/(\\w+)\"")

var blacklist = []string{"mock", "vendor"}

// Parse parses a go project path
func Parse(path string) error {
	fileset, fileCount, err := buildFileSet(path)
	if err != nil {
		return err
	}
	awsFiles, err := parseAWSFiles(fileset, fileCount)
	if err != nil {
		return err
	}
	fmt.Println(awsFiles)
	for _, f := range awsFiles {
		awsImports := getAWSServices(f)
		fmt.Println(awsImports)
		ast.Print(fileset, f)

		ast.Inspect(f, func(n ast.Node) bool {
			var s string
			switch x := n.(type) {
			case *ast.BasicLit:
				s = x.Value
			case *ast.Ident:
				s = x.Name
			case *ast.AssignStmt:
				rhs := x.Rhs
				for _, expr := range rhs {
					switch y := expr.(type) {
					case *ast.CallExpr:
						fun := y.Fun
						switch z := fun.(type) {
						case *ast.SelectorExpr:
							pkg := z.X
							switch q := pkg.(type) {
							case *ast.Ident:
								pkgFn := q.Name
								if pkgFn.Name == "New" {
									for _, imp := range awsImports {
										if imp.ImportName == pkg {
											fmt.Println(pkg)
										}
									}
								}
							}
						}

					}
				}

			}
			if s != "" {
				//fmt.Printf("%s:\t%s\n", fileset.Position(n.Pos()), s)
			}
			return true
		})
		for _, f := range f.Decls {
			fn, ok := f.(*ast.FuncDecl)
			if !ok {
				continue
			}
			fmt.Println(fn.Name.Name)
		}

	}
	return nil
}

func parseAWSFiles(fileset *token.FileSet, fileCount int) ([]*ast.File, error) {
	awsFiles := make([]*ast.File, 0)
	parseCount := 0
	fileset.Iterate(
		func(f *token.File) bool {
			parsed, err := goparser.ParseFile(fileset, f.Name(), nil, goparser.AllErrors)
			if err != nil {
				return false
			}
			for _, imp := range parsed.Imports {
				if awsImportPrefix.MatchString(imp.Path.Value) {
					awsFiles = append(awsFiles, parsed)
					fmt.Println(parsed.Name.Name)
				}
			}
			parseCount++
			return parseCount <= fileCount
		},
	)
	return awsFiles, nil
}

func buildFileSet(path string) (*token.FileSet, int, error) {
	goFiles := token.NewFileSet()
	fileCount := 0
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") || ignore(path) {
				return nil
			}
			body, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			file := goFiles.AddFile(path, goFiles.Base(), int(info.Size()))
			file.SetLinesForContent(body)
			fileCount++
			return nil
		})
	if err != nil {
		return nil, -1, err
	}
	return goFiles, fileCount, nil
}

func ignore(filename string) bool {
	for _, term := range blacklist {
		if strings.Contains(filename, term) {
			return true
		}
	}
	return false
}

type AWSServiceImport struct {
	ImportName string
	Service    string
}

func getAWSServices(file *ast.File) []AWSServiceImport {
	services := make([]AWSServiceImport, 0)
	for _, imp := range file.Imports {
		importMatch := awsImportPrefix.FindStringSubmatch(imp.Path.Value)
		if len(importMatch) > 1 {
			service := importMatch[1]
			importName := service
			if imp.Name != nil {
				importName = imp.Name.Name
			}
			serviceImport := AWSServiceImport{
				ImportName: importName,
				Service:    service,
			}
			services = append(services, serviceImport)
		}
	}
	return services
}
