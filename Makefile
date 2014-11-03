
all: plugins	

plugins: collector-plugins pusblisher-plugins

collector-plugins: facter
	
pusblisher-plugins:

# temp while evaluating gox
facter:
	go build -v ./plugin/collector/pulse-collector-facter