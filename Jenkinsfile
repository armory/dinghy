#!/usr/bin/env groovy

node {
  stage("Checking out code") {
      checkout scm
  }
  def buildNb = env.BUILD_NUMBER
  def props = [:]
  def armStr = sh(returnStdout: true, script: "arm dependencies service ${buildNb}").trim()
  def builds = new groovy.json.JsonSlurperClassic().parseText(armStr)
  def dinghyVersion = sh(returnStdout: true, script: "cat ./dinghy_version").trim()

  if (builds.isEmpty()) {
    currentBuild.result = 'NOT_BUILT'
    error("Nothing to build")
    return;
  }
  println "Found ${builds.size()} build(s) to perform"

  for (def buildDef : builds) {
    try {
      def commit = sh(returnStdout: true, script: "git rev-parse --short=7 HEAD").trim()
      def branch = sh(returnStdout: true, script: "git rev-parse --abbrev-ref HEAD").trim()
      def dependencies = buildDef.dependencies

      // The version of the image we're building
      def fullVersion = "${dinghyVersion}-${commit}-${buildDef.build.buildPrefix}${buildNb}"
      def prefix = "${buildDef.build.name}_dinghy"

      stage("Building ${buildDef.build.name} (${buildDef.build.buildPrefix})") {
          props << [
          "${prefix}_BUILD_NUMBER": env.BUILD_NUMBER,
          "${prefix}_BUILD_JOB": env.JOB_NAME,
          "${prefix}_BUILD_GIT_HASH": commit,
          "${prefix}_BUILD_GIT_BRANCH": branch,
          "${prefix}_SERVICE_VERSION": dinghyVersion,
          "${prefix}_SERVICE_FULL_VERSION": fullVersion
          ]
          sh """
            export DINGHY_VERSION=${fullVersion}
            arm build
          """
      }
      stage("Testing ${buildDef.build.name} (${buildDef.build.buildPrefix})") {
          sh """
            export DINGHY_VERSION=${fullVersion}
            arm integration
          """
      }

      if (buildDef.details.publish) {
        stage("Publishing ${buildDef.build.name} (${buildDef.build.buildPrefix})") {
          sh """
            export DINGHY_VERSION=${fullVersion}
            arm push
          """
          props << [
            "${prefix}_DOCKER_IMAGE_REG" : "docker.io",
            "${prefix}_DOCKER_IMAGE_ORG" : "armory",
            "${prefix}_DOCKER_IMAGE_NAME": "armory/dinghy:${fullVersion}"
          ]
        }
        stage("Assembling halconfig archive") {
          sh "(cd halconfig && tar cvf ../dinghy-halconfig.tgz .)"
          archiveArtifacts artifacts: "dinghy-halconfig.tgz"
        }
      }

      // Add children props
      buildDef.dependencies.each { k, v ->
        props << v.props
      }

      // Trigger any necessary job
      for (def job : buildDef.jobsToTrigger) {
        build job: "${job.jobName}", wait: false
      }
    } catch (caughtError) {
      slackSend color: 'danger', message: "${buildDef.build.name} of dinghy failed: ${env.JOB_NAME} - ${buildNb} (<${env.BUILD_URL}|Open>)"
      // Fail on stable
      if (buildDef.build.type == "stable") {
        throw caughtError;
      }
    }
  }
  writeFile file: 'build.properties', text: props.collect { k, v -> "${k}=${v}" }.join("\n")
  archiveArtifacts artifacts: 'build.properties'
}
