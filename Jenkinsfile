pipeline {
    agent any

    environment {
        DOCKER_REGISTRY = 'daniil9090'
        APP_NAME = 'film-api'
        SWARM_MASTER = 'daniil@192.168.1.104'
    }

    stages {
        stage('Checkout') {
            steps {
                echo 'Cloning repository'
                checkout scm
            }
        }

        stage('Upload Configs to Swarm Master') {
                    steps {
                        echo 'Uploading configuration files to Swarm Master'
                        sshagent(['swarm-master-ssh']) {
                            sh '''
                                ssh ${SWARM_MASTER} 'mkdir -p /home/daniil/film-api/{nginx,prometheus,grafana,static}'

                                scp nginx/nginx.conf ${SWARM_MASTER}:/home/daniil/film-api/nginx/
                                scp prometheus/prometheus.yml ${SWARM_MASTER}:/home/daniil/film-api/prometheus/
                                scp grafana/datasource.yml ${SWARM_MASTER}:/home/daniil/film-api/grafana/
                                scp internal/sql/init.sql ${SWARM_MASTER}:/home/daniil/film-api/

                                scp docker-compose.swarm.yml ${SWARM_MASTER}:/home/daniil/film-api/

                                echo "Configs uploaded successfully"
                            '''
                        }
                    }
                }

        stage('Build Docker Image') {
            steps {
                echo 'Building Docker image for Go app...'
                script {
                    def BUILD_NUM = env.BUILD_NUMBER
                    sh """
                        docker build -t ${DOCKER_REGISTRY}/${APP_NAME}:build-${BUILD_NUM} .
                        docker tag ${DOCKER_REGISTRY}/${APP_NAME}:build-${BUILD_NUM} ${DOCKER_REGISTRY}/${APP_NAME}:latest
                    """
                }
            }
        }

        stage('Push to Docker Hub') {
            steps {
                echo 'Pushing to Docker Hub'
                withDockerRegistry([credentialsId: 'docker-hub-credentials', url: '']) {
                    script {
                        def BUILD_NUM = env.BUILD_NUMBER
                        sh """
                            docker push ${DOCKER_REGISTRY}/${APP_NAME}:build-${BUILD_NUM}
                            docker push ${DOCKER_REGISTRY}/${APP_NAME}:latest
                        """
                    }
                }
            }
        }

        stage('Deploy to Swarm') {
            steps {
                echo 'Deploying to Swarm cluster...'
                sshagent(['swarm-master-ssh']) {
                    script {
                        def BUILD_NUM = env.BUILD_NUMBER
                        sh """
                            ssh -o StrictHostKeyChecking=no ${SWARM_MASTER} '
                                mkdir -p /home/daniil/film-api/{nginx,prometheus,grafana,static}

                                docker stack deploy -c /home/daniil/film-api/docker-compose.swarm.yml film-api --with-registry-auth

                                echo "=== Stack Services ==="
                                docker stack services film-api
                                echo ""
                                echo "=== Service Tasks ==="
                                docker service ls | grep film-api
                            '
                        """
                    }
                }
            }
        }

        stage('Verify Deployment') {
            steps {
                echo 'Verifying deployment...'
                sshagent(['swarm-master-ssh']) {
                    sh """
                        ssh ${SWARM_MASTER} '
                            echo "Health Checks"
                            curl -s http://localhost:8080/health || echo "API health check failed"
                            echo ""
                            echo "Service Status"
                            docker service ps film-api_web --no-trunc | head -10
                            echo ""
                            echo "Postgres Status"
                            docker service ps film-api_postgres --no-trunc | head -5
                            echo ""
                            echo "Redis Status"
                            docker service ps film-api_redis --no-trunc | head -5
                        '
                    """
                }
            }
        }
    }

    post {
        success {
            echo 'Deployment successful'
            echo 'Access points:'
            echo '  - API: http://192.168.1.104:8080'
            echo '  - Nginx: http://192.168.1.104'
            echo '  - Prometheus: http://192.168.1.104:9090'
            echo '  - Grafana: http://192.168.1.104:3000 (admin/admin)'
        }
        failure {
            echo 'Pipeline failed! Check logs above'
        }
    }
}