pipeline {
  agent any

  stages {
    stage('Mock tests') {
      steps {
        sh 'go test ./internal/service -run "^TestURLService_" -count=1 -race'
        sh 'go test ./internal/handler -run "^TestHandler_" -count=1 -race'
      }
    }
  }
}
