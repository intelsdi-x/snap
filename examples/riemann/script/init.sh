# deps
sudo yum -y install epel-release
sudo yum -y install ruby gcc mysql-devel ruby-devel rubygems java-1.7.0-openjdk 
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
