// Copyright 2024 The argocd-agent Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e2

import (
	"context"
	"testing"
	"time"

	"github.com/argoproj-labs/argocd-agent/test/e2e2/fixture"
	argoapp "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type BasicTestSuite struct {
	suite.Suite
}

func (suite *BasicTestSuite) Test_1_Agent_Managed() {
	requires := suite.Require()

	ctx := context.Background()

	config, err := fixture.GetSystemKubeConfig("vcluster-control-plane")
	requires.Nil(err)
	controlPlaneClient, err := fixture.NewKubeClient(config)
	requires.Nil(err)

	config, err = fixture.GetSystemKubeConfig("vcluster-agent-managed")
	requires.Nil(err)
	managedAgentClient, err := fixture.NewKubeClient(config)
	requires.Nil(err)

	// Create a managed application in the control-plane's cluster
	app := argoapp.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guestbook",
			Namespace: "agent-managed",
		},
		Spec: argoapp.ApplicationSpec{
			Project: "default",
			Source: &argoapp.ApplicationSource{
				RepoURL:        "https://github.com/argoproj/argocd-example-apps",
				TargetRevision: "HEAD",
				Path:           "kustomize-guestbook",
			},
			Destination: argoapp.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: "guestbook",
			},
			SyncPolicy: &argoapp.SyncPolicy{
				SyncOptions: argoapp.SyncOptions{
					"CreateNamespace=true",
				},
			},
		},
	}
	err = controlPlaneClient.Create(ctx, &app, metav1.CreateOptions{})
	requires.Nil(err)

	key := fixture.ToNamespacedName(&app)

	// Ensure the app has been pushed to the managed-agent
	requires.Eventually(func() bool {
		app := argoapp.Application{}
		err := managedAgentClient.Get(ctx, key, &app, metav1.GetOptions{})
		return err == nil
	}, 30*time.Second, 1*time.Second)

	// Delete the app from the control-plane
	err = controlPlaneClient.Delete(ctx, &app, metav1.DeleteOptions{})
	requires.Nil(err)

	// Ensure the app has been deleted from the managed-agent
	requires.Eventually(func() bool {
		app := argoapp.Application{}
		err := managedAgentClient.Get(ctx, key, &app, metav1.GetOptions{})
		return err != nil && errors.IsNotFound(err)
	}, 30*time.Second, 1*time.Second)
}

func (suite *BasicTestSuite) Test_2_Agent_Autonomous() {
	requires := suite.Require()

	ctx := context.Background()

	config, err := fixture.GetSystemKubeConfig("vcluster-control-plane")
	requires.Nil(err)
	controlPlaneClient, err := fixture.NewKubeClient(config)
	requires.Nil(err)

	config, err = fixture.GetSystemKubeConfig("vcluster-agent-autonomous")
	requires.Nil(err)
	autonomousAgentClient, err := fixture.NewKubeClient(config)
	requires.Nil(err)

	// Create an autonomous application on the autonomous-agent's cluster
	app := argoapp.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guestbook",
			Namespace: "argocd",
			Finalizers: []string{
				"resources-finalizer.argocd.argoproj.io",
			},
		},
		Spec: argoapp.ApplicationSpec{
			Project: "default",
			Source: &argoapp.ApplicationSource{
				RepoURL:        "https://github.com/argoproj/argocd-example-apps",
				TargetRevision: "HEAD",
				Path:           "kustomize-guestbook",
			},
			Destination: argoapp.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: "guestbook",
			},
			SyncPolicy: &argoapp.SyncPolicy{
				SyncOptions: argoapp.SyncOptions{
					"CreateNamespace=true",
				},
			},
		},
	}
	err = autonomousAgentClient.Create(ctx, &app, metav1.CreateOptions{})
	requires.Nil(err)

	key := types.NamespacedName{Name: app.Name, Namespace: "agent-autonomous"}

	// Ensure the app has been pushed to the control-plane
	requires.Eventually(func() bool {
		app := argoapp.Application{}
		err := controlPlaneClient.Get(ctx, key, &app, metav1.GetOptions{})
		return err == nil
	}, 30*time.Second, 1*time.Second)

	// Delete the app from the autonomous-agent
	err = autonomousAgentClient.Delete(ctx, &app, metav1.DeleteOptions{})
	requires.Nil(err)

	// Ensure the app has been deleted from the control-plane
	requires.Eventually(func() bool {
		app := argoapp.Application{}
		err := controlPlaneClient.Get(ctx, key, &app, metav1.GetOptions{})
		return err != nil && errors.IsNotFound(err)
	}, 30*time.Second, 1*time.Second)
}

func TestBasiceTestSuite(t *testing.T) {
	suite.Run(t, new(BasicTestSuite))
}
