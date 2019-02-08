package parser

import (
	"encoding/json"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rifelpet/iam-policy-builder/pkg/builder"
	"github.com/rifelpet/iam-policy-builder/pkg/iam"
)

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
	awsServices := make([]*iam.AWSServiceUsage, 0)
	for _, f := range awsFiles {
		fileServices := getAWSServices(f)

		// First parse to find the names of the client variables
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
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
								for _, awsService := range fileServices {
									if awsService.ImportName == pkgFn && z.Sel.Name == "New" {
										switch w := x.Lhs[0].(type) {
										case *ast.Ident:
											awsService.ClientNames = append(awsService.ClientNames, w.Name)
										}
									}
								}
							}
						}
					}
				}
			}
			return true
		})

		// Next, parse to find function calls from the client variables
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.AssignStmt:
				callExpr := x.Rhs[0]
				switch y := callExpr.(type) {
				case *ast.CallExpr:
					fn := y.Fun
					switch z := fn.(type) {
					case *ast.SelectorExpr:
						switch w := z.X.(type) {
						case *ast.Ident:
							for _, service := range fileServices {
								for _, clientName := range service.ClientNames {
									if clientName == w.Name {
										functionCall := z.Sel.Name
										service.FunctionCalls = append(service.FunctionCalls, functionCall)
									}
								}
							}
						}
					}
				}
			}
			return true
		})
		awsServices = append(awsServices, fileServices...)
	}
	doc, err := builder.BuildDocument(awsServices)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(doc, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

// parseAWSFiles returns a list of files in the project that contain AWS service imports
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
				}
			}
			parseCount++
			return parseCount <= fileCount
		},
	)
	return awsFiles, nil
}

// buildFileSet creates a set of files from a project path
// returning a *token.FileSet and the number of files
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

// getAWSServices creates a list of AWSServiceImports in a given file
func getAWSServices(file *ast.File) []*iam.AWSServiceUsage {
	services := make([]*iam.AWSServiceUsage, 0)
	for _, imp := range file.Imports {
		importMatch := awsImportPrefix.FindStringSubmatch(imp.Path.Value)
		if len(importMatch) > 1 {
			service := importMatch[1]
			importName := service
			if imp.Name != nil {
				importName = imp.Name.Name
			}
			serviceImport := iam.AWSServiceUsage{
				ImportName:    importName,
				Service:       service,
				ClientNames:   make([]string, 0),
				FunctionCalls: make([]string, 0),
			}
			services = append(services, &serviceImport)
		}
	}
	return services
}
