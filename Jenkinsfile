pipeline {
    options {
        buildDiscarder logRotator(artifactDaysToKeepStr: '5', artifactNumToKeepStr: '5', daysToKeepStr: '5', numToKeepStr: '5')
        durabilityHint 'PERFORMANCE_OPTIMIZED'
        timeout(5)
    }
    agent { label 'docker' }
    libraries {
        lib('jenkins-pipeline-library@master')
    }
    environment {
        CHART_NAME = 'cat-nip'
        CM_ADDR = 'https://charts.kearos.net/'
        VERSION = ''
        CHART_VERSION = ''
        DOCKER_IMAGE_NAME = 'cat-nip'
        DOCKER_REPO_NAME = 'caladreas'
        DOCKER_IMAGE_TAG = ''
        FULL_IMAGE_NAME = ''
    }
    stages {
        stage('Test versions') {
            steps {
                sh 'uname -a'
                sh 'docker version'
                sh 'java -version'
            }
        }
        stage('Prepare') {
            steps {
                script {
                    def chart = readYaml file: 'helm/cat-nip/Chart.yaml'
                    CHART_VERSION = chart.version
                    def jenkinsConfig = readYaml file: 'jenkins.yml'
                    VERSION = jenkinsConfig.version
                }
            }
        }
        stage('Analysis') {
            environment {
                SONARCLOUD_TOKEN = credentials('sonarcloud')
            }
            steps {
                parallel(
                        Sonar: {
                            sh """
                        docker run -v \$(pwd):/root/src newtmitch/sonar-scanner sonar-scanner \
                          -Dsonar.projectName=cat-nip \
                          -Dsonar.projectKey=joostvdg_cat-nip \
                          -Dsonar.organization=joostvdg-github \
                          -Dsonar.sources=. \
                          -Dsonar.host.url=https://sonarcloud.io \
                          -Dsonar.login=${SONARCLOUD_TOKEN}
                        """
                        },
                        DockerLint: {
                            dockerfileLint()
                        }
                )
            }
        }
        stage('Build Docker') {
            steps {
                script {
                    DOCKER_IMAGE_TAG = gitNextSemverTag(VERSION) + "${env.BRANCH_NAME}"
                    FULL_IMAGE_NAME = "${DOCKER_REPO_NAME}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"
                }
                sh 'docker image build -t ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} .'
            }
        }
        stage('Tag & Push Docker') {
            environment {
                DOCKERHUB = credentials('dockerhub')
            }
            steps {
                sh 'docker login -u ${DOCKERHUB_USR} -p ${DOCKERHUB_PSW}'
                sh 'docker image tag ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ${FULL_IMAGE_NAME}'
                sh 'docker image push ${FULL_IMAGE_NAME}'
            }
        }
        stage('Anchore Validation') {
            steps {
                anchoreScan(FULL_IMAGE_NAME)
            }
        }
        stage('Helm Chart update') {
            when {
                branch 'master'
            }
            environment {
                CM = credentials('chartmuseum')
            }
            steps {

//                container("helm") {
//                    sh 'helm package helm/cat-nip'
//                    sh 'curl -u ${CM_USR}:${CM_PSW} --data-binary "@cat-nip-${CHART_VER}.tgz" http://${CM_ADDR}/api/charts'
//                }
                sh 'docker run -v $(pwd):/root/src/ -w /root/src vfarcic/helm:2.9.1 helm package helm/cat-nip'
                sh 'curl -u ${CM_USR}:${CM_PSW} --data-binary "@cat-nip-${CHART_VER}.tgz" http://${CM_ADDR}/api/charts'
            }
        }
    }
    post {
        always {
            cleanWs notFailBuild: true
        }
    }
}
