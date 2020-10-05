package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo/workflow/controller/cache"
)

var sampleConfigMapCacheEntry = apiv1.ConfigMap{
	Data: map[string]string{
		"hi-there-world": `{"nodeID":"memoized-simple-workflow-5wj2p","outputs":{"parameters":[{"name":"hello","value":"foobar","valueFrom":{"path":"/tmp/hello_world.txt"}}],"artifacts":[{"name":"main-logs","archiveLogs":true,"s3":{"endpoint":"minio:9000","bucket":"my-bucket","insecure":true,"accessKeySecret":{"name":"my-minio-cred","key":"accesskey"},"secretKeySecret":{"name":"my-minio-cred","key":"secretkey"},"key":"memoized-simple-workflow-5wj2p/memoized-simple-workflow-5wj2p/main.log"}}]},"creationTimestamp":"2020-09-21T18:12:56Z"}`,
	},
	TypeMeta: metav1.TypeMeta{
		Kind:       "ConfigMap",
		APIVersion: "v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:            "whalesay-cache",
		ResourceVersion: "1630732",
	},
}

var sampleConfigMapEmptyCacheEntry = apiv1.ConfigMap{
	TypeMeta: metav1.TypeMeta{
		Kind:       "ConfigMap",
		APIVersion: "v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:            "whalesay-cache",
		ResourceVersion: "1630732",
	},
}

func TestConfigMapCacheLoadHit(t *testing.T) {
	cancel, controller := newController()
	defer cancel()
	_, err := controller.kubeclientset.CoreV1().ConfigMaps("default").Create(&sampleConfigMapCacheEntry)
	assert.NoError(t, err)
	c := cache.NewConfigMapCache("default", controller.kubeclientset, "whalesay-cache")
	entry, err := c.Load("hi-there-world")
	assert.NoError(t, err)
	outputs := entry.Outputs
	assert.NoError(t, err)
	if assert.Len(t, outputs.Parameters, 1) {
		assert.Equal(t, "hello", outputs.Parameters[0].Name)
		assert.Equal(t, "foobar", *outputs.Parameters[0].Value)
	}
}

func TestConfigMapCacheLoadMiss(t *testing.T) {
	cancel, controller := newController()
	defer cancel()
	_, err := controller.kubeclientset.CoreV1().ConfigMaps("default").Create(&sampleConfigMapEmptyCacheEntry)
	assert.NoError(t, err)
	c := cache.NewConfigMapCache("default", controller.kubeclientset, "whalesay-cache")
	entry, err := c.Load("hi-there-world")
	assert.NoError(t, err)
	assert.Nil(t, entry)
}

func TestConfigMapCacheSave(t *testing.T) {
	var MockParamValue string = "Hello world"
	var MockParam = wfv1.Parameter{
		Name:  "hello",
		Value: &MockParamValue,
	}
	cancel, controller := newController()
	defer cancel()
	c := cache.NewConfigMapCache("default", controller.kubeclientset, "whalesay-cache")
	outputs := wfv1.Outputs{}
	outputs.Parameters = append(outputs.Parameters, MockParam)
	err := c.Save("hi-there-world", "", &outputs)
	assert.NoError(t, err)
	cm, err := controller.kubeclientset.CoreV1().ConfigMaps("default").Get("whalesay-cache", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, cm)
}