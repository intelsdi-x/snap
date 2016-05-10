#!/bin/bash
snapctl plugin load ../build/plugin/snap-collector-mock1
snapctl plugin load ../build/plugin/snap-publisher-file
snapctl plugin load ../build/plugin/snap-processor-passthru

#snapctl task create -t ./tasks/mock-file_specific.json
snapctl task create -t ./tasks/mock-file_query.json
#snapctl task create -t ./tasks/mock-file_query.json
#snapctl plugin load ../build/plugin/snap-collector-mock2
#snapctl plugin load /home/test/Pulse_telemetry/pulse/src/github.com/intelsdi-x/snap-plugin-collector-users/build/rootfs/snap-plugin-collector-users

