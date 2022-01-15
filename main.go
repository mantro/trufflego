package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
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

var opts struct {
	Threshold           float64 `short:"t" long:"threshold" default:"4.8" description:"Default 4.8 (higher->lesser detections)"`
	MinimumStringLength int     `short:"m" long:"minimum" default:"12" description:"Minimum length of detected passwords"`
	Args                struct {
		Directory string `positional-arg-name:"directory" required:"yes" description:"Directory"`
	} `positional-args:"yes"`
}

var foundColor = color.New(color.FgHiYellow).Add(color.BgBlack).Add(color.Bold)

func isIgnored(line string) bool {

	ignored := []string{
		".lock",
		"package-lock.json",
		"/Migrations/",
		".git/",
		"node_modules/",
		".blackbox/",
		".gpg",
		"go.sum",
		".svg",
		".css",
		".deps.json",
		"project.assets.json",
		"project.nuget.cache",
		"/obj/",
		"/bin",
		"pnpm-lock.yaml",
		".sln.DotSettings",
		"project.pbxproj",
		"yarn-error.log",
		".sln",
		".csproj",
	}

	for _, needle := range ignored {
		if strings.Contains(line, needle) {
			return true
		}
	}

	return false
}

func shannon(filename string, line string, lineno int) {

	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=#!§$%&()[]|{}*-_.:,;'\"?"

	// split line into words
	words := strings.Fields(line)

	// iterate the words
	for _, word := range words {

		var stringSet []string
		count := 0
		letters := ""

		// check for consecutive strings with >16 length that consist solely of
		// characters in the above charset
		for _, char := range word {

			if strings.ContainsRune(charset, char) {
				// character is from the charset, add it.
				letters = letters + string(char)
				count++
			} else {
				// if the word has a minimum length, add it.
				if count >= opts.MinimumStringLength {
					stringSet = append(stringSet, letters)
				}

				// one way or another reset the state machine
				letters = ""
				count = 0
			}
		}

		// we might have a leftover interesting word in the buffer
		if count >= opts.MinimumStringLength {
			stringSet = append(stringSet, letters)
		}

		// stringSet now contains possible words that match the above criteria
		// now we have to calculate the entropy of these strings
		for _, token := range stringSet {

			// inspired by http://blog.dkbza.org/2007/05/scanning-data-for-entropy-anomalies.html
			entropy := float64(0)
			strCount := len(token)

			for _, char := range charset {
				cnt := strings.Count(token, string(char))
				temp := float64(cnt) / float64(strCount)
				if temp > 0 {
					entropy = entropy + ((0 - temp) * math.Log2(temp))
				}
			}

			// entropy now contains the shannon entropy. Higher is more entropic.
			// around 5 seems to be a good value, but this should be configurable
			if entropy > opts.Threshold {
				// this might be a password, log it out
				logrus.Warnf("%s:%d (Score: %s)", filename, lineno, color.HiGreenString(" (Score: %f)", entropy))

				// try to color highlight the token
				output := strings.Replace(line, token, foundColor.Sprintf(token), 1)

				// truncate long line
				if len(output) > 1000 {
					output = output[0:1000] + "... <truncated>"
				}

				fmt.Println(output)
				fmt.Println()
			}
		}
	}
}

// check if an array of strings contains a specific string
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func processFile(filename string) error {

	supported := []string{
		"text/plain; charset=utf-8",
		"text/xml; charset=utf-8",
		"text/html; charset=utf-8",
	}

	knownUnsupported := []string{
		"application/octet-stream",
		"application/zip",
		"application/pdf",
		"image/jpeg",
		"image/png",
		"image/webp",
		"font/woff",
		"font/woff2",
		"font/ttf",
		"image/x-icon",
		"application/vnd.ms-fontobject",
		"font/otf",
		"application/x-gzip",
	}

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

	// read first 512 bytes (or less if file is smaller)
	buffer := make([]byte, 512)

	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}

	// do content detection
	contentType := http.DetectContentType(buffer[:n])

	// check if it is not whitelisted
	if !contains(supported, contentType) {

		// if it is not blacklisted yet, log it out (there is maybe some content types that we still need to sort it)
		if !contains(knownUnsupported, contentType) {
			logrus.Warn(filename + " (" + contentType + ")")
		}

		return nil
	}

	// reset file pointer to 0
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)

	lineno := 0

	// start iterating the file line by line
	for scanner.Scan() {
		lineno++
		// and check the line
		shannon(filename, scanner.Text(), lineno)
	}

	if err := scanner.Err(); err != nil {
		logrus.Error("Error parsing: " + filename)
		logrus.Error(err)
		fmt.Println()
	}

	return nil
}

func main() {

	if _, err := flags.ParseArgs(&opts, os.Args[1:]); err != nil {
		os.Exit(1)
	}

	root := opts.Args.Directory

	if root[0:1] != "/" {
		if cwd, err := os.Getwd(); err != nil {
			panic(err)
		} else {
			root = filepath.Join(cwd, root)
		}
	}

	logrus.Info("Directory: " + root)

	err := filepath.WalkDir(root, func(filename string, directory fs.DirEntry, e error) error {

		// certain files can be safely ignored
		if isIgnored(filename) {
			return nil
		}

		// iteration returns both directories and files, so sort out the files
		if stat, err := os.Stat(filename); err == nil {
			if stat.IsDir() {
				return nil
			}
		} else {
			panic(err)
		}

		// analyze the file
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
