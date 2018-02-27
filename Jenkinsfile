node {
    checkout scm

    stage("Build Image") {
        sh '''
          arm build
        '''
    }

    stage("Push Image") {
        sh '''
          arm push
        '''
    }
}