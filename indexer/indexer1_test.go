package indexer_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/validation/path"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
)

func mokPod(podname, podnamespace, nodeName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podname, Namespace: podnamespace},
		Spec:       corev1.PodSpec{NodeName: nodeName},
	}
}

var (
	KeyRootFunc = func(ctx context.Context) string {
		return namespaceKeyRootFunc(ctx, "/pod")
	}
	KeyFunc = func(ctx context.Context, name string) (string, error) {
		return namespaceKeyFunc(ctx, "/pod", name)
	}

	keyFunc = func(obj runtime.Object) (string, error) {
		accessor, err := meta.Accessor(obj)
		if err != nil {
			return "", err
		}
		return KeyFunc(genericapirequest.WithNamespace(genericapirequest.NewContext(), accessor.GetNamespace()), accessor.GetName())
	}
)

func TestAPIKeyFuncs(t *testing.T) {
	pod11 := mokPod("foo1", "fakeNs1", "fakeNode1")
	key, _ := keyFunc(pod11)
	assert.Equal(t, "/pod/fakeNs1/foo1", key, "")
}

func namespaceKeyRootFunc(ctx context.Context, prefix string) string {
	key := prefix
	ns, ok := genericapirequest.NamespaceFrom(ctx)
	if ok && len(ns) > 0 {
		key = key + "/" + ns
	}
	return key
}

func namespaceKeyFunc(ctx context.Context, prefix string, name string) (string, error) {
	key := namespaceKeyRootFunc(ctx, prefix)
	ns, ok := genericapirequest.NamespaceFrom(ctx)
	if !ok || len(ns) == 0 {
		return "", apierrors.NewBadRequest("Namespace parameter required.")
	}
	if len(name) == 0 {
		return "", apierrors.NewBadRequest("Name parameter required.")
	}
	if msgs := path.IsValidPathSegmentName(name); len(msgs) != 0 {
		return "", apierrors.NewBadRequest(fmt.Sprintf("Name parameter invalid: %q: %s", name, strings.Join(msgs, ";")))
	}
	key = key + "/" + name
	return key, nil
}
