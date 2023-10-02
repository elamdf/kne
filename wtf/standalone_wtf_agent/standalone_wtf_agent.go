package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "k8s.io/client-go/kubernetes"
    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/tools/clientcmd"

)

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
                if true { 
                //if event.Type == "ADDED" || event.Type == "MODIFIED" {
                    eventObj := event.Object.(*v1.Event)
                    fmt.Printf("%s/%s: %s - %s\n", eventObj.Namespace, eventObj.Name, eventObj.Reason, eventObj.Message)
                }
            case <-ctx.Done():
                return
            }
        }
    }
}
