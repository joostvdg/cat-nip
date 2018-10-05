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
        CHART_VERSION = 'v0.1.0'
    }
    stages {
        stage('Build Docker') {
            steps {
                sh "docker image build -t catnip-${env.BRANCH_NAME} ."
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
                        },
                        Anchore: {
                            anchoreScan("catnip-${env.BRANCH_NAME}")
                        }
                )
            }
        }
        stage('Tag & Push Docker') {
            when {
                branch 'master'
            }
            environment { 
                DOCKERHUB = credentials('dockerhub') 
            }
            steps {
                sh 'docker login -u ${DOCKERHUB_USR} -p ${DOCKERHUB_PSW}'
                sh "docker image tag catnip-${env.BRANCH_NAME} caladreas/catnip-${env.BRANCH_NAME}"
                sh "docker image push caladreas/catnip-${env.BRANCH_NAME}"
            }
        }
        stage('Helm Chart update') {
            environment {
                CM = credentials('chartmuseum')
            }
            steps {
                script {
                    CHART_VERSION = readYaml('helm/Chart.yml').version
                }
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
