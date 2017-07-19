/*
Copyright 2017 Heptio Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package discovery

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/golang/glog"
	"github.com/heptio/sonobuoy/pkg/config"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// PodLogsLocation is the location within the results tarball where pod
	// information is stored.
	PodLogsLocation = "podlogs"
)

// gatherPodLogs will loop through collecting pod logs and placing them into a directory tree
func gatherPodLogs(kubeClient kubernetes.Interface, ns string, opts metav1.ListOptions, cfg *config.Config) error {
	// 1 - Collect the list of pods
	podlist, err := kubeClient.CoreV1().Pods(ns).List(opts)
	if err != nil {
		return errors.WithStack(err)
	}

	glog.Info("Collecting Pod Logs...")

	// 2 - Foreach pod, dump each of its containers' logs in a tree in the following location:
	//   pods/:podname/logs/:containername.txt
	for _, pod := range podlist.Items {
		for _, container := range pod.Spec.Containers {
			body, err := kubeClient.CoreV1().Pods(ns).GetLogs(
				pod.Name,
				&v1.PodLogOptions{
					Container: container.Name,
				},
			).Do().Raw()

			if err != nil {
				return errors.WithStack(err)
			}

			outdir := path.Join(cfg.OutputDir(), PodLogsLocation, ns, pod.Name, "logs")
			if err = os.MkdirAll(outdir, 0755); err != nil {
				return errors.WithStack(err)
			}

			outfile := path.Join(outdir, container.Name) + ".txt"
			if err = ioutil.WriteFile(outfile, body, 0644); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}
