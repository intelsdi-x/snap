# Large Tests: Building and Testing

This guide gets you started with writing and running large tests on Snap plugins. If your plugin needs updated to the large test framework you can find step by step instructions [here](https://github.com/kjlyon/snap/blob/master/Pluginsync_for_large_tests.md).

## Effective Large Tests

As of 1.0 we introduced a large test framework which can be added to your plugin by running our pluginsync tool. The default large test performs the following actions:
* Use environment variable to populate docker compose specification in: `scripts/test/docker_compose.yml`
* Download the latest containers via `docker pull` and run them
* Conditionally run `scripts/test/setup.rb` before any test (use this to create test database, test service etc)
* Download and run the appropriate version of Snap per `$SNAP_VERSION`
* Scan `examples/task/*.yml` for list of tasks and metrics
* Load Snap plugins, first from the local `build/linux/x86_64/*` directory, then try `build.snap-telemetry.io` s3 [bucket](http://snap.ci.snap-telemetry.io)
* Verify plugins are loaded successfully
* Attempt to create, verify, and stop every yaml task in the examples directory
* Shutdown and cleanup containers

If this is not the appropriate behavior, you can write custom large test as `{test_name}_spec.rb` in the `scripts/test` directory.

### Docker Compose
A default `docker_compose.yml` file should be supplied by the developer and placed in `./scripts/test` directory. This will be used by the default large spec test. Additional docker compose config files can be supplied for complex test scenarios and they require their own `custom_spec.rb` test.

Currently the following environment variables are passed to the Snap container:

* OS: any os available in snap-docker repo (default: alpine)
* SNAP_VERSION: any Snap version, or git sha1 that's available in the s3 bucket (default: latest)
* PLUGIN_PATH: used by large test framework, this must be included in the Snap container
Single container:
```
version: '2'
services:
   snap:                              # NOTE: do not change the snap container name
    image: intelsdi/snap:${OS}_test
    environment:
      SNAP_VERSION: "${SNAP_VERSION}"
    volumes:
      - "${PLUGIN_PATH}:/plugin"
```
Multiple container:
```
version: '2'
services:
  snap:                                 # NOTE: do not change the snap container name
    image: intelsdi/snap:alpine_test    # OS can be locked down to a specific version
    environment:
      SNAP_VERSION: "${SNAP_VERSION}"
      INFLUXDB_HOST: "${INFLUXDB_HOST}" # Custom environment variables require updates to large.sh
    volumes:
      - "${PLUGIN_PATH}:/plugin"
    links:
      - influxdb
  influxdb:
    image: influxdb:1.0
    expose:
      - "8083"
      - “8086"
```
### Travis CI:

To enable large tests on Travis CI, please enable sudo, docker, and add the appropriate test matrix settings in `.sync.yml`:
```
.travis.yml:
  sudo: true # large tests require travis.ci VMs instead of containers (enabled via sudo: true)
  services:  # this ensures docker/docker-compose is installed on the travis agent
    - docker
  env:
    global:   # If you change the matrix, please preserve environment globals:
      - ORG_PATH=/home/travis/gopath/src/github.com/intelsdi-x
      - SNAP_PLUGIN_SOURCE=/home/travis/gopath/src/github.com/${TRAVIS_REPO_SLUG}
    matrix:
      - TEST_TYPE: small             # preserve existing small tests
      - TEST_TYPE: medium            # preserve existing medium tests (make sure they exist)
      # if SNAP_VERSION:latest and OS:alpine is sufficient simply add TEST_TYPE: large
      - TEST_TYPE: large
      # if multiple SNAP_VERSION, OS needs to be tested, provide an array of versions:
      - SNAP_VERSION=latest OS=xenial TEST_TYPE=large
      - SNAP_VERSION=latest_build OS=centos7 TEST_TYPE=large
  matrix:
    # travis doesn't have an easy way to exclude large tests with a regex, so
    # please list every large test to exclude it from running on go 1.6.x
    exclude:
      - go: 1.6.x
        env: TEST_TYPE=large
      - go: 1.6.x
        env: SNAP_VERSION=latest OS=xenial TEST_TYPE=large
      - go: 1.6.x
        env: SNAP_VERSION=latest_build OS=centos7 TEST_TYPE=large
```
NOTE: If you did not set `sudo: true` and enable docker services, in travis.ci large test will fail with the following error:
```
2017-02-06 23:00:35 UTC [    error] docker needs to be installed
```

### Serverspec

The large tests are written using [serverspec](http://serverspec.org/changes.html) as the system test framework. An example installing and testing `ping`:
```
set :docker_compose_container, :snap    # required if you use the os["family"] detection functionality

context "network is functional" do
  if os["family"] == "ubuntu"
    describe package("iputils-ping") do
      it { should be_installed }
    end
  elsif os["family"] == "redhat"
    describe package("iputils") do
      it { should be_installed }
    end
  end

  describe command('ping -c1 8.8.8.8') do
    its(:exit_status) { should eq 0 }
    its(:stdout) { should contain(/1 packets received/) }
  end
end
```

If you have more than one container specified in docker compose, tests can be executed in each container separately:
```
describe docker_compose('./docker_compose.yml') do
  its_container(:snap) do
    # these tests would only run in the snap container
  end

  its_container(:influxdb) do
    # these tests would only run in the influxdb container
  end
end
```

## Running Tests
In addition to `make test-large` which is described in [BUILD_AND_TEST.md](BUILD_AND_TEST.md), you have the additional following options when using the large test framework:

Custom environment variables can be supplied such as:
```
OS=trusty SNAP_VERSION=1.0.0 make test-large
```
A subset of tasks can be selected for testing via the TASK environment variable:
```
TASK="psutil*.yml" make test-large
```
To troubleshoot a failing large test, enable the debug flag:
```
DEBUG=true make test-large
```
When the test encounters any failures in debug mode, it will be paused at a [pry session](http://pryrepl.org/). The test containers will remain running and available for further examination. When the problem has been identified, simply exit the debug session to resume testing, or use `exit-program` to quit immediately.

To spin up the environment in demo mode and pause after loading the first task:
```
DEMO=true make test-large
```
A specific task can be selected for usage in demo mode:
```
DEMO=true TASK="psutil-file.yml" make test-large
```
When you are done checking out the containers, simply type `exit-program`.

NOTE: some useful commands once the containers are running in debug or demo mode:

* Login to Snap container:  `$ docker exec -it $(docker ps | sed -n 's/\(\)\s*intelsdi\/snap.*/\1/p') /bin/bash`
* View Snap daemon log:  `$ docker logs $(docker ps | sed -n 's/\(\)\s*intelsdi\/snap.*/\1/p') `




