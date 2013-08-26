package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"code.google.com/p/goauth2/oauth"
	"github.com/google/go-github/github"
)

var (
	name        = flag.String("name", "", "Name of STDIN file")
	description = flag.String("desc", "", "Gist description")
	public      = flag.Bool("public", false, "Public (defaults to private)")
	token       = flag.String("token", os.Getenv("GITHUB_TOKEN"), "Github token, empty is anonymous")
)

func main() {
	flag.Parse()

	var stdin []byte
	if !isTerminal(0) {
		stdin, _ = ioutil.ReadAll(os.Stdin)
	}
	if len(stdin) == 0 && flag.NArg() == 0 {
		log.Fatalln("No stdin or files provided")
	}

	gist := &github.Gist{Public: public, Files: make(map[github.GistFilename]github.GistFile)}
	if *description != "" {
		gist.Description = description
	}

	if len(stdin) > 0 {
		content := string(stdin)
		gist.Files[github.GistFilename(*name)] = github.GistFile{Content: &content}
	}

	for _, file := range flag.Args() {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("Error reading %s: %s", file, err)
		}
		content := string(data)
		gist.Files[github.GistFilename(filepath.Base(file))] = github.GistFile{Content: &content}
	}

	client := github.NewClient(nil)
	if *token != "" {
		t := &oauth.Transport{Token: &oauth.Token{AccessToken: *token}}
		client = github.NewClient(t.Client())
	}

	res, _, err := client.Gists.Create(gist)
	if err != nil {
		log.Fatalln("Unable to create gist:", err)
	}

	fmt.Println(*res.HTMLURL)
}

func isTerminal(fd int) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlReadTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}
