// Command module-check compiles a consumer of the checkout's Go module archive.
package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
	modzip "golang.org/x/mod/zip"
)

type sourceFile struct{ root, name string }

func (f sourceFile) Path() string                 { return f.name }
func (f sourceFile) Lstat() (os.FileInfo, error)  { return os.Lstat(filepath.Join(f.root, f.name)) }
func (f sourceFile) Open() (io.ReadCloser, error) { return os.Open(filepath.Join(f.root, f.name)) }

func command(ctx context.Context, dir string, env []string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return out, nil
}

func archive(root string, names []string, manifest []byte) ([]byte, module.Version, error) {
	parsed, err := modfile.Parse("go.mod", manifest, nil)
	if err != nil {
		return nil, module.Version{}, err
	}
	if parsed.Module == nil {
		return nil, module.Version{}, fmt.Errorf("missing module declaration")
	}
	var files []modzip.File
	hash := sha256.New()
	for _, name := range names {
		if name == "" {
			continue
		}
		f := sourceFile{root: root, name: filepath.ToSlash(name)}
		info, err := f.Lstat()
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, module.Version{}, err
		}
		if !info.Mode().IsRegular() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return nil, module.Version{}, err
		}
		fmt.Fprintf(hash, "%s\x00%d\x00", name, len(data))
		hash.Write(data)
		files = append(files, f)
	}
	revision := fmt.Sprintf("%x", hash.Sum(nil))[:16]
	version := module.Version{Path: parsed.Module.Mod.Path, Version: "v0.0.0-check.g" + revision}
	var output bytes.Buffer
	if err := modzip.Create(&output, version, files); err != nil {
		return nil, version, err
	}
	return output.Bytes(), version, nil
}

func run(root string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	names, err := command(ctx, root, os.Environ(), "git", "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	if err != nil {
		return err
	}
	manifest, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return err
	}
	data, version, err := archive(root, strings.Split(string(names), "\x00"), manifest)
	if err != nil {
		return err
	}
	entries, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	for _, entry := range entries.File {
		relative := strings.TrimPrefix(entry.Name, version.Path+"@"+version.Version+"/")
		if strings.HasPrefix(relative, "examples/") || strings.HasPrefix(relative, "scripts/") {
			return fmt.Errorf("nested module leaked into archive: %s", relative)
		}
	}
	scratch, err := os.MkdirTemp("", "ownframe-consumer-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(scratch)
	proxy := filepath.Join(scratch, "proxy", version.Path, "@v")
	if err := os.MkdirAll(proxy, 0o700); err != nil {
		return err
	}
	info, _ := json.Marshal(map[string]any{"Version": version.Version, "Time": time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})
	for suffix, body := range map[string][]byte{".zip": data, ".mod": manifest, ".info": info} {
		if err := os.WriteFile(filepath.Join(proxy, version.Version+suffix), body, 0o600); err != nil {
			return err
		}
	}
	consumer := filepath.Join(scratch, "consumer")
	if err := os.Mkdir(consumer, 0o700); err != nil {
		return err
	}
	parsed, err := modfile.Parse("go.mod", manifest, nil)
	if err != nil {
		return err
	}
	moduleText := fmt.Sprintf("module consumer.example/app\n\ngo %s\n\nrequire %s %s\n", parsed.Go.Version, version.Path, version.Version)
	if err := os.WriteFile(filepath.Join(consumer, "go.mod"), []byte(moduleText), 0o600); err != nil {
		return err
	}
	source := fmt.Sprintf(`package app
import (
    "context"
    frame %q
)
func App(ctx context.Context) error {
    p,err:=frame.NewWithOptions(frame.Config{HTML:"<div>{{.}}</div>",Width:320,Height:200},frame.WithPerf(true))
    if err!=nil { return err }
    p.SetData("consumer")
    p.Handle(frame.Handlers{Submit:func(ctx context.Context) error { p.SetData("saved"); return p.Redraw(ctx) }})
    p.SetTick(func(context.Context) error { return nil })
    return p.Redraw(ctx)
}
`, version.Path)
	if err := os.WriteFile(filepath.Join(consumer, "app.go"), []byte(source), 0o600); err != nil {
		return err
	}
	proxyURL := "file://" + filepath.ToSlash(filepath.Join(scratch, "proxy"))
	env := append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local", "GONOSUMDB="+version.Path, "GOPRIVATE=", "GOPROXY="+proxyURL+",https://proxy.golang.org")
	if _, err := command(ctx, consumer, env, "go", "mod", "tidy"); err != nil {
		return err
	}
	if _, err := command(ctx, consumer, env, "go", "vet", "-mod=readonly", "-p", "1", "./..."); err != nil {
		return err
	}
	graph, err := command(ctx, consumer, env, "go", "list", "-m", "-json", version.Path)
	if err != nil {
		return err
	}
	var resolved struct {
		Version string
		Replace any
	}
	if err := json.Unmarshal(graph, &resolved); err != nil {
		return err
	}
	if resolved.Version != version.Version || resolved.Replace != nil {
		return fmt.Errorf("consumer did not resolve the archived module: %s", graph)
	}
	fmt.Printf("module-check: %d archive entries; isolated consumer compiled %s with no workspace or replace\n", len(entries.File), version.Version)
	return nil
}

func validateExamples(tag string, data []byte) error {
	if !strings.HasPrefix(tag, "examples/") || !semver.IsValid(strings.TrimPrefix(tag, "examples/")) {
		return fmt.Errorf("invalid examples tag: %s", tag)
	}
	parsed, err := modfile.Parse("examples/go.mod", data, nil)
	if err != nil {
		return err
	}
	for _, replacement := range parsed.Replace {
		if replacement.New.Version == "" {
			return fmt.Errorf("tagged examples cannot use filesystem replace: %s", replacement.New.Path)
		}
	}
	for _, requirement := range parsed.Require {
		if requirement.Mod.Path == "github.com/chinmay-sawant/ownframe" && semver.IsValid(requirement.Mod.Version) {
			return nil
		}
	}
	return fmt.Errorf("examples must require a published library version")
}

func main() {
	root := flag.String("root", "..", "repository root")
	examplesTag := flag.String("examples-tag", "", "validate a candidate examples/vX.Y.Z tag")
	flag.Parse()
	if *examplesTag != "" {
		data, err := os.ReadFile(filepath.Join(*root, "examples/go.mod"))
		if err == nil {
			err = validateExamples(*examplesTag, data)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if _, err := command(ctx, filepath.Join(*root, "examples"), append(os.Environ(), "GOWORK=off"), "go", "vet", "-mod=readonly", "-p", "1", "./..."); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if err := run(*root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
