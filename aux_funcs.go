package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
)

func CommandInPath(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func Copy(src, dst string) (err error) {
	sourceFileStat, err := os.Stat(src)
	LogAndExit(err)

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	LogAndExit(err)

	defer source.Close()

	destination, err := os.Create(dst)
	LogAndExit(err)

	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func CreateDir(path string) {
	err := os.MkdirAll(path, os.ModePerm)
	LogAndExit(err)
}

func CreateFile(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	LogAndExit(err)

	return file.Close()
}

func Expand(path string) string {
	usr, err := user.Current()
	if err != nil {
		log.Println("Can not get current user: ", err)
		runtime.Goexit()
	}

	return strings.Replace(path, "~", usr.HomeDir, 1)
}

func Exists(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}

func MakeTmpDir() string {
	dest, err := ioutil.TempDir(os.TempDir(), "demetra")
	LogAndExit(err)

	return dest
}

func GetStem(uri string) string {
	fname := path.Base(uri)
	return strings.Split(fname, ".")[0]
}

func GetSstateCacheDir() string {
	xdg_cache_dir, err := os.UserCacheDir()
	LogAndExit(err)

	cache_dir := xdg_cache_dir + "/demetra/sstate-cache"
	CreateDir(cache_dir)

	return cache_dir
}

func GetDlDir() string {
	xdg_cache_dir, err := os.UserCacheDir()
	LogAndExit(err)

	dl_dir := xdg_cache_dir + "/demetra/downloads"
	CreateDir(dl_dir)

	return dl_dir
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

func RemoveContents(dir string) {
	d, err := os.Open(dir)
	LogAndExit(err)

	defer d.Close()
	names, err := d.Readdirnames(-1)
	LogAndExit(err)

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		LogAndExit(err)
	}
}

func LogAndExit(err error) {
	if err != nil {
		log.Print(err)
		debug.PrintStack()
		runtime.Goexit()
	}
}
