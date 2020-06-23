package assets

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	texttmpl "text/template"
)

const (
	cssTagTemplate    = `<link href="%s" rel="stylesheet" type="text/css" />`
	jsTagTemplate     = `<script src="%s" type="text/javascript" ></script>`
	cssInlineTemplate = `<style>%s</style>`
	jsInlineTemplate  = `<script>%s</script>`
)

var (
	TemplateFunctions = texttmpl.FuncMap{
		"cssTag":    CSSTag,
		"cssInline": CSSInline,
		"jsTag":     JSTag,
		"jsInline":  JSInline,
		"imgPath":   ImgPath,
	}
)

func hash(bytes []byte) string {
	sum := md5.Sum(bytes)
	return hex.EncodeToString([]byte(sum[:]))
}

func combine(writer *bytes.Buffer, paths ...string) error {
	for _, path := range paths {
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		writer.Write(bytes)
		writer.WriteString("\n\n")
	}
	return nil
}

func filePrefix(compiledName string) string {
	return fmt.Sprintf("__%s-", compiledName)
}

func findFile(path string, compiledName string, ext string, abs bool) (string, error) {
	glob := filepath.Join(path, filePrefix(compiledName)+"*"+ext)
	files, err := filepath.Glob(glob)
	if err != nil {
		return "", err
	}

	if len(files) > 1 {
		return "", fmt.Errorf("More than one file found at %s with name %s", path, compiledName)
	}

	if len(files) < 1 {
		return "", fmt.Errorf("No files found at %s with name %s", path, compiledName)
	}

	if abs {
		return filepath.Join("/", files[0]), nil
	}

	return files[0], nil
}

func removeGlob(outDir string, compiledName string, ext string) error {
	glob := filepath.Join(outDir, filePrefix(compiledName)+"*"+ext)
	files, err := filepath.Glob(glob)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}

func outPath(outDir string, compiledName string, buf []byte, ext string) string {
	return filepath.Join(outDir, filePrefix(compiledName)+hash(buf)+ext)
}

func compile(outDir string, compiledName string, ext string, paths ...string) error {
	var writer bytes.Buffer
	var min []byte

	/* concat all files together */
	err := combine(&writer, paths...)
	if err != nil {
		return err
	}

	/* delete any other asset with the same compiledName */
	err = removeGlob(outDir, compiledName, ext)
	if err != nil {
		return err
	}

	/* TODO: minify the concat file */
	min = writer.Bytes()

	/* determine the path with md5 */
	dst := outPath(outDir, compiledName, min, ext)

	/* write out the file */
	err = ioutil.WriteFile(dst, min, 0644)
	if err != nil {
		return err
	}

	return nil
}

func CSSCompile(outDir string, compiledName string, paths ...string) error {
	return compile(outDir, compiledName, ".min.css", paths...)
}

func JSCompile(outDir string, compiledName string, paths ...string) error {
	return compile(outDir, compiledName, ".min.js", paths...)
}

/* for any images or favicons that we want to set a long expiry on */
func ImgCompile(outDir string, compiledName string, ext string, path string) error {
	return compile(outDir, compiledName, ext, path)
}

func inline(path string, compiledName string, ext string) (template.HTML, error) {
	var tmplt string

	file, err := findFile(path, compiledName, ext, false)
	if err != nil {
		return template.HTML(""), err
	}

	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return template.HTML(""), err
	}

	if strings.HasSuffix(ext, "css") {
		tmplt = cssInlineTemplate
	} else if strings.HasSuffix(ext, "js") {
		tmplt = jsInlineTemplate
	} else {
		return template.HTML(""), fmt.Errorf("No inline template for ext %s", ext)
	}

	return template.HTML(fmt.Sprintf(tmplt, string(bytes[:]))), nil
}

func CSSInline(path string, compiledName string) (template.HTML, error) {
	return inline(path, compiledName, ".min.css")
}

func JSInline(path string, compiledName string) (template.HTML, error) {
	return inline(path, compiledName, ".min.js")
}

func tag(path string, compiledName string, ext string) (template.HTML, error) {
	var tmplt string

	file, err := findFile(path, compiledName, ext, true)
	if err != nil {
		return template.HTML(""), err
	}

	if strings.HasSuffix(ext, "css") {
		tmplt = cssTagTemplate
	} else if strings.HasSuffix(ext, "js") {
		tmplt = jsTagTemplate
	} else {
		return template.HTML(""), fmt.Errorf("No tag template for ext %s", ext)
	}

	return template.HTML(fmt.Sprintf(tmplt, file)), nil
}

func CSSTag(path string, compiledName string) (template.HTML, error) {
	return tag(path, compiledName, ".min.css")
}

func JSTag(path string, compiledName string) (template.HTML, error) {
	return tag(path, compiledName, ".min.js")
}

func ImgPath(path string, compiledName string, ext string) (string, error) {
	return findFile(path, compiledName, ext, true)
}
