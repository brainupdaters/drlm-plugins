// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"encoding/json"
	"flag"
	"log"
	"os/exec"

	"github.com/brainupdaters/drlm-agent/cfg"
	"github.com/brainupdaters/drlm-common/pkg/fs"
	"github.com/brainupdaters/drlm-common/pkg/minio"
	sdk "github.com/minio/minio-go/v6"
)

// Config is the configuration that is going to be passed in ALWAYS through the -config flag as a string
// This configuration is going to be stored in the Core and it will be available for editing, so it can be
// in any format
type Config struct {
	Files       []string `json:"files,omitempty"`
	Compression string   `json:"compression,omitempty"`
	Name        string   `json:"name,omitempty"`
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

	compression := ""
	ext := ".tar"
	switch jobCfg.Compression {
	case "", "none":

	case "gzip", "gz":
		compression = "z"
		ext += ".gz"

	case "bzip2", "bz2":
		compression = "j"
		ext += ".bz2"

	default:
		log.Fatalln("unsupported / unknown compression algorithm")
	}

	args := []string{compression + "cf", "-"}
	args = append(args, jobCfg.Files...)

	cmd := exec.Command("tar", args...)

	p, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("pipe tar output: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("start tar: %v", err)
	}

	if _, err := cli.PutObject(target, jobCfg.Name+ext, p, -1, sdk.PutObjectOptions{}); err != nil {
		log.Fatalf("upload to minio: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatalf("tar: %v", err)
	}
}
