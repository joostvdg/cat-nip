pipeline {
    options {
        buildDiscarder logRotator(artifactDaysToKeepStr: '5', artifactNumToKeepStr: '5', daysToKeepStr: '5', numToKeepStr: '5')
    }
    agent {
        kubernetes {
            label 'jenkins-agent-cat-nip'
            yaml """
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: helm
    image: caladreas/helm:2.11.0
    command: ["cat"]
    tty: true
  - name: golang
    image: golang:1.11
    command:
    - cat
    tty: true
  - name: hadolint
    image: hadolint/hadolint:latest-debian
    command:
    - cat
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
  - name: sonar
    image:  newtmitch/sonar-scanner
    command: ["cat"]
    tty: true    
  - name: kaniko
    image: gcr.io/kaniko-project/executor:debug
    imagePullPolicy: Always
    command:
    - /busybox/cat
    tty: true
    volumeMounts:
      - name: jenkins-docker-cfg
        mountPath: /root
  volumes:
  - name: jenkins-docker-cfg
    projected:
      sources:
      - secret:
          name: regcred
          items:
            - key: .dockerconfigjson
              path: .docker/config.json
        """
        }
    }
    libraries {
        lib('core@master')
        lib('gitops-k8s@master')
    }
    environment {
        label = "jenkins-slave-catnip"
        CHART_VERSION = ''
        CM_CREDS = 'chartmuseum'
        CHART_NAME = 'cat-nip'
        CM_ADDR = 'https://charts.kearos.net'
        VERSION = ''
        DOCKER_IMAGE_TAG = ''
        FULL_IMAGE_NAME = ''
        IMAGE = ''
        TAG = ''
        FULL_NAME = ''
        DOCKER_IMAGE_TAG_PRD = ''
        DOCKER_REPO_NAME = 'caladreas'
        DOCKER_IMAGE_NAME = 'cat-nip'
        NEW_VERSION = ''
        scmVars =''
        envGitInfo = ''
        envBranchName = ''
    }
    stages {
        stage('SCM & Prepare') {
            steps {
                script {
                    scmVars = checkout scm
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
                        container('sonar') {
                            // because the workspace is automatically mounted via the jnlp agent
                            // and the sonar scanner image is fixed on /root/src, we first create a symlink
                            sh "ln -s ${WORKSPACE} /root/src"
                            sh '''sonar-scanner \
                              -Dsonar.projectName=cat-nip \
                              -Dsonar.projectKey=joostvdg_cat-nip \
                              -Dsonar.organization=joostvdg-github \
                              -Dsonar.sources=. \
                              -Dsonar.host.url=https://sonarcloud.io \
                              -Dsonar.login=${SONARCLOUD_TOKEN} 
                            '''
                        }
                    },
                    DockerLint: {
                        container('hadolint') {
                            dockerfileLintK8s()
                        }
                    }
                )
            }
        }
        stage('Build') {
            steps {
                script {
                    DOCKER_IMAGE_TAG_PRD = gitNextSemverTag("${VERSION}")
                    DOCKER_IMAGE_TAG =  "${DOCKER_IMAGE_TAG_PRD}" + "-" + "${env.BRANCH_NAME}"
                    FULL_IMAGE_NAME = "${DOCKER_REPO_NAME}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"
                }
                container('golang') {
                    sh './build-go-bin.sh'
                }
                kanikoBuild("index.docker.io/${FULL_IMAGE_NAME}", 'Dockerfile.run')
            }
        }
        stage('Tag repo') {
            environment {
                NEW_VERSION = "${DOCKER_IMAGE_TAG_PRD}"
            }
            steps {
                gitRemoteConfigByUrl(scmVars.GIT_URL, 'githubtoken')
                sh '''git config --global user.email "jenkins@jenkins.io"
                git config --global user.name "Jenkins"
                '''
                gitTag("v${NEW_VERSION}")
            }
        }
        // TODO: have to figure out how we can run Anchore via pod
//        stage('Anchore Validation') {
//            anchoreScan("${FULL_IMAGE_NAME}")
//        }
        stage('Update Chart') {
            when {
                not {// if with this version does NOT exist
                    expression {
                        container("helm") {
                            chartExists("${CM_ADDR}", "${CHART_NAME}", "${CHART_VERSION}", "200", "${CM_CREDS}", true)
                        }
                    }
                }
            }
            steps {
                chartCreateAndPublish("${CHART_NAME}", "${CHART_VERSION}", "${CM_ADDR}", "${CM_CREDS}")
            }
        }
        stage("Staging") {
            // TODO: does this secret file thingy work in declarative?
            environment {
                CA_PEM = credentials('letsencrypt-staging-ca')
                CM = credentials("${CM_CREDS}")
            }
            steps {
                parallel(
                    HelmInstall: {
                        container("helm") {
                            // avoid multiple sh's, this causes inter machine/process round trips
                            sh '''helm version
                            helm ls
                            helm repo add chartmuseum https://charts.kearos.net --username ${CM_USR} --password ${CM_PSW} --ca-file ${CA_PEM}
                            helm repo list
                            helm repo update
                            helm install --name cat-nip-staging chartmuseum/cat-nip --set image.tag=${DOCKER_IMAGE_TAG} --set nameOverride=cat-nip-staging
                            helm ls
                            '''
                        }
                    },
                    ZAP: {
                        container("kubectl") {
                            // wait for helm install to succeed
                            sleep 20
                            script {
                                try {
                                    sh 'kubectl run zapcli --image=owasp/zap2docker-stable --restart=Never -- zap-cli quick-scan -sc -f json --start-options \'-config api.disablekey=true\' http://cat-nip-staging.build'
                                    sleep 45
                                    sh 'kubectl logs zapcli > zap.json'
                                    archiveArtifacts 'zap.json'
                                } finally {
                                    sh 'kubectl delete pod zapcli'
                                }
                            }
                        }
                    }, Hey: {
                        // wait for helm install to succeed
                        sleep 20
                        container("hey") {
                            sh 'hey -n 1000 -c 100 http://cat-nip-staging.build > perf.txt'
                            archiveArtifacts 'perf.txt'
                        }
                    }
                )
            }
            post {
                always {
                    // TODO: can we still enter a container?
                    container("helm") {
                        // TODO: do we still have the credential available?
                        sh 'ls -lath ${CA_PEM}'
                        script {
                            withCredentials([file(credentialsId: 'letsencrypt-staging-ca', variable: 'CA_PEM')]) {
                                sh 'helm ls'
                                sh 'helm del --purge cat-nip-staging'
                            }
                        }
                    }
                }
            }
        } // end stage
        stage('Promote Image') {
            when {
                branch 'master'
            }
            environment {
                PRD = "${DOCKER_REPO_NAME}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG_PRD}"
            }
            steps {
                kanikoBuild("index.docker.io/${PRD}", 'Dockerfile.run')
            }
        }
        stage('Update PROD') {
            when {
                branch 'master'
            }
            environment {
                PR_CHANGE_NAME = "chart-${CHART_NAME}-${DOCKER_IMAGE_TAG_PRD}"
                IMAGE_TAG = "${DOCKER_IMAGE_TAG_PRD}"
                CHART = "${CHART_NAME}"
            }
            steps {
                // TODO: can we do this within environment {} ?
                script {
                    envGitInfo = git 'https://github.com/joostvdg/environments.git'
                    echo "${envGitInfo}"
                }

                sh 'git checkout -b ${PR_CHANGE_NAME}'
                container('yq') {
                    sh 'yq w -i cb/aws-eks/cat-nip/image-values.yml image.tag ${IMAGE_TAG}'
                }
                container('hub') {
                    sh 'git status'
                    gitRemoteConfigByUrl(envGitInfo.GIT_URL, 'githubtoken')
                    sh '''git config --global user.email "jenkins@jenkins.io"
                        git config --global user.name "Jenkins"
                    '''
                    sh '''git add cb/aws-eks/cat-nip/image-values.yml
                    git commit -m "update ${CHART} to image ${IMAGE_TAG}"
                    git push origin ${PR_CHANGE_NAME}
                    '''

                    // has to be indented like that, else the indents will be in the pr description
                    writeFile encoding: 'UTF-8', file: 'pr-info.md', text: """update ${CHART} to image ${IMAGE_TAG} 
\n
This pr is automatically generated via Jenkins.\\n
\n
The job: ${env.JOB_URL}
                    """

                    // TODO: unfortunately, environment {}'s credentials have fixed environment variable names
                    // TODO: in this case, they need to be EXACTLY GITHUB_PASSWORD and GITHUB_USER
                    script {
                        withCredentials([usernamePassword(credentialsId: 'github', passwordVariable: 'GITHUB_PASSWORD', usernameVariable: 'GITHUB_USER')]) {
                            sh """
                            set +x
                            hub pull-request --force -F pr-info.md -l '${CHART}' --no-edit
                            """
                        }
                    }
                }
            }
        }
    }
}
