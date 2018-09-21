package main

import (
	"github.com/joostvdg/cat-nip/webserver"
	"fmt"
	"github.com/google/uuid"
	"github.com/joostvdg/cat/application"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	serverPort := "8087"
	if len(os.Getenv("SERVER_PORT")) > 0 {
		serverPort = os.Getenv("SERVER_PORT")
	}
	fmt.Printf("=== STARTING WEB SERVER @%s\n", serverPort)
	fmt.Println("=============================================")

	applications := getApplications()
	webserverData := &webserver.WebserverData{Applications: applications, Title: "Central Application Tracker - CAT"}

	c := make(chan bool)
	go webserver.StartServer(serverPort, webserverData, c)
	fmt.Println("> Started the web server, now polling swarm")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	for i := 1; ; i++ { // this is still infinite
		t := time.NewTicker(time.Second * 30)
		select {
		case <-stop:
			fmt.Println("> Shutting down polling")
			break
		case <-t.C:
			fmt.Println("  > Updating Stacks")
			applications := getApplications()
			webserverData.UpdateContainers(applications)
			continue
		}
		break // only reached if the quitCh case happens
	}
	fmt.Println("> Shutting down webserver")
	c <- true
	if b := <-c; b {
		fmt.Println("> Webserver shut down")
	}
	fmt.Println("> Shut down app")

	fmt.Printf("-----------------\n")
}

func getApplications() []application.Application{
	apps := make([]application.Application, 0, 1)
	app1 := application.Application{
		Name:        "Maven Demo Library",
		Description: "A small Maven Java library for demo purposes",
		UUID:        uuid.New().String(),
		Namespace:   "joostvdg",
		ArtifactIDs: []string{"gav://com.github.joostvdg.demo:maven-demo-lib:0.1.1"},
		Sources:     []string{"https://github.com/joostvdg/maven-demo-lib.git"},
		Labels:      []application.Label{ application.Label{Key: "Category", Value: "BuildTool"}},
		Annotations: []application.Annotation { application.Annotation{ Key: "MetricsGroup", Value: "CI", Origin: "com.github.joostvdg"}},
	}
	apps = append(apps, app1)

	app2 := application.Application{
		Name:        "Jenkins",
		Description: "Jenkins, the most awesome CI engine",
		UUID:        uuid.New().String(),
		Namespace:   "CI",
		ArtifactIDs: []string{"https://registry.hub.docker.com/library/jenkins@sha256:81040e35ee59322a02f67ca2584f814d543d5f2f5d361fb8bf4f9e0046f3e809"},
		Sources:     []string{"https://github.com/jenkinsci/jenkins.git"},
		Labels:      []application.Label{ application.Label{Key: "Category", Value: "BuildTool"}},
		Annotations: []application.Annotation { application.Annotation{ Key: "MetricsGroup", Value: "CI", Origin: "com.github.joostvdg"}},
	}
	apps = append(apps, app2)

	return apps
}