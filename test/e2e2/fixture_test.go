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
	"encoding/json"
	"testing"

	"github.com/argoproj-labs/argocd-agent/test/e2e2/fixture"
	argoapp "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

type FixtureTestSuite struct {
	suite.Suite
}

func (suite *FixtureTestSuite) SetupSuite() {
	requires := suite.Require()

	ctx := context.Background()

	config, err := fixture.GetSystemKubeConfig("")
	requires.Nil(err)
	kclient, err := fixture.NewKubeClient(config)
	requires.Nil(err)

	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-argocd-agent",
		},
	}
	err = kclient.Create(ctx, &namespace, metav1.CreateOptions{})
	requires.Nil(err)

	app := argoapp.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guestbook",
			Namespace: "test-argocd-agent",
		},
		Spec: argoapp.ApplicationSpec{
			Source: &argoapp.ApplicationSource{
				RepoURL:        "https://github.com/argoproj/argocd-example-apps",
				TargetRevision: "HEAD",
				Path:           "kustomize-guestbook",
			},
			Destination: argoapp.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: "foo",
			},
		},
	}
	err = kclient.Create(ctx, &app, metav1.CreateOptions{})
	requires.Nil(err)

	app = argoapp.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guestbook1",
			Namespace: "test-argocd-agent",
		},
		Spec: argoapp.ApplicationSpec{
			Source: &argoapp.ApplicationSource{
				RepoURL:        "https://github.com/argoproj/argocd-example-apps",
				TargetRevision: "HEAD",
				Path:           "kustomize-guestbook",
			},
			Destination: argoapp.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: "foo",
			},
		},
	}
	err = kclient.Create(ctx, &app, metav1.CreateOptions{})
	requires.Nil(err)

	app = argoapp.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guestbook2",
			Namespace: "test-argocd-agent",
		},
		Spec: argoapp.ApplicationSpec{
			Source: &argoapp.ApplicationSource{
				RepoURL:        "https://github.com/argoproj/argocd-example-apps",
				TargetRevision: "HEAD",
				Path:           "kustomize-guestbook",
			},
			Destination: argoapp.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: "foo",
			},
		},
	}
	err = kclient.Create(ctx, &app, metav1.CreateOptions{})
	requires.Nil(err)
}

func (suite *FixtureTestSuite) TearDownSuite() {
	requires := suite.Require()

	ctx := context.Background()

	config, err := fixture.GetSystemKubeConfig("")
	requires.Nil(err)
	kclient, err := fixture.NewKubeClient(config)
	requires.Nil(err)

	app := argoapp.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guestbook",
			Namespace: "test-argocd-agent",
		},
	}
	err = kclient.Delete(ctx, &app, metav1.DeleteOptions{})
	requires.Nil(err)

	app = argoapp.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guestbook1",
			Namespace: "test-argocd-agent",
		},
	}
	err = kclient.Delete(ctx, &app, metav1.DeleteOptions{})
	requires.Nil(err)

	app = argoapp.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "guestbook2",
			Namespace: "test-argocd-agent",
		},
	}
	err = kclient.Delete(ctx, &app, metav1.DeleteOptions{})
	requires.Nil(err)

	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-argocd-agent",
		},
	}
	err = kclient.Delete(ctx, &namespace, metav1.DeleteOptions{})
	requires.Nil(err)
}

func (suite *FixtureTestSuite) Test_Sanity() {
	requires := suite.Require()
	requires.True(true)
}

func (suite *FixtureTestSuite) Test_KubeConfig() {
	requires := suite.Require()

	config, err := fixture.GetSystemKubeConfig("")
	requires.Nil(err)
	requires.NotNil(config)

	kclient, err := kubernetes.NewForConfig(config)
	requires.Nil(err)
	requires.NotNil(kclient)
}

