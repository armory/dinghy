properties(
    [
        parameters([
            booleanParam(name: 'RUN_INTEGRATION', defaultValue: true, description: 'Run integration tests before pushing'),
        ]),
    ]
)

node {
    checkout scm

    stage("Build Image") {
        sh '''
          arm build
        '''
    }

    stage("Integration Test") {
        if ( params.RUN_INTEGRATION ) {
            sh '''
              arm integration
            '''
        } else {
            echo '!!!WARNING!!! Skipping integration tests'
        }
    }

    stage("Push Image") {
        sh '''
          arm push
        '''
    }
}