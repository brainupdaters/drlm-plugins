// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/brainupdaters/drlm-agent/cfg"
	"github.com/brainupdaters/drlm-common/pkg/fs"
	"github.com/brainupdaters/drlm-common/pkg/minio"
	sdk "github.com/minio/minio-go/v6"
	"github.com/spf13/afero"
)

// Config is the configuration that is going to be passed in ALWAYS through the -config flag as a string
// This configuration is going to be stored in the Core and it will be available for editing, so it can be
// in any format
type Config struct {
	Files []string `json:"files,omitempty"`
}

var cli *sdk.Client
var target string

func main() {
	var strCfg string
	flag.StringVar(&strCfg, "config", "", "")
	// target is the directory where the backup / output HAS to be stored. It's always passed
	flag.StringVar(&target, "target", "", "")
	flag.Parse()

	var jobCfg Config
	var err error
	if err = json.Unmarshal([]byte(strCfg), &jobCfg); err != nil {
		log.Fatalf("parse configuration: %v", err)
	}

	fs.Init()
	cfg.Init("")

	cli, err = minio.NewSDK(
		cfg.Config.Minio.Host,
		cfg.Config.Minio.Port,
		cfg.Config.Minio.AccessKey,
		cfg.Config.Minio.SecretKey,
		cfg.Config.Minio.SSL,
		cfg.Config.Minio.CertPath,
	)
	if err != nil {
		log.Fatalln(err)
	}

	for _, f := range jobCfg.Files {
		// Copy the file
		if err := cp(f, target); err != nil {
			log.Fatalln(err)
		}
	}
}

func cp(src, dst string) error {
	stat, err := fs.FS.Stat(src)
	if err != nil {
		return err
	}

	if stat.IsDir() {
		subFiles, err := afero.ReadDir(fs.FS, src)
		if err != nil {
			return fmt.Errorf("list '%s' content: %v", src, err)
		}

		for _, f := range subFiles {
			if err := cp(filepath.Join(src, f.Name()), dst); err != nil {
				return err
			}
		}

	} else {
		name := src
		if filepath.IsAbs(src) {
			name = src[1:]
		}

		if _, err := cli.FPutObject(target, name, src, sdk.PutObjectOptions{}); err != nil {
			return fmt.Errorf("copy file '%s': %v", src, err)
		}
	}

	return nil
}
