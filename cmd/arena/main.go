// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/kubeflow/arena/cmd/arena/commands"
	log "github.com/sirupsen/logrus"
)

func main() {

	if isPProfEnabled() {
		cpuf, err := os.Create("/tmp/cpu_profile")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(cpuf)
		log.Infof("Dump cpu profile file into /tmp/cpu_profile")
		defer pprof.StopCPUProfile()
	}

	if err := commands.NewCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func isPProfEnabled() (enable bool) {
	for _, arg := range os.Args {
		if arg == "--pprof" {
			enable = true
			break
		}
	}

	return
}
