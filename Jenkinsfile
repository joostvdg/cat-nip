@Library('jenkins-pipeline-library@master') _

import java.text.SimpleDateFormat

currentBuild.displayName = new SimpleDateFormat("yy.MM.dd").format(new Date()) + "-" + env.BUILD_NUMBER
//env.REPO = "https://github.com/vfarcic/go-demo-3.git"
//env.IMAGE = "vfarcic/go-demo-3"
//env.ADDRESS = "go-demo-3-${env.BUILD_NUMBER}-${env.BRANCH_NAME}.acme.com"
//env.CM_ADDR = "cm.acme.com"
//env.TAG = "${currentBuild.displayName}"
//env.TAG_BETA = "${env.TAG}-${env.BRANCH_NAME}"
//env.CHART_VER = "0.0.1"
//env.CHART_NAME = "go-demo-3-${env.BUILD_NUMBER}-${env.BRANCH_NAME}"


def label = "jenkins-slave-${UUID.randomUUID().toString()}"
def CHART_VERSION = ''
def VERSION = ''
def DOCKER_IMAGE_TAG = ''
def FULL_IMAGE_NAME = ''
def IMAGE = ''
def TAG = ''
def FULL_NAME = ''

podTemplate(
        label: label,
        namespace: "go-demo-3-build",
        serviceAccount: "build",
        yaml: """
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: helm
    image: caladreas/helm:2.11.0
    command: ["cat"]
    tty: true
  - name: kubectl
    image: vfarcic/kubectl
    command: ["cat"]
    tty: true
  - name: golang
    image: golang:1.11
    command: ["cat"]
    tty: true
"""
) {
    node(label) {
        node("docker") {
            stage('SCM & Prepare') {
                checkout scm
                def chart = readYaml file: 'helm/cat-nip/Chart.yaml'
                CHART_VERSION = chart.version
                def jenkinsConfig = readYaml file: 'jenkins.yml'
                VERSION = jenkinsConfig.version
            }
            stage('Analysis') {
                // TODO: rewrite this to pods?
                parallel Sonar: {
                    withCredentials([string(credentialsId: 'sonarcloud', variable: 'SONARCLOUD_TOKEN')]) {
                        sh """docker run --rm -v \$(pwd):/root/src \
                        -v /tmp/.scannerwork:/root/src/.scannerwork \
                        -v /tmp/.sonar:/root/src/.sonar \
                        newtmitch/sonar-scanner sonar-scanner \
                          -Dsonar.projectName=cat-nip \
                          -Dsonar.projectKey=joostvdg_cat-nip \
                          -Dsonar.organization=joostvdg-github \
                          -Dsonar.sources=. \
                          -Dsonar.host.url=https://sonarcloud.io \
                          -Dsonar.login=${SONARCLOUD_TOKEN} 
                        """
                    }
                },
                DockerLint: {
                    dockerfileLint()
                }
            }
            stage('Build Docker') {
                DOCKER_IMAGE_TAG = gitNextSemverTag("${VERSION}") + "${env.BRANCH_NAME}"
                FULL_IMAGE_NAME = "${DOCKER_REPO_NAME}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"
                sh "docker image build -t ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ."
            }
            stage('Tag & Push Docker') {
                IMAGE = "${DOCKER_IMAGE_NAME}"
                TAG = "${DOCKER_IMAGE_TAG}"
                FULL_NAME = "${FULL_IMAGE_NAME}"

                withCredentials([usernamePassword(credentialsId: "dockerhub", usernameVariable: "USER", passwordVariable: "PASS")]) {
                    sh "sudo docker login -u $USER -p $PASS"
                }
                sh "docker image tag ${IMAGE}:${TAG} ${FULL_NAME}"
                sh "docker image push ${FULL_NAME}"
            }
            stage('Anchore Validation') {
                anchoreScan("${FULL_IMAGE_NAME}")
            } // end stage
        } // end node docker
        stage("func-test") {
            try {
                container("helm") {
                    sh 'helm version'
                }
                container("kubectl") {
                    sh 'kubectl version'
                }
                container("golang") {
                    sh 'go version'
                }
            } catch(e) {
                error "Failed functional tests"
            }
        }
    } // end node random label
} // end pod def