func (suite *FixtureTestSuite) Test_Get_Application_Via_Dynamic() {
	requires := suite.Require()

	config, err := fixture.GetSystemKubeConfig("")
	requires.Nil(err)

	dclient, err := dynamic.NewForConfig(config)
	requires.Nil(err)

	appResource := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}

	ctx := context.Background()

	unstructuredApp, err := dclient.Resource(appResource).Namespace("test-argocd-agent").Get(ctx, "guestbook", metav1.GetOptions{})
	requires.Nil(err)

	b, err := unstructuredApp.MarshalJSON()
	requires.Nil(err)
	requires.NotNil(b)
	app := argoapp.Application{}
	err = json.Unmarshal(b, &app)
	requires.Nil(err)
	requires.Equal("Application", app.Kind)
}

func (suite *FixtureTestSuite) Test_List_Application_Via_Dynamic() {
	requires := suite.Require()

	config, err := fixture.GetSystemKubeConfig("")
	requires.Nil(err)

	dclient, err := dynamic.NewForConfig(config)
	requires.Nil(err)

	appResource := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}

	ctx := context.Background()

	ulist, err := dclient.Resource(appResource).Namespace("test-argocd-agent").List(ctx, metav1.ListOptions{})
	requires.Nil(err)

	b, err := ulist.MarshalJSON()
	requires.Nil(err)
	requires.NotNil(b)

	list := argoapp.ApplicationList{}
	err = json.Unmarshal(b, &list)
	requires.Nil(err)
	requires.NotEmpty(list)
}

func (suite *FixtureTestSuite) Test_Get_Application_Via_RESTMapper() {
	requires := suite.Require()

	ctx := context.Background()

	config, err := fixture.GetSystemKubeConfig("")
	requires.Nil(err)

	scheme := runtime.NewScheme()
	argoapp.AddToScheme(scheme)

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	requires.Nil(err)
	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	requires.Nil(err)

	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	app := argoapp.Application{}

	gvks, unversioned, err := scheme.ObjectKinds(&app)
	requires.Nil(err)
	requires.False(unversioned)
	requires.NotEmpty(gvks)

	mapping, err := mapper.RESTMapping(gvks[0].GroupKind())
	requires.Nil(err)

	requires.Equal("argoproj.io", mapping.Resource.Group)
	requires.Equal("v1alpha1", mapping.Resource.Version)
	requires.Equal("applications", mapping.Resource.Resource)

	dclient, err := dynamic.NewForConfig(config)
	requires.Nil(err)

	unstructuredApp, err := dclient.Resource(mapping.Resource).Namespace("test-argocd-agent").Get(ctx, "guestbook", metav1.GetOptions{})
	requires.Nil(err)

	b, err := unstructuredApp.MarshalJSON()
	requires.Nil(err)
	requires.NotEmpty(b)
	app = argoapp.Application{}
	err = json.Unmarshal(b, &app)
	requires.Nil(err)
	requires.Equal("Application", app.Kind)
}

func (suite *FixtureTestSuite) Test_Get() {
	requires := suite.Require()

	ctx := context.Background()

	config, err := fixture.GetSystemKubeConfig("")
	requires.Nil(err)

	kclient, err := fixture.NewKubeClient(config)
	requires.Nil(err)

	app := argoapp.Application{}
	err = kclient.Get(ctx, types.NamespacedName{Namespace: "test-argocd-agent", Name: "guestbook"}, &app, metav1.GetOptions{})
	requires.Nil(err)
	requires.Equal("guestbook", app.Name)
	requires.Equal("Application", app.Kind)

	app = argoapp.Application{}
	err = kclient.Get(ctx, types.NamespacedName{Namespace: "test-argocd-agent", Name: "guestbook"}, &app, metav1.GetOptions{})
	requires.Nil(err)
	requires.Equal("guestbook", app.Name)
	requires.Equal("Application", app.Kind)

	ns := corev1.Namespace{}
	err = kclient.Get(ctx, types.NamespacedName{Namespace: "", Name: "test-argocd-agent"}, &ns, metav1.GetOptions{})
	requires.Nil(err)
	requires.Equal("test-argocd-agent", ns.Name)
}

