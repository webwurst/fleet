package crd

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"

	fleet "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	"github.com/rancher/wrangler/pkg/crd"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

func Objects() (result []runtime.Object, err error) {
	for _, crdDef := range List() {
		crd, err := crdDef.ToCustomResourceDefinition()
		if err != nil {
			return nil, err
		}
		result = append(result, &crd)
	}
	return
}

func List() []crd.CRD {
	return []crd.CRD{
		newCRD(&fleet.Bundle{}, func(c crd.CRD) crd.CRD {
			return c.
				WithCategories("fleet").
				WithColumn("Clusters-Ready", ".status.summary.ready").
				WithColumn("Clusters-Desired", ".status.summary.desiredReady").
				WithColumn("Status", ".status.conditions[?(@.type==\"Ready\")].message")
		}),
		newCRD(&fleet.BundleDeployment{}, func(c crd.CRD) crd.CRD {
			return c.WithColumn("Status", ".status.conditions[?(@.type==\"Ready\")].message")
		}),
		newCRD(&fleet.ClusterGroup{}, func(c crd.CRD) crd.CRD {
			return c.
				WithCategories("fleet").
				WithColumn("Cluster-Count", ".status.clusterCount").
				WithColumn("NonReady-Clusters", ".status.nonReadyClusters").
				WithColumn("Bundles-Ready", ".status.summary.ready").
				WithColumn("Bundles-Desired", ".status.summary.desiredReady").
				WithColumn("Status", ".status.conditions[?(@.type==\"Ready\")].message")

		}),
		newCRD(&fleet.Cluster{}, func(c crd.CRD) crd.CRD {
			return c.
				WithColumn("Bundles-Ready", ".status.summary.ready").
				WithColumn("Bundles-Desired", ".status.summary.desiredReady").
				WithColumn("Status", ".status.conditions[?(@.type==\"Ready\")].message")
		}),
		newCRD(&fleet.ClusterGroupToken{}, func(c crd.CRD) crd.CRD {
			return c.
				WithColumn("Cluster-Group", ".spec.clusterGroupName").
				WithColumn("Secret-Name", ".status.secretName")
		}),
		newCRD(&fleet.ClusterRegistrationRequest{}, func(c crd.CRD) crd.CRD {
			return c.
				WithColumn("Cluster-Name", ".status.clusterName").
				WithColumn("Labels", ".spec.clusterLabels")
		}),
		newCRD(&fleet.Content{}, func(c crd.CRD) crd.CRD {
			c.NonNamespace = true
			c.Status = false
			return c
		}),
	}
}

func Create(ctx context.Context, cfg *rest.Config) error {
	factory, err := crd.NewFactoryFromClient(cfg)
	if err != nil {
		return err
	}

	return factory.BatchCreateCRDs(ctx, List()...).BatchWait()
}

func newCRD(obj interface{}, customize func(crd.CRD) crd.CRD) crd.CRD {
	crd := crd.CRD{
		GVK: schema.GroupVersionKind{
			Group:   "fleet.cattle.io",
			Version: "v1alpha1",
		},
		Status:       true,
		SchemaObject: obj,
	}
	if customize != nil {
		crd = customize(crd)
	}
	return crd
}
