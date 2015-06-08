# deps
sudo yum -y install epel-release
sudo yum -y install ruby gcc mysql-devel ruby-devel rubygems java-1.7.0-openjdk git hg
sudo yum -y install golang
sudo yum -y install rubygem-nokogiri
 
# riemann server
wget https://aphyr.com/riemann/riemann-0.2.9.tar.bz2
tar xvfj riemann-0.2.9.tar.bz2
sudo mv riemann-0.2.9 /usr/local/lib/riemann

# riemann dash
sudo gem install --no-ri --no-rdoc riemann-client riemann-tools riemann-dash

# systemd
sudo cp /vagrant/service/riemann.service /usr/lib/systemd/system/
sudo cp /vagrant/service/riemann-dash.service /usr/lib/systemd/system/

sudo systemctl enable riemann
sudo systemctl enable riemann-dash

sudo systemctl start riemann
sudo systemctl start riemann-dash

# gopath
echo "export GOPATH=/vagrant/go" >> /home/vagrant/.bash_profile
echo "export GOBIN=/vagrant/go/bin" >> /home/vagrant/.bash_profile
echo "export PATH=$PATH:$GOBIN" >> /home/vagrant/.bash_profile

export GOPATH=/vagrant/go
export GOBIN=/vagrant/go/bin
export PATH=$PATH:$GOBIN
go get github.com/tools/godep

cd $GOPATH/src/github.com/intelsdi-x/pulse
scripts/deps.sh
make

echo "PATH=$PATH:$GOPATH/src/github.com/intelsdi-x/pulse/build/bin" >> /home/vagrant/.bash_profile
export PATH=$PATH:$GOPATH/src/github.com/intelsdi-x/pulse/build/bin
