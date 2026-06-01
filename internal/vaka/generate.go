package vaka

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/infrasecture/myHarness/internal/profiles"
)

type Config struct {
	Profile     profiles.Profile
	ServiceName string
}

type TemplateData struct {
	ServiceName  string
	HarnessName  string
	AllowedHosts []string
}

func Write(w io.Writer, cfg Config) error {
	tmplPath := cfg.Profile.VakaPolicyTemplate
	if tmplPath == "" {
		return fmt.Errorf("profile %q does not declare vakaPolicyTemplate", cfg.Profile.Name)
	}
	raw, err := readTemplate(tmplPath)
	if err != nil {
		return fmt.Errorf("read Vaka policy template %s: %w", tmplPath, err)
	}
	service := cfg.ServiceName
	if service == "" {
		service = "myharness"
	}
	hosts := append([]string(nil), cfg.Profile.VakaAllowedHosts...)
	sort.Strings(hosts)
	data := TemplateData{
		ServiceName:  service,
		HarnessName:  cfg.Profile.Name,
		AllowedHosts: hosts,
	}
	tmpl, err := template.New(tmplPath).Option("missingkey=error").Parse(string(raw))
	if err != nil {
		return fmt.Errorf("parse Vaka policy template %s: %w", tmplPath, err)
	}
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		return fmt.Errorf("render Vaka policy template %s: %w", tmplPath, err)
	}
	_, err = w.Write(rendered.Bytes())
	return err
}

func readTemplate(path string) ([]byte, error) {
	if filepath.IsAbs(path) {
		return os.ReadFile(path)
	}
	if data, err := os.ReadFile(path); err == nil {
		return data, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	for {
		candidate := filepath.Join(cwd, path)
		if data, err := os.ReadFile(candidate); err == nil {
			return data, nil
		}
		next := filepath.Dir(cwd)
		if next == cwd {
			break
		}
		cwd = next
	}
	return nil, os.ErrNotExist
}
