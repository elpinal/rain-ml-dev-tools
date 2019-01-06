package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}

func run() error {
	if err := update(); err != nil {
		return errors.Wrap(err, "updating")
	}
	return test(os.Getenv("HOME"))
}

func update() error {
	cmd := exec.Command("rainy", "update", "master")
	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "RUST_LOG=rainy")
	return errors.Wrap(cmd.Run(), "rainy")
}

func test(home string) error {
	bin := filepath.Join(home, ".rain", "bin")
	f, err := os.Open("testdata")
	if err != nil {
		return errors.Wrapf(err, "looking for %q directory", "testdata")
	}
	files, err := f.Readdirnames(0)
	if err != nil {
		return errors.Wrapf(err, "reading entries in %q directory", "testdata")
	}
	for _, file := range files {
		i := strings.LastIndexByte(file, '.')
		if i < 0 {
			continue
		}
		if !strings.HasSuffix(file, ".rml") {
			continue
		}
		err := testFile(
			filepath.Join("testdata", file),
			filepath.Join("testdata", file[:i]+".rvm"),
			filepath.Join("testdata", file[:i]+".expected"),
			filepath.Join(bin, "rain-ml"),
			filepath.Join(bin, "rain-vm"),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func testFile(srcFile, asmFile, expectFile, ml, vm string) error {
	cmd := exec.Command(ml, "build", srcFile, asmFile)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "executing rain-ml")
	}

	cmd = exec.Command(vm, asmFile)
	bs1, err := cmd.Output()
	if err != nil {
		return errors.Wrap(err, "executing rain-vm")
	}

	bs2, err := ioutil.ReadFile(expectFile)
	if err != nil {
		return errors.Wrap(err, `reading ".expected" file`)
	}

	got := string(bs1)
	expected := string(bs2)

	if got != expected {
		return fmt.Errorf("%s: got %q, but expected %q", srcFile, got, expected)
	}
	return nil
}