func (suite *FixtureTestSuite) Test_Create_Get_Delete_Application() {
	requires := suite.Require()

	ctx := context.Background()

	config, err := fixture.GetSystemKubeConfig("")
	requires.Nil(err)

	kclient, err := fixture.NewKubeClient(config)
	requires.Nil(err)

	app := argoapp.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "test-argocd-agent",
		},
		Spec: argoapp.ApplicationSpec{
			Destination: argoapp.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: "foo",
			},
		},
	}

	err = kclient.Create(ctx, &app, metav1.CreateOptions{})
	requires.Nil(err)
	requires.Equal("foo", app.Name)
	requires.Equal("Application", app.Kind)
	requires.NotEmpty(app.UID)

	app = argoapp.Application{}
	err = kclient.Get(ctx, types.NamespacedName{Namespace: "test-argocd-agent", Name: "foo"}, &app, metav1.GetOptions{})
	requires.Nil(err)
	requires.Equal("foo", app.Name)
	requires.Equal("Application", app.Kind)
	requires.NotEmpty(app.UID)

	app = argoapp.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "test-argocd-agent",
		},
	}
	err = kclient.Delete(ctx, &app, metav1.DeleteOptions{})
	requires.Nil(err)

	app = argoapp.Application{}
	err = kclient.Get(ctx, types.NamespacedName{Namespace: "test-argocd-agent", Name: "foo"}, &app, metav1.GetOptions{})
	requires.NotNil(err)
	requires.True(errors.IsNotFound(err))
}

func (suite *FixtureTestSuite) Test_Update_Application() {
	requires := suite.Require()

	ctx := context.Background()

	config, err := fixture.GetSystemKubeConfig("")
	requires.Nil(err)

	kclient, err := fixture.NewKubeClient(config)
	requires.Nil(err)

	app := argoapp.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "test-argocd-agent",
		},
		Spec: argoapp.ApplicationSpec{
			Source: &argoapp.ApplicationSource{
				RepoURL:        "https://github.com/argoproj/argocd-example-apps",
				TargetRevision: "HEAD",
				Path:           "kustomize-guestbook",
			},
			Destination: argoapp.ApplicationDestination{
				Server:    "https://kubernetes.default.svc",
				Namespace: "foo",
			},
		},
	}
	err = kclient.Create(ctx, &app, metav1.CreateOptions{})
	requires.Nil(err)
	requires.Equal("foo", app.Name)
	requires.Equal("Application", app.Kind)
	requires.NotEmpty(app.UID)

	app.Spec.Source.TargetRevision = "TAIL"
	err = kclient.Update(ctx, &app, metav1.UpdateOptions{})
	requires.Nil(err)
	requires.Equal("TAIL", app.Spec.Source.TargetRevision)

	app = argoapp.Application{}
	err = kclient.Get(ctx, types.NamespacedName{Namespace: "test-argocd-agent", Name: "foo"}, &app, metav1.GetOptions{})
	requires.Nil(err)
	requires.Equal("TAIL", app.Spec.Source.TargetRevision)

	err = kclient.Delete(ctx, &app, metav1.DeleteOptions{})
	requires.Nil(err)
}

func (suite *FixtureTestSuite) Test_List_Applications() {
	requires := suite.Require()

	ctx := context.Background()

	config, err := fixture.GetSystemKubeConfig("")
	requires.Nil(err)

	kclient, err := fixture.NewKubeClient(config)
	requires.Nil(err)

	list := argoapp.ApplicationList{}
	err = kclient.List(ctx, "test-argocd-agent", &list, metav1.ListOptions{})
	requires.Nil(err)
	requires.Len(list.Items, 3)
}

func TestFixtureTestSuite(t *testing.T) {
	suite.Run(t, new(FixtureTestSuite))
}
