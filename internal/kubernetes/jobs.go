package jobs

import (
    "context"
    "fmt"
    "log"
    "bufio"
    "time"

    batchv1 "k8s.io/api/batch/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/apimachinery/pkg/api/resource"
    "compute-share/backend/internal/models"
)

type podResult struct {
    pod string
    text string
}

// CreateKubernetesJob creates a job in Kubernetes based on the JobRequest
func CreateKubernetesJob(req models.Job) (*batchv1.Job, error) {
    config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
    if err != nil {
        return nil, err
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, err
    }

    // Convert resource limit strings to resource.Quantity
    cpuLimits, err := resource.ParseQuantity(req.CPULimits)
    if err != nil {
        return nil, err
    }
    memoryLimits, err := resource.ParseQuantity(req.MemoryLimits)
    if err != nil {
        return nil, err
    }

    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      req.JobName,
            Namespace: req.Namespace,
        },
        Spec: batchv1.JobSpec{
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:    req.JobName,
                            Image:   req.ImageName,
                            Command: req.Command,
                            Args:    req.Args,
                            Resources: corev1.ResourceRequirements{
                                Limits: corev1.ResourceList{
                                    corev1.ResourceCPU:    cpuLimits,
                                    corev1.ResourceMemory: memoryLimits,
                                },
                            },
                        },
                    },
                    RestartPolicy: corev1.RestartPolicyNever,
                },
            },
        },
    }

    // Create the job
    return clientset.BatchV1().Jobs(req.Namespace).Create(context.TODO(), job, metav1.CreateOptions{})
}

// Waits for job to complete
// func WatchJobCompletion(namespace string, jobName string) {
//     config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
//     if err != nil {
//         log.Fatalf("Failed to build config: %v", err)
//         return
//     }

//     clientset, err := kubernetes.NewForConfig(config)
//     if err != nil {
//         log.Fatalf("Failed to create clientset: %v", err)
//         return
//     }

//     watcher, err := clientset.BatchV1().Jobs(namespace).Watch(context.TODO(), metav1.ListOptions{
//         FieldSelector: fmt.Sprintf("metadata.name=%s", jobName),
//     })
//     if err != nil {
//         log.Fatalf("Failed to watch job: %v", err)
//         return
//     }

//     ch := watcher.ResultChan()
//     log.Printf("Watching for completion of job: %s", jobName)
//     for event := range ch {
//         job, ok := event.Object.(*batchv1.Job)
//         if !ok {
//             log.Fatal("Error: Unexpected type in job watch")
//             watcher.Stop()
//             break
//         }

//         if job.Status.Succeeded > 0 {
//             duration := calculateJobDuration(job)
//             log.Printf("Job completed successfully in %d seconds", duration)
            
//             results := fetchJobResult(job, clientset)
//             logResults(results)

//             watcher.Stop()
//             deleteJob(job, clientset)
//             break
//         }
//         if job.Status.Failed > 0 {
//             log.Println("Job failed")

//             watcher.Stop()
//             deleteJob(job, clientset)
//             break
//         }
//     }
// }
func WatchJobStatus(namespace string, jobName string) {
    config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
    if err != nil {
        log.Fatalf("Failed to build config: %v", err)
        return
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatalf("Failed to create clientset: %v", err)
        return
    }

    watcher, err := clientset.BatchV1().Jobs(namespace).Watch(context.TODO(), metav1.ListOptions{
        FieldSelector: fmt.Sprintf("metadata.name=%s", jobName),
    })
    if err != nil {
        log.Fatalf("Failed to watch job: %v", err)
        return
    }

    ch := watcher.ResultChan()
    log.Printf("Watching for completion of job: %s", jobName)
    for event := range ch {
        job, ok := event.Object.(*batchv1.Job)
        if !ok {
            log.Fatal("Error: Unexpected type in job watch")
            watcher.Stop()
            break
        }

        podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
            LabelSelector: fmt.Sprintf("job-name=%s", jobName),
        })
        if err != nil {
            log.Fatalf("Failed to list pods: %v", err)
            watcher.Stop()
            break
        }

        for _, pod := range podList.Items {
            for _, status := range pod.Status.ContainerStatuses {
                if status.State.Waiting != nil {
                    reason := status.State.Waiting.Reason
                    log.Printf("Waiting reason: %s", reason)
                    if reason == "ErrImagePull" || reason == "ImagePullBackOff" {
                        log.Printf("Pod error: %s - %s", reason, status.State.Waiting.Message)
                        log.Printf("Stopping watcher and deleting job %s", jobName)
                        watcher.Stop()
                        deleteJob(job, clientset)
                        return
                    }
                }
            }
        }

        if job.Status.Succeeded > 0 {
            duration := calculateJobDuration(job)
            log.Printf("Job completed successfully in %d seconds", duration)

            results := fetchJobResult(job, clientset)
            logResults(results)

            watcher.Stop()
            deleteJob(job, clientset)
            break
        }
        if job.Status.Failed > 0 {
            log.Println("Job failed")

            watcher.Stop()
            deleteJob(job, clientset)
            break
        }
        time.Sleep(5 * time.Second)
    }
}



//// helpers

func calculateJobDuration(job *batchv1.Job) int64 {
    if job.Status.StartTime == nil || job.Status.CompletionTime == nil {
        log.Println("Job start or completion time is nil")
        return 0
    }
    duration := job.Status.CompletionTime.Time.Sub(job.Status.StartTime.Time)
    return int64(duration.Seconds())
}

func deleteJob(job *batchv1.Job, clientset *kubernetes.Clientset) {
    deletePolicy := metav1.DeletePropagationForeground
    deleteOptions := metav1.DeleteOptions{
        PropagationPolicy: &deletePolicy,
    }
    err := clientset.BatchV1().Jobs(job.Namespace).Delete(context.TODO(), job.Name, deleteOptions)
    if err != nil {
        log.Fatalf("Failed to delete job: %v", err)
    } else {
        log.Printf("Job %s deleted successfully", job.Name)
    }
}

// Retrieve and log the results of the job's pods
func fetchJobResult(job *batchv1.Job, clientset *kubernetes.Clientset) []podResult  {
    pods, err := clientset.CoreV1().Pods(job.Namespace).List(context.TODO(), metav1.ListOptions{
        LabelSelector: fmt.Sprintf("job-name=%s", job.Name),
    })
    var results []podResult 
    if err != nil {
        log.Printf("Failed to list pods for job %s: %v", job.Name, err)
    } else {
        for _, pod := range pods.Items {
            log.Printf("Fetching logs for pod: %s", pod.Name)
            logStream, err := clientset.CoreV1().Pods(job.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{}).Stream(context.TODO())
            if err != nil {
                log.Printf("Failed to get logs for pod %s: %v", pod.Name, err)
                continue
            }
            defer logStream.Close()
            scanner := bufio.NewScanner(logStream)
            for scanner.Scan() {
                result := podResult {
                    pod: pod.Name,
                    text: scanner.Text(),
                }
                results = append(results, result)
            }
            if err := scanner.Err(); err != nil {
                log.Printf("Error reading logs for pod %s: %v", pod.Name, err)
            }
        }
    }
    return results
}

func logResults(results []podResult) {
    for _, result := range results {
        log.Printf("Pod %s log: %s", result.pod, result.text)
    }
}