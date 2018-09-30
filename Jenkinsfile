
pipeline {
    agent { label 'docker' }
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
