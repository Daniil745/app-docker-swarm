pipeline {
    agent any

    environment {
        DOCKERHUB_USER = 'daniil9090'
        APP_NAME = 'app-docker-swarm'
        TARGET_HOST = '192.168.1.104'
        TARGET_USER = 'daniil'
        VERSION = "build-${BUILD_NUMBER}"
        PREVIOUS_VERSION = ""
    }

    stages {
        stage('Checkout') {
            steps {
                echo 'Cloning repository...'
                checkout scm
            }
        }

        stage('Get Previous Version') {
            steps {
                echo 'Getting previous deployed version'
                script {
                    sh """
                        ssh -o StrictHostKeyChecking=no ${TARGET_USER}@${TARGET_HOST} "
                            docker inspect ${DOCKERHUB_USER}/${APP_NAME}:latest --format='{{.Config.Image}}' 2>/dev/null | cut -d: -f2 || echo 'none'
                        " > /tmp/prev_version.txt || echo "none" > /tmp/prev_version.txt
                    """
                    PREVIOUS_VERSION = readFile('/tmp/prev_version.txt').trim()
                    echo "Previous version: ${PREVIOUS_VERSION}"
                }
            }
        }

        stage('Build Docker Image') {
            steps {
                echo 'Building Docker image'
                sh """
                    cd app
                    docker build -t ${DOCKERHUB_USER}/${APP_NAME}:${VERSION} .
                    docker tag ${DOCKERHUB_USER}/${APP_NAME}:${VERSION} ${DOCKERHUB_USER}/${APP_NAME}:latest
                """
            }
        }

        stage('Push to Docker Hub') {
            steps {
                echo 'Pushing to Docker Hub'
                withDockerRegistry([credentialsId: 'docker-hub', url: 'https://index.docker.io/v1/' ]) {
                    sh """
                        docker push ${DOCKERHUB_USER}/${APP_NAME}:${VERSION}
                        docker push ${DOCKERHUB_USER}/${APP_NAME}:latest
                    """
                }
            }
        }

        stage('Deploy to VM2') {
            steps {
                echo 'Deploying to prod'
                sshagent(['vm2-sshkey']) {
                    sh """
                        ssh -o StrictHostKeyChecking=no ${TARGET_USER}@${TARGET_HOST} "
                            mkdir -p /opt/fact-bot &&
                            cat > /opt/fact-bot/docker-compose.yml << 'EOF'
version: '3.8'
services:
  app-docker-swarm:
    image: ${DOCKERHUB_USER}/${APP_NAME}:${VERSION}
    container_name: app-docker-swarm
    restart: unless-stopped
    ports:
      - '5000:5000'
    environment:
      - BUG_MODE=true
EOF
                            cd /opt/fact-bot &&
                            docker pull ${DOCKERHUB_USER}/${APP_NAME}:${VERSION} &&
                            docker-compose down || true &&
                            docker-compose up -d
                        "
                    """
                }
            }
        }

        stage('Healthcheck') {
            steps {
                echo 'Running healthcheck (or hc)'
                script {
                    sleep(10)

                    def maxRetries = 3
                    def healthy = false

                    for (int i = 1; i <= maxRetries; i++) {
                        echo "Healthcheck attempt ${i}/${maxRetries}"
                        def status = sh(
                            script: "curl -f -s -o /dev/null -w '%{http_code}' http://${TARGET_HOST}:5000/health",
                            returnStatus: true
                        )

                        if (status == 0) {
                            healthy = true
                            echo "Heal PASSED"
                            break
                        } else {
                            echo "Heal FAILED (attempt ${i})"
                            if (i < maxRetries) sleep(5)
                        }
                    }

                    if (!healthy) {
                        error("Healthcheck failed after ${maxRetries} attempts!")
                    }
                }
            }
        }
    }

    post {
        success {
            echo 'Pipeline completed successfully'
	    script {
	       def message = """
	       Deploy Successfully
	       Service: ${APP_NAME}
               Version: ${VERSION}
	       Time: ${new java.text.SimpleDateFormat("yyyy-MM-dd HH:mm:ss").format(new java.util.Date())}
	       URL: http://${TARGET_HOST}:5000
	       Healthcheck: PASSED
	       """
	       sendTelegramMessage(message)
	    }
        }
        failure {
            echo 'Pipeline failed'

            script {
                if (PREVIOUS_VERSION != "none" && PREVIOUS_VERSION != "") {
                    echo "Rolling back to version: ${PREVIOUS_VERSION}"
                    sshagent(['vm2-sshkey']) {
                        sh """
                            ssh -o StrictHostKeyChecking=no ${TARGET_USER}@${TARGET_HOST} "
                                cd /opt/fact-bot &&
                                cat > /opt/fact-bot/docker-compose.yml << 'EOF'
version: '3.8'
services:
  app-docker-swarm:
    image: ${DOCKERHUB_USER}/${APP_NAME}:${PREVIOUS_VERSION}
    container_name: app-docker-swarm
    restart: unless-stopped
    ports:
      - '5000:5000'
    environment:
      - BUG_MODE=false
EOF
                                docker-compose down || true &&
                                docker-compose up -d
                            "
                        """
                    }
                    echo "Rollback completed. Previous version restored."
                } else {
                    echo "No previous version found for rollback"
                }
            }
        }
    }
}

def sendTelegramMessage(String message) {
    withCredentials([
        string(credentialsId: 'tg-bot', variable: 'TELEGRAM_TOKEN'),
        string(credentialsId: 'id-user-tg', variable: 'CHAT_ID')
    ]) {
        sh """
            curl -s -X POST \
                https://api.telegram.org/bot${TELEGRAM_TOKEN}/sendMessage \
                -H 'Content-Type: application/json' \
                -d '{"chat_id": "${CHAT_ID}", "text": "${message}"}' \
                > /dev/null 2>&1 || echo "Telegram notification failed"
        """
    }
}