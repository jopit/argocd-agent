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

package fixture

import (
	"context"

	argoapp "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BaseSuite struct {
	suite.Suite
	Ctx                   context.Context
	PrincipalClient       KubeClient
	ManagedAgentClient    KubeClient
	AutonomousAgentClient KubeClient
}

func (suite *BaseSuite) SetupSuite() {
	requires := suite.Require()

	suite.Ctx = context.Background()

	config, err := GetSystemKubeConfig("vcluster-control-plane")
	requires.Nil(err)
	suite.PrincipalClient, err = NewKubeClient(config)
	requires.Nil(err)

	config, err = GetSystemKubeConfig("vcluster-agent-managed")
	requires.Nil(err)
	suite.ManagedAgentClient, err = NewKubeClient(config)
	requires.Nil(err)

	config, err = GetSystemKubeConfig("vcluster-agent-autonomous")
	requires.Nil(err)
	suite.AutonomousAgentClient, err = NewKubeClient(config)
	requires.Nil(err)
}

func (suite *BaseSuite) SetupTest() {
	err := CleanUp(suite.Ctx, suite.PrincipalClient, suite.ManagedAgentClient, suite.AutonomousAgentClient)
	suite.Assert().Nil(err)
}

func (suite *BaseSuite) TearDownTest() {
	err := CleanUp(suite.Ctx, suite.PrincipalClient, suite.ManagedAgentClient, suite.AutonomousAgentClient)
	suite.Assert().Nil(err)
}

func CleanUp(ctx context.Context, principalClient KubeClient, managedAgentClient KubeClient, autonomousAgentClient KubeClient) error {

	var list argoapp.ApplicationList
	var err error

	// Delete all managed applications from the principal
	list = argoapp.ApplicationList{}
	err = principalClient.List(ctx, "agent-managed", &list, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, app := range list.Items {
		err = principalClient.Delete(ctx, &app, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	// Delete all applications from the autonomous agent
	list = argoapp.ApplicationList{}
	err = autonomousAgentClient.List(ctx, "argocd", &list, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, app := range list.Items {
		err = autonomousAgentClient.Delete(ctx, &app, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	// Delete any remaining managed applications left on the managed agent
	list = argoapp.ApplicationList{}
	err = managedAgentClient.List(ctx, "agent-managed", &list, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, app := range list.Items {
		err = managedAgentClient.Delete(ctx, &app, metav1.DeleteOptions{})
		// We may get NotFound errors because the principal may
		// have deleted the app after we got the list of apps
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	// Delete any remaining autonomous applications left on the principal
	list = argoapp.ApplicationList{}
	err = principalClient.List(ctx, "agent-autonomous", &list, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, app := range list.Items {
		err = principalClient.Delete(ctx, &app, metav1.DeleteOptions{})
		// We may get NotFound errors because the autonomous agent may
		// have deleted the app after we got the list of apps
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	return nil
}
