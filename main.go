package main

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func isIgnored(line string) bool {

	ignored := []string{
		".lock",
		"package-lock.json",
		"/Migrations/",
		".git/",
		"node_modules/",
		".blackbox/",
		".gpg",
	}

	for _, needle := range ignored {
		if strings.Contains(line, needle) {
			return true
		}
	}

	return false
}

func shannon(filename string, line string, lineno int) {

	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=#!§$%&()[]|{}*-_.:,;\\'\"?"

	words := strings.Fields(line)

	for _, word := range words {

		var stringSet []string
		count := 0
		letters := ""

		for _, char := range word {

			if strings.ContainsRune(charset, char) {
				letters = letters + string(char)
				count++
			} else {
				if count > 16 {
					stringSet = append(stringSet, letters)
				}
				letters = ""
				count = 0
			}
		}

		if count > 16 {
			stringSet = append(stringSet, letters)
		}

		for _, token := range stringSet {

			sum := float64(0)
			strCount := len(token)

			for _, char := range charset {
				cnt := strings.Count(token, string(char))
				temp := float64(cnt) / float64(strCount)
				if temp > 0 {
					sum = sum + ((0 - temp) * math.Log2(temp))
				}
			}

			if sum > 5 {
				logrus.Warn(filename + ":" + fmt.Sprintf("%d", lineno) + " (Score: " + fmt.Sprintf("%f", sum))
				logrus.Info(token)
				fmt.Println()
			}
		}
	}
}

func processFile(filename string) error {

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	buffer := make([]byte, 512)

	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}
	contentType := http.DetectContentType(buffer[:n])
	if contentType != "text/plain; charset=utf-8" {
		logrus.Warn(filename + " (" + contentType + ")")
		return nil
	}

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)

	lineno := 0
	for scanner.Scan() {
		lineno++
		shannon(filename, scanner.Text(), lineno)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return nil
}

func main() {

	if len(os.Args) < 2 {
		logrus.Error("Please provide a directory as parameter")
		os.Exit(1)
	}

	root := os.Args[1]
	if root[0:1] != "/" {
		if cwd, err := os.Getwd(); err != nil {
			panic(err)
		} else {
			root = filepath.Join(cwd, root)
		}
	}

	logrus.Info("Directory: " + root)

	err := filepath.WalkDir(root, func(filename string, directory fs.DirEntry, e error) error {

		if isIgnored(filename) {
			return nil
		}

		if stat, err := os.Stat(filename); err == nil {
			if stat.IsDir() {
				return nil
			}
		} else {
			panic(err)
		}

		err := processFile(filename)
		if err != nil {
			panic(err)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

}
