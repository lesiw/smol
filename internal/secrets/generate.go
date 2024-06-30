package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"lesiw.io/flag"
	"lesiw.io/smol/internal/randstr"
)

var (
	flags = flag.NewSet(os.Stderr, "usage: generate.go -n NAMESPACE")
	ns    = flags.String("n", "Kubernetes namespace")

	secrets = []string{"db-secret"}

	errFlag = errors.New("flag parse error")
)

func main() {
	if err := run(); err != nil {
		if err != errFlag {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func run() error {
	if err := flags.Parse(os.Args[1:]...); err != nil {
		return errFlag
	}
	if *ns == "" {
		flags.PrintError("bad namespace")
		return errFlag
	}

	namespace := strings.Replace(*ns, ".", "-", -1)
	ctx := context.Background()

	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	_, err = clientset.CoreV1().Namespaces().
		Get(context.TODO(), namespace, metav1.GetOptions{})
	if err == nil {
		fmt.Fprintf(os.Stderr, "namespace '%s' exists\n", namespace)
	} else {
		_, err = clientset.CoreV1().Namespaces().Create(
			ctx, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}, metav1.CreateOptions{},
		)
		if err != nil {
			return fmt.Errorf("failed to create namespace '%s': %w",
				namespace, err)
		}
		fmt.Fprintf(os.Stderr, "created namespace '%s'\n", namespace)
	}

	for _, secret := range secrets {
		_, err = clientset.CoreV1().Secrets(namespace).
			Get(ctx, secret, metav1.GetOptions{})
		if err == nil {
			fmt.Fprintf(os.Stderr, "secret '%s' exists\n", secret)
			continue
		}
		val, err := randstr.New(32)
		if err != nil {
			return fmt.Errorf("failed generating secret: %w", err)
		}
		_, err = clientset.CoreV1().
			Secrets(namespace).
			Create(ctx, &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: secret,
				},
				Data: map[string][]byte{
					"secret": []byte(val),
				},
			}, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed writing secret '%s': %w", secret, err)
		}
		fmt.Fprintf(os.Stderr, "created secret '%s'\n", secret)
	}
	return nil
}
