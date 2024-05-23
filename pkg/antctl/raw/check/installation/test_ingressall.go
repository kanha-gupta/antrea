package installation

import (
	"antrea.io/antrea/pkg/antctl/raw/check"
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

type AllIngressConnectivityTest struct{}

func init() {
	RegisterTest("all-ingress-connectivity", &AllIngressConnectivityTest{})
}

func (a AllIngressConnectivityTest) Run(ctx context.Context, testContext *testContext) error {
	networkPolicy := check.ApplyIngressAll(testContext.namespace)
	_, createErr := testContext.client.NetworkingV1().NetworkPolicies(testContext.namespace).Create(ctx, networkPolicy, metav1.CreateOptions{})
	if createErr != nil {
		return fmt.Errorf("error creating network policy: %v", createErr)
	}

	waitErr := wait.PollUntilContextTimeout(ctx, 2*time.Second, 1*time.Minute, true, func(ctx context.Context) (bool, error) {
		_, err := testContext.client.NetworkingV1().NetworkPolicies(testContext.namespace).Get(ctx, "all-ingress-deny", metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	if waitErr != nil {
		return fmt.Errorf("error while waiting for network policy to get ready: %v", waitErr)
	}

	fmt.Println("Network policy successfully implemented.")

	podToServiceTest := &PodToServiceInterNodeConnectivityTest{}
	if err := podToServiceTest.Run(ctx, testContext); err != nil {
		fmt.Errorf("Inter Node Connectivity Test failed: %v", err)
	}
	podToServiceIntraNodeTest := &PodToServiceIntraNodeConnectivityTest{}
	if err := podToServiceIntraNodeTest.Run(ctx, testContext); err != nil {
		fmt.Errorf("Inter Node Connectivity Test failed: %v", err)
	}

	deleteErr := testContext.client.NetworkingV1().NetworkPolicies(testContext.namespace).Delete(ctx, networkPolicy.Name, metav1.DeleteOptions{})
	if deleteErr != nil {
		return fmt.Errorf("error deleting network policy: %v", deleteErr)
	}
	fmt.Println("Network policy successfully deleted.")
	return nil
}
