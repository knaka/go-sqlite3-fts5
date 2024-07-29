package main

import (
	"archive/zip"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	neturl "net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

func main() {
	targetFileBases := os.Args[1:]
	libVersion, _, sourceID := sqlite3.Version()
	divs := strings.SplitN(sourceID, " ", 2)
	yearStr := strings.SplitN(divs[0], "-", 3)[0]
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		log.Fatalln(err)
	}
	divs = strings.SplitN(libVersion, ".", 3)
	major := divs[0]
	minor := divs[1]
	patch := divs[2]
	branch := "0" // Released version does not have branch number.
	sourceVersion := fmt.Sprintf("%s%02s%02s%02s", major, minor, patch, branch)
	tempDir, err := os.MkdirTemp("", "fetcher")
	if err != nil {
		log.Fatalln(err)
	}
	defer (func() { _ = os.RemoveAll(tempDir) })()
	dirName := "sqlite-preprocessed-" + sourceVersion
	zipFileBase := dirName + ".zip"
	var resp *http.Response
	for i := 0; i <= 1; i++ {
		url := neturl.URL{
			Scheme: "https",
			Host:   "www.sqlite.org",
			Path:   path.Join(strconv.Itoa(year+i), zipFileBase),
		}
		resp, err = http.Get(url.String())
		if err != nil {
			log.Fatalln(err)
		}
		if resp.StatusCode == http.StatusOK {
			break
		}
		_ = resp.Body.Close()
	}
	defer (func() { _ = resp.Body.Close() })()
	zipFilePath := path.Join(tempDir, zipFileBase)
	zipWriter, err := os.Create(zipFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer (func() { _ = zipWriter.Close() })()
	_, err = io.Copy(zipWriter, resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	err = zipWriter.Close()
	if err != nil {
		log.Fatalln(err)
	}
	zipFileReader, err := os.Open(zipFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer (func() { _ = zipFileReader.Close() })()
	zipReader, err := zip.NewReader(zipFileReader, resp.ContentLength)
	if err != nil {
		log.Fatalln(err)
	}
	defer (func() { _ = zipFileReader.Close() })()
	for _, file := range zipReader.File {
		base := path.Base(file.Name)
		if !slices.Contains(targetFileBases, base) {
			continue
		}
		(func() {
			fileReader, err := file.Open()
			if err != nil {
				log.Fatalln(err)
			}
			defer (func() { _ = fileReader.Close() })()
			filePath := filepath.Join(".", base)
			fileWriter, err := os.Create(filePath)
			if err != nil {
				log.Fatalln(err)
			}
			defer (func() { _ = fileWriter.Close() })()
			_, err = io.Copy(fileWriter, fileReader)
			if err != nil {
				log.Fatalln(err)
			}
		})()
	}
}
