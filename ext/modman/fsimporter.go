package modman

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/d5/tengo/v2"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io"
	"io/fs"
	"lightbox/env"
	"lightbox/ext/transpile"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//func NewBaseImporter(basePath string) ImportFunc {
//	if strings.HasPrefix(basePath, "http://") || strings.HasPrefix(basePath, "https://") {
//		return NewHTTPImporter(basePath)
//	} else {
//		return NewDirImporter(basePath)
//	}
//}
//func NewDirImporter(basePath string) ImportFunc {
//	return NewDirImporterWith(env.Global(), basePath, tengo.SourceFileExtDefault)
//}
//

func ImportFromFS(fsys fs.FS, name string) (tengo.Importable, error) {
	return ImportFromFSWithTranspiler(fsys, name, &transpile.G)
}

func ImportFromFSWithTranspiler(fsys fs.FS, name string, t transpile.Transpiler) (tengo.Importable, error) {
	if _, err := fs.Stat(fsys, name); err != nil {
		return nil, err
	}
	if buf, err := fs.ReadFile(fsys, name); err == nil {
		buf, err = t.Transpile(buf)
		if err != nil {
			return nil, err
		}
		return &tengo.SourceModule{Src: buf}, nil
	} else {
		return nil, err
	}
}

//func ImportFromFile(fileName string) (tengo.Importable, error) {
//	if _, err := os.Stat(fileName); err != nil {
//		return nil, err
//	}
//	if buf, err := ioutil.ReadFile(fileName); err == nil {
//		buf, _ = transpile.Transpile(buf)
//		return &tengo.SourceModule{Src: buf}, nil
//	} else {
//		return nil, err
//	}
//}
const (
	pkgBase    = ".tengo_module"
	pkgNameFmt = "%s@%s" //name@version
	pkgTemp    = "%s@%s-tmp"
	manifest   = "manifest.yml"
)

type PackageMeta struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func emptyImporter(name string) tengo.Importable {
	return nil
}

func NewZipImporter(zipFile string, environ *env.Environment, t transpile.Transpiler, ext string) ImportFunc {
	//if _, err := os.Stat(pkgBase); err != nil {
	//	err = os.Mkdir(pkgBase, os.ModePerm)
	//}
	//pkgDir, err := unpackage(zipFile, pkgBase)
	//if err != nil {
	//	log.Errorf("unpackage %s error:%v", zipFile, err)
	//	return emptyImporter
	//}
	//return NewFSImporter(os.DirFS(filepath.Join(pkgBase, pkgDir)), environ, t, ext)
	f, err := NewZipImporterWithDest(zipFile, pkgBase, environ, t, ext)
	if err != nil {
		log.WithField("package", zipFile).Error(err)
	}
	return f
}
func NewZipImporterWithDest(zipFile string, dest string, environ *env.Environment, t transpile.Transpiler, ext string) (ImportFunc, error) {
	if _, err := os.Stat(dest); err != nil {
		err = os.Mkdir(dest, os.ModePerm)
	}
	pkgDir, err := unpackage(zipFile, dest)
	if err != nil {
		return nil, err
	}
	return NewFSImporter(os.DirFS(filepath.Join(pkgBase, pkgDir)), environ, t, ext), nil
}

