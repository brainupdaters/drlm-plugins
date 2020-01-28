// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

var fs = afero.NewOsFs()

// Config is the configuration that is going to be passed in ALWAYS through the -config flag as a string
// This configuration is going to be stored in the Core and it will be available for editing, so it can be
// in any format
type Config struct {
	Files []string `json:"files,omitempty"`
}

func main() {
	var strCfg string
	flag.StringVar(&strCfg, "config", "", "")
	// target is the directory where the backup / output HAS to be stored. It's always passed
	var target string
	flag.StringVar(&target, "target", "", "")
	flag.Parse()

	var cfg Config
	if err := json.Unmarshal([]byte(strCfg), &cfg); err != nil {
		fmt.Printf("parse configuration: %v\n", err)
		os.Exit(1)
	}

	for _, f := range cfg.Files {
		// Copy the file
		if err := cp(f, target); err != nil {
			fmt.Printf("copy the file: %v", err)
			os.Exit(1)
		}
	}
}

func cp(src, dst string) error {
	stat, err := fs.Stat(src)
	if err != nil {
		return err
	}

	dstPath := filepath.Join(dst, src)

	if stat.IsDir() {
		subFiles, err := afero.ReadDir(fs, src)
		if err != nil {
			return fmt.Errorf("list '%s' content: %v", src, err)
		}

		for _, f := range subFiles {
			if err := cp(filepath.Join(src, f.Name()), dst); err != nil {
				return err
			}
		}

	} else {
		if err := fs.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			fmt.Printf("create directory tree: %v\n", err)
			os.Exit(1)
		}

		srcF, err := fs.Open(src)
		if err != nil {
			return err
		}
		defer srcF.Close()

		dstF, err := fs.Create(dstPath)
		if err != nil {
			return fmt.Errorf("create '%s': %v", dstPath, err)
		}
		defer dstF.Close()

		if _, err := io.Copy(dstF, srcF); err != nil {
			return fmt.Errorf("copy '%s': %v", dstPath, err)
		}

		if err := dstF.Sync(); err != nil {
			return fmt.Errorf("syncing '%s': %v", dstPath, err)
		}

	}

	return nil
}
