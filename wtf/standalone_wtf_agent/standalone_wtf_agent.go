package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type HealthBit int

const (
    Red HealthBit = iota
    Yellow
    Green
)
var decode_bit = map[HealthBit]string{Red:"Red", Yellow:"Yellow", Green:"Green"}

var m = make(map[string]HealthBit)
var lock = sync.RWMutex{}

func main() {
    kubeconfig, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
    if err != nil {
        fmt.Printf("Error loading kubeconfig: %v\n", err)
        os.Exit(1)
    }
    // Create a Kubernetes clientset.
    clientset, err := kubernetes.NewForConfig(kubeconfig)
    if err != nil {
        fmt.Printf("Error creating clientset: %v\n", err)
        os.Exit(1)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    signalCh := make(chan os.Signal, 1)
    signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-signalCh
        fmt.Println("Received termination signal. Stopping...")
        cancel()
    }()
    go func () {
        for {
            // declaring the variable using the var keyword
            var id string
            fmt.Println("Please enter your query in the form 'namespace/pod' ")
            
            // TODO should be able to dump all statuses to a file
            // scanning the input by the user
            fmt.Scanln(&id)
            if val, ok := m[id]; ok {
                //do something here
                fmt.Printf("%s status: %s\n", id, decode_bit[val])
            } else {
                fmt.Printf("pod/namespace '%s' not found!\n", id)
            }
        }
    }()

    for {
        watchOptions := metav1.ListOptions{}
        eventWatch, err := clientset.CoreV1().Events(metav1.NamespaceAll).Watch(ctx, watchOptions)
        if err != nil {
            fmt.Printf("Error watching events: %v\n", err)
            os.Exit(1)
        }

        for {
            select {
            case event := <-eventWatch.ResultChan():
                eventObj := event.Object.(*v1.Event)
                if eventObj.Type != "Normal" { 
                    lock.Lock()
                    // TODO should this have a latch behavior
fmt.Printf("%s/%s\n", eventObj.Namespace, eventObj.Name)
                    m[fmt.Sprintf("%s/%s", eventObj.Namespace, eventObj.Name)] = Red
                    lock.Unlock()
                }
            case <-ctx.Done():
                return
            }
        }
    }
}
