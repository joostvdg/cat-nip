@Library('jenkins-pipeline-library@master') _

import java.text.SimpleDateFormat

currentBuild.displayName = new SimpleDateFormat("yy.MM.dd").format(new Date()) + "-" + env.BUILD_NUMBER

def label = "jenkins-slave-${UUID.randomUUID().toString()}"
def CHART_VERSION = ''
def CM_CREDS = 'chartmuseum'
def CHART_NAME = 'cat-nip'
def CM_ADDR = 'https://charts.kearos.net'
def VERSION = ''
def DOCKER_IMAGE_TAG = ''
def FULL_IMAGE_NAME = ''
def IMAGE = ''
def TAG = ''
def FULL_NAME = ''
def DOCKER_IMAGE_TAG_PRD = ''
def DOCKER_REPO_NAME = 'caladreas'
def DOCKER_IMAGE_NAME = 'cat-nip'
def NEW_VERSION = ''

def scmVars
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
  - name: yq
    image: mikefarah/yq
    command: ['cat']
    tty: true
  - name: zapcli
    image: owasp/zap2docker-stable
    command: ["cat"]
    tty: true
  - name: hey
    image:  caladreas/rakyll-hey
    command: ["cat"]
    tty: true
  - name: hub
    image:  caladreas/hub
    command: ["cat"]
    tty: true  
"""
) {
    node(label) {
        node("docker") {
            stage('SCM & Prepare') {
                scmVars = checkout scm
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
                DOCKER_IMAGE_TAG_PRD = gitNextSemverTag("${VERSION}")
                DOCKER_IMAGE_TAG =  "${DOCKER_IMAGE_TAG_PRD}" + "${env.BRANCH_NAME}"
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
            stage('Tag repo') {
                gitRemoteConfigByUrl(scmVars.GIT_URL, 'githubtoken')
                sh '''
                git config --global user.email "jenkins@jenkins.io"
                git config --global user.name "Jenkins"
                '''
                NEW_VERSION = gitNextSemverTag("${VERSION}")
                gitTag("v${NEW_VERSION}")
            }
            stage('Anchore Validation') {
                anchoreScan("${FULL_IMAGE_NAME}")
            } // end stage
        } // end node docker
        stage('Prepare Pod') {
            checkout scm
            // make sure we continue with the tag we've just created
            sh "git checkout v${NEW_VERSION}"
        }
        stage('Update Chart') {
            container("helm") {
                def chartExists = chartExists("${CM_ADDR}", "${CHART_NAME}", "${CHART_VERSION}", "200", "${CM_CREDS}", true)
                if (chartExists) {
                    echo "Chart already exists, not uploading"
                } else {
                    withCredentials([usernamePassword(credentialsId: 'chartmuseum', passwordVariable: 'PSS', usernameVariable: 'USR')]) {
                        sh 'helm package helm/cat-nip'
                        def result = sh returnStdout: true, script: "curl --insecure -u ${USR}:${PSS} --data-binary \"@cat-nip-${CHART_VERSION}.tgz\" ${CM_ADDR}/api/charts"
                        echo "Result=${result}"
                    }
                }
            }
        }
        stage("Staging") {
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
                        sh "helm install --name cat-nip-staging chartmuseum/cat-nip --set image.tag=${DOCKER_IMAGE_TAG} --set nameOverride=cat-nip-staging"
                        sh 'helm ls'
                    }
                }
                parallel Kubectl: {
                    container("kubectl") {
                        sh 'kubectl version'
                    }
                }, Zap: {
//                    container("zapcli") {
//                        sh 'zap-cli quick-scan -sc -f json --start-options \'-config api.disablekey=true\' http://cat-nip-staging > zap.json'
//                        archiveArtifacts 'zap.json'
//                    }
                    container("kubectl") {
                        sh 'kubectl run zapcli --image=owasp/zap2docker-stable --restart=Never -- zap-cli quick-scan -sc -f json --start-options \'-config api.disablekey=true\' http://cat-nip-staging.build'
                        sleep 30
                        sh 'kubectl logs zapcli > zap.json'
                        archiveArtifacts 'zap.json'
                        sh 'kubectl delete pod zapcli'
                    }
                }, Hey: {
                    container("hey") {
                        sh 'hey -n 1000 -c 100 http://cat-nip-staging.build > perf.txt'
                        archiveArtifacts 'perf.txt'
                    }
                }
            } catch(e) {
                error "Failed functional tests"
            } finally {
                container("helm") {
                    withCredentials([file(credentialsId: 'letsencrypt-staging-ca', variable: 'CA_PEM')]) {
                        sh 'helm ls'
                        sh 'helm del --purge cat-nip-staging'
                    }
                }
            }
        } // end stage
        stage('Promote Image') {
            // TODO: retag image
            // push updated tagged image
            // create git tag
            def STAGING = "${FULL_IMAGE_NAME}"
            def PRD = "${DOCKER_REPO_NAME}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG_PRD}"
            node('docker') {
                withCredentials([usernamePassword(credentialsId: "dockerhub", usernameVariable: "USER", passwordVariable: "PASS")]) {
                    sh "docker login -u $USER -p $PASS"
                }
                sh "docker image tag ${STAGING} ${PRD}"
                sh "docker image push ${PRD}"
            }
        }
        stage('Update PROD') {
            // TODO: create PR for environment config
            def gitInfo = git 'https://github.com/joostvdg/environments.git'
            echo "${gitInfo}"
            def branchName = "chart-${CHART_NAME}-${DOCKER_IMAGE_TAG_PRD}"
            sh "git checkout -b ${branchName}"
            container('yq') {
                script {
                    sh 'yq r cb/aws-eks/cat-nip/image-values.yml image.tag'
                    sh "yq w -i cb/aws-eks/cat-nip/image-values.yml image.tag ${DOCKER_IMAGE_TAG_PRD}"
                    sh 'yq r cb/aws-eks/cat-nip/image-values.yml image.tag'
                }
            }
            container('hub') {
                sh 'git status'
                gitRemoteConfigByUrl(gitInfo.GIT_URL, 'githubtoken')
                sh '''git config --global user.email "jenkins@jenkins.io"
                    git config --global user.name "Jenkins"
                '''
                sh """git add cb/aws-eks/cat-nip/image-values.yml
                git commit -m "update ${CHART_NAME} to image ${DOCKER_IMAGE_TAG_PRD}"
                git push origin ${branchName}
                """


                writeFile encoding: 'UTF-8', file: 'pr-info.md', text: """update ${CHART_NAME} to image ${DOCKER_IMAGE_TAG_PRD} 
                This pr is automatically generated via Jenkins.
                The job: ${env.JOB_URL}"""

                // TODO: create PR
                // Do we need '--no-edit' ?
                sh "hub pull-request -F pr-info.md -l '${CHART_NAME}'"
            }
        }
    } // end node random label
} // end pod def