func unzip(fsys fs.FS, zipFile string, dest string) (string, error) {
	//todo: 由于zip本身不支持抽象的fs.FS,所以zip文件全部载入内存中，内存消耗较大。 尤其是对只读取manifest文件的情况下
	meta := &PackageMeta{}
	buf, err := fs.ReadFile(fsys, zipFile)
	if err != nil {
		return "", err
	}
	bufReader := bytes.NewReader(buf)
	zr, err := zip.NewReader(bufReader, bufReader.Size())
	manifestBuf, err := fs.ReadFile(zr, manifest)
	if err == nil {
		err = yaml.Unmarshal(manifestBuf, meta)
		if err != nil {
			return "", fmt.Errorf("%s found but unmarshal error:%v", manifest, err)
		}
	}
	if meta.Name == "" {
		meta.Name = filepath.Base(zipFile)[0:strings.LastIndex(filepath.Base(zipFile), ".")]
	}
	if meta.Version == "" {
		meta.Version, err = sha1file(zipFile)
		if err != nil {
			return "", err
		}
	}
	tmpDir := filepath.Join(dest, fmt.Sprintf(pkgTemp, meta.Name, meta.Version))
	targetDir := fmt.Sprintf(pkgNameFmt, meta.Name, meta.Version)
	if _, err = os.Stat(filepath.Join(dest, targetDir)); err == nil {
		return targetDir, nil
	}
	if _, err := os.Stat(tmpDir); err == nil {
		_ = os.RemoveAll(tmpDir)
	}
	_ = os.Mkdir(tmpDir, os.ModePerm)
	//先创建所有目录
	for _, fi := range zr.File {
		if fi.FileInfo().IsDir() {
			err = os.MkdirAll(filepath.Join(tmpDir, fi.Name), os.ModePerm)
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		return "", err
	}
	for _, fi := range zr.File {
		if !fi.FileInfo().IsDir() {
			var input fs.File
			var output io.WriteCloser
			input, err = zr.Open(fi.Name)
			if err != nil {
				break
			}
			destFile := filepath.Join(tmpDir, fi.Name)
			output, err = os.Create(destFile)
			if err != nil {
				break
			}
			_, err = io.Copy(output, input)
		}
	}

	if err != nil {
		return "", err
	}
	err = os.Rename(tmpDir, filepath.Join(dest, targetDir))
	if err != nil {
		return "", err
	}
	return targetDir, nil
}

func unpackage(zipFile string, dest string) (string, error) {
	meta := &PackageMeta{}
	zr, err := zip.OpenReader(zipFile)
	if err != nil {
		return "", err
	}
	buf, err := fs.ReadFile(zr, manifest)
	if err == nil {
		err = yaml.Unmarshal(buf, meta)
		if err != nil {
			return "", fmt.Errorf("%s found but unmarshal error:%v", manifest, err)
		}
	}
	if meta.Name == "" {
		meta.Name = filepath.Base(zipFile)[0:strings.LastIndex(filepath.Base(zipFile), ".")]
	}
	if meta.Version == "" {
		meta.Version, err = sha1file(zipFile)
		if err != nil {
			return "", err
		}
	}
	tmpDir := filepath.Join(dest, fmt.Sprintf(pkgTemp, meta.Name, meta.Version))
	targetDir := fmt.Sprintf(pkgNameFmt, meta.Name, meta.Version)
	if _, err = os.Stat(filepath.Join(dest, targetDir)); err == nil {
		return targetDir, nil
	}
	if _, err := os.Stat(tmpDir); err == nil {
		_ = os.RemoveAll(tmpDir)
	}
	_ = os.Mkdir(tmpDir, os.ModePerm)
	//先创建所有目录
	for _, fi := range zr.File {
		if fi.FileInfo().IsDir() {
			err = os.MkdirAll(filepath.Join(tmpDir, fi.Name), os.ModePerm)
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		return "", err
	}
	for _, fi := range zr.File {
		if !fi.FileInfo().IsDir() {
			var input fs.File
			var output io.WriteCloser
			input, err = zr.Open(fi.Name)
			if err != nil {
				break
			}
			destFile := filepath.Join(tmpDir, fi.Name)
			output, err = os.Create(destFile)
			if err != nil {
				break
			}
			_, err = io.Copy(output, input)
		}
	}

	if err != nil {
		return "", err
	}
	err = os.Rename(tmpDir, filepath.Join(dest, targetDir))
	if err != nil {
		return "", err
	}
	return targetDir, nil
}
func sha1file(path string) (string, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return "", err
	}
	h := sha1.New()
	_, err = io.Copy(h, file)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

//NewFSImporter 简化,去除zip的导入
func NewFSImporter(fsys fs.FS, environ *env.Environment, t transpile.Transpiler, ext string) ImportFunc {
	return func(src string) tengo.Importable {
		if strings.Contains(src, "{") {
			src, _ = environ.Parse(src)
		}
		fi, err := fs.Stat(fsys, src)
		if err == nil && fi.IsDir() {
			src = filepath.Join(src, "index"+ext)
		}

		if path.Ext(src) == "" {
			src = src + ext
		}

		ret, err := ImportFromFSWithTranspiler(fsys, src, t)
		if err != nil {
			log.Trace("import file error:", err)
		}
		return ret

	}
}

//func NewDirImporterWith(environ *env.Environment, basePath string, ext string) ImportFunc {
//	return NewFSImporter(os.DirFS(basePath), environ, ext)
//}
//func NewHTTPImporter(baseUrl string) ImportFunc {
//	return func(srcName string) tengo.Importable {
//		if path.Ext(srcName) == "" {
//		}
//		url := path.Join(baseUrl, srcName)
//		return ImportUrl(url)
//	}
//}
