package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Ec2Instancetype struct {
	id     string
	status string
}

var (
	ec2Instances = make(map[string]*Ec2Instancetype)
	mu           sync.Mutex
)

func ec2MockCreation(instanceId string) {
	time.Sleep(time.Duration(rand.Intn(20)+5) * time.Second)

	mu.Lock()
	defer mu.Unlock()
	ec2Instances[instanceId].status = "running"

}

func shortPoll(w http.ResponseWriter, r *http.Request) {
	instanceID := r.URL.Query().Get("id")
	if instanceID == "" {
		http.Error(w, "Instance ID is required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	instance, exists := ec2Instances[instanceID]
	mu.Unlock()

	if !exists {
		http.Error(w, "instance doesn't exist", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "id %s status is %s", instance.id, instance.status)
	w.WriteHeader(http.StatusOK)
}

func longPoll(w http.ResponseWriter, r *http.Request) {

	instanceID := r.URL.Query().Get("id")
	if instanceID == "" {
		http.Error(w, "Instance ID is required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	instance, exists := ec2Instances[instanceID]
	mu.Unlock()

	if !exists {
		http.Error(w, "instance doesn't exist", http.StatusNotFound)
		return
	}

	for {
		mu.Lock()
		if instance.status == "running" {
			mu.Unlock()
			fmt.Fprintf(w, "id %s status is %s", instance.id, instance.status)
			return
		}
		mu.Unlock()
		time.Sleep(1 * time.Second)
	}

}

func startEc2InstancesMock() {
	ec2InstanceIds := []string{"i-143", "i-144", "i-145"}

	for _, instanceId := range ec2InstanceIds {
		ec2Instances[instanceId] = &Ec2Instancetype{id: instanceId, status: "pending"}
		go ec2MockCreation(instanceId)
	}

}

func main() {
	http.HandleFunc("/short-poll", shortPoll)
	http.HandleFunc("/long-poll", longPoll)

	startEc2InstancesMock()

	http.ListenAndServe(":6969", nil)
}
