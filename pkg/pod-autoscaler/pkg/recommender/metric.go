package recommender

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type Client struct {
	httpClient http.Client
	host       string
}

// NewMetricClient returns a new MetricClient representing a metric client.
func NewMetricClient() *Client {
	httpClient := http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 90 * time.Second,
			}).DialContext,
			// TODO: Some of those value should be tuned
			MaxIdleConns:          10,
			IdleConnTimeout:       90 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: 5 * time.Second,
	}
	client := &Client{
		httpClient: httpClient,
		host:       "ingress-nginx-controller.ingress-nginx.svc.cluster.local:7000",
	}
	return client
}

// getMetrics returns a list of the the metrics of a pod.
func (c Client) getMetrics(pod *v1.Pod) (map[string]interface{}, error) {

	// Retrieve the location of the pod's metrics server
	address := pod.Status.PodIP

	// Compose host and path
	host := c.host
	path := fmt.Sprintf("%s/metrics/window/minute/5", address)

	// Create the request
	metricServerURL := url.URL{
		// TODO: http is not a good protocol for polling data
		// grpc can be a good alternative, but the pods should implement a grpc server
		// offerings the metrics
		Scheme: "http",
		Host:   host,
		Path:   path,
	}
	request, err := http.NewRequest(http.MethodGet, metricServerURL.String(), nil)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// Send the request
	response, err := c.httpClient.Do(request)
	if err != nil {
		klog.Error(err)
	}

	// Parse the response
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var metricMap map[string]interface{}
	klog.Info(string(body))
	err = json.Unmarshal(body, &metricMap)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return metricMap, nil
}
