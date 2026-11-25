// Copyright 2026 The go-commons Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package podfiles reads files out of a running container.
//
// A tool running in a pod writes its output to the pod's filesystem, which
// disappears with the pod. This execs into the container and pulls those
// files back, so a caller can do something durable with them.
package podfiles

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// Target identifies the container to read from.
type Target struct {
	Namespace string
	Pod       string
	Container string
}

// File is one file read out of a container.
type File struct {
	// Path is the absolute path inside the container.
	Path string
	// Content is the file's bytes, exactly as they were on disk.
	Content []byte
}

// Scan returns every file under dir, recursively, whose name matches one of
// the given shell patterns. With no patterns it returns every file.
//
// The container must still be running: a completed pod has no process to
// exec into, so collect before the workload's pod exits.
//
// Content is transferred base64 encoded, one file per line, so a single exec
// fetches the whole tree and arbitrary bytes survive intact.
func Scan(ctx context.Context, restConfig *rest.Config, target Target, dir string, patterns ...string) ([]File, error) {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	stdout, stderr, err := exec(ctx, clientset, restConfig, target, listCommand(dir, patterns))
	if err != nil {
		return nil, fmt.Errorf("reading %s from %s/%s: %w (%s)",
			dir, target.Namespace, target.Pod, err, strings.TrimSpace(stderr))
	}
	return parse(stdout)
}

// listCommand finds matching files and prints "<path> <base64>" per line.
func listCommand(dir string, patterns []string) string {
	match := ""
	if len(patterns) > 0 {
		quoted := make([]string, 0, len(patterns))
		for _, pattern := range patterns {
			quoted = append(quoted, fmt.Sprintf("-name '%s'", pattern))
		}
		match = " \\( " + strings.Join(quoted, " -o ") + " \\)"
	}
	// base64 is piped through tr rather than using -w0, which is GNU only.
	return fmt.Sprintf(
		"find '%s' -type f%s 2>/dev/null | sort | "+
			"while IFS= read -r f; do printf '%%s %%s\\n' \"$f\" \"$(base64 < \"$f\" | tr -d '\\n')\"; done",
		dir, match)
}

func parse(stdout string) ([]File, error) {
	var files []File
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		path, encoded, found := strings.Cut(line, " ")
		if !found {
			// A file that exists but is empty encodes to nothing.
			files = append(files, File{Path: path})
			continue
		}
		content, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return files, fmt.Errorf("decoding %s: %w", path, err)
		}
		files = append(files, File{Path: path, Content: content})
	}
	return files, nil
}

func exec(ctx context.Context, clientset kubernetes.Interface, restConfig *rest.Config,
	target Target, command string) (string, string, error) {
	request := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(target.Pod).
		Namespace(target.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: target.Container,
			Command:   []string{"/bin/sh", "-c", command},
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(restConfig, "POST", request.URL())
	if err != nil {
		return "", "", err
	}
	var stdout, stderr bytes.Buffer
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{Stdout: &stdout, Stderr: &stderr})
	return stdout.String(), stderr.String(), err
}
