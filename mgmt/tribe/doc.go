/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package tribe provides clustered management of nodes to run telemetry workflows. It allows the masterless
clustering of Pulse agents. Nodes in the same tribe bind to the same agreements, plugins and tasks. Newly added
nodes will automatically inherit tribe agreements.

For futher information refer to example video:
https://github.com/intelsdi-x/pulse/blob/master/examples/videos.md

*/
package tribe
