package indexer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	NodeNameIndex = "nodeName"
)

func NodeNameIndexFunc(obj interface{}) ([]string, error) {
	pod, _ := obj.(*corev1.Pod)
	return []string{pod.Spec.NodeName}, nil
}

func newFakePod(podname, podnamespace, nodeName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podname, Namespace: podnamespace},
		Spec:       corev1.PodSpec{NodeName: nodeName},
	}
}

func TestIndex(t *testing.T) {
	indexer := cache.NewIndexer(
		cache.MetaNamespaceKeyFunc, // Use <namespace>/<name> as a key if <namespace> exists, otherwise <name>
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc, NodeNameIndex: NodeNameIndexFunc}, // default index function that indexes based on an object's namespace
	)

	pod11 := newFakePod("foo1", "fakeNs1", "fakeNode1")
	pod12 := newFakePod("bar1", "fakeNs1", "fakeNode1")
	pod21 := newFakePod("foo2", "fakeNs2", "fakeNode2")
	pod22 := newFakePod("bar2", "fakeNs2", "fakeNode2")

	indexer.Add(pod11)
	indexer.Add(pod12)
	indexer.Add(pod21)
	indexer.Add(pod22)

	// Index(indexName string, obj interface{}) ([]interface{}, error)
	s0, _ := indexer.Index("nodeName", pod11) // expected []object{foo1, bar1}
	assert.Equal(t, []interface{}{pod11, pod12}, s0)

	// IndexKeys(indexName, indexedValue string) ([]string, error)
	s1, _ := indexer.IndexKeys("nodeName", "fakeNode1") // expected []string{foo1, bar1}
	assert.EqualValues(t, []string{"fakeNs1/bar1", "fakeNs1/foo1"}, s1, "")

	// ListIndexFuncValues(indexName string) []string
	s2 := indexer.ListIndexFuncValues("namespace") // expected []string{fakeNs1, fakeNs2}
	assert.Equal(t, []string{"fakeNs1", "fakeNs2"}, s2, "")

	// ByIndex(indexName, indexedValue string) ([]interface{}, error)
	s3, _ := indexer.ByIndex("nodeName", "fakeNode1") // expectd []Object
	assert.EqualValues(t, []interface{}{pod11, pod12}, s3, "")

}
