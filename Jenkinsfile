
pipeline {
    agent { label 'docker' }
    libraries {
        lib('jenkins-pipeline-library@master')
    }
    stages {
        stage('Test Docker Version') {
            steps {
                sh 'docker version'
            }
        }
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
                          -Dsonar.projectName=cat-nip
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
    }
    post {
        always {
            cleanWs notFailBuild: true
        }
    }
}
