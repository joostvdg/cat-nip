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
def DOCKER_REPO_NAME = 'caladreas'
def DOCKER_IMAGE_NAME = 'cat-nip'

// TODO: introduce specific namespace
// TODO: introduce service account

podTemplate(
        label: label,
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
  - name: zapcli
    image: owasp/zap2docker-stable
    command: ["cat"]
    tty: true
  - name: hey
    image:  caladreas/rakyll-hey
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
                    sh "docker login -u $USER -p $PASS"
                }
                sh "docker image tag ${IMAGE}:${TAG} ${FULL_NAME}"
                sh "docker image push ${FULL_NAME}"
            }
            stage('Anchore Validation') {
                anchoreScan("${FULL_IMAGE_NAME}")
            } // end stage
        } // end node docker
        stage("func-test") {
            // TODO: deploy staging version via helm chart with current 'staging image tag'
            // docker run -i --rm --name zapcli -v $(pwd):/tmp -w /tmp owasp/zap2docker-stable zap-cli quick-scan -f json -sc --start-options '-config api.disablekey=true' https://catnip.kearos.net
            // kubectl run zapcli --image=owasp/zap2docker-stable --restart=Never -- zap-cli quick-scan -sc -f json --start-options '-config api.disablekey=true' https://catnip.kearos.net
            // sh 'docker run -v $(pwd):/tmp -w /tmp caladreas/rakyll-hey hey -n 1000 -c 100 https://catnip.kearos.net/ > perf.txt'
            // sh 'cat perf.txt'
            // archiveArtifact 'perf.txt'
            try {
                container("helm") {
                    sh 'helm version'
                    sh 'helm ls'
                    withCredentials([file(credentialsId: 'letsencrypt-staging-ca', variable: 'CA_PEM')]) {
                        withCredentials([usernamePassword(credentialsId: 'chartmuseum', passwordVariable: 'PSS', usernameVariable: 'USR')]) {
                            sh "helm repo add chartmuseum https://charts.kearos.net --username ${USR} --password ${PSS}  --ca-file ${CA_PEM}"
                        }
                        sh 'helm repo list'
                        sh 'helm repo update'
                        sh "helm install --name cat-nip-staging chartmuseum/cat-nip --set image.tag=${DOCKER_IMAGE_TAG}"
                        sh 'helm ls'
                    }

                }
                parallel Kubectl: {
                    container("kubectl") {
                        sh 'kubectl version'
                    }
                }, Zap: {
//                    container("zapcli") {
//                        sh 'zap-cli quick-scan -sc -f json --start-options \'-config api.disablekey=true\' https://catnip.kearos.net > zap.json'
//                        archiveArtifacts 'zap.json'
//                    }
                    container("kubectl") {
                        sh 'kubectl run zapcli --image=owasp/zap2docker-stable --restart=Never -- zap-cli quick-scan -sc -f json --start-options \'-config api.disablekey=true\' https://catnip.kearos.net'
                        sleep 30
                        sh 'kubectl logs zapcli > zap.json'
                        archiveArtifacts 'zap.json'
                        sh 'kubectl delete pod zapcli'
                    }
                }, Hey: {
                    container("hey") {
                        sh 'hey -n 1000 -c 100 https://catnip.kearos.net/ > perf.txt'
                        archiveArtifacts 'perf.txt'
                    }
                }
            } catch(e) {
                container("helm") {
                    withCredentials([file(credentialsId: 'letsencrypt-staging-ca', variable: 'CA_PEM')]) {
                        sh 'helm ls'
                        sh 'helm delete cat-nip-staging --purge'
                    }
                }
                error "Failed functional tests"
            }
        }
    } // end node random label
} // end pod def

