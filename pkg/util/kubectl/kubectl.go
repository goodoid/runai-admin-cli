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

package kubectl

import (
	"fmt"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

var KubeConfig string

var kubectlCmd = []string{"kubectl"}

func Apply(pathToFile string) error {
	args := []string{"apply", "-f", pathToFile}
	out, err := kubectl(args)

	log.Debugf("%s\n", out)
	if err != nil {
		log.Errorf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}

	return err
}

/**
* dry-run creating kubernetes App Info for delete in future
* Exec /usr/local/bin/kubectl, [create --dry-run -f /tmp/values313606961 --namespace default]
**/

func kubectl(args []string) (string, error) {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return "", err
	}

	// 1. prepare the arguments
	// args := []string{"create", "configmap", name, "--namespace", namespace, fmt.Sprintf("--from-file=%s=%s", name, configFileName)}
	log.Debugf("Exec %s, %v", binary, args)

	env := os.Environ()
	if KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", KubeConfig))
	}

	// return syscall.Exec(cmd, args, env)
	// 2. execute the command
	cmd := exec.Command(binary, args...)
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf(string(output))
	} else {
		return string(output), nil
	}
}
