require 'spec_helper'
require 'docker'

describe "Dockerfile" do
  before(:all) do
    image = Docker::Image.build_from_dir('Users/nahall/gopath/src/github.com/intelsdi-x/snap/examples/multi-plugins/influxdb/0.12/Dockerfile')
    set :backend, :docker
    set :docker_image, image.id
  end

describe port(8083) do
  it {should be_listening }
end

it "is listening on port 8083" do
    expect(port(8083)).to be_listening
end

describe command ('SHOW DATABASES') do
        its(:stdout){should contain}
        ('playground')
end

describe command ('SHOW DIAGNOSTICS') do
        its(:stdout){should contain}
        ('0.13.0')
end

