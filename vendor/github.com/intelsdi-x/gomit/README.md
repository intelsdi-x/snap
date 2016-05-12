<!--
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
-->

# GoMit

[![Build Status](https://travis-ci.org/intelsdi-x/gomit.svg?branch=master)](https://travis-ci.org/intelsdi-x/gomit/)

1. [Overview](#overview)
2. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Installation](#installation)
3. [Documentation](#documentation)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
4. [Contributing](#contributing)
5. [License](#license)
6. [Maintainers](#maintainers)
7. [Thank You](#thank-you)

## Overview
**GoMit** provides facilities for defining, emitting, and handling events within a Go service.

Core principles:  
* Speed over abstraction  
* No order guarantees  
* No persistence  

## Getting Started
### System Requirements
* [Golang](https://golang.org/dl/)

### Installation
Clone repo into your `$GOPATH` intelsdi-x folder:  
`git clone https://github.com/intelsdi-x/gomit.git`  
If you plan to make changes, you can fork the repository and clone that. 

## Documentation
### Examples
* [gomit_test.go](https://github.com/intelsdi-x/gomit/blob/master/gomit_test.go)

```go
type MockEventBody struct {
}

type MockThing struct {
	LastNamespace string
}

func (m *MockEventBody) Namespace() string {
	return "Mock.Event"
}
//create a function to handle the gomit event
func (m *MockThing) HandleGomitEvent(e Event) {
	m.LastNamespace = e.Namespace()
}

//create an event controller
event_controller := new(EventController)
//add registration to handler
mt := new(MockThing)
event_controller.RegisterHandler("m1", mt)
//emit event
eb := new(MockEventBody)
i, e := event_controller.Emit(eb)
//unregister handler
event_controller.UnregisterHandler("m1")
//check if handler is registered
b := event_controller.IsHandlerRegistered("m1")
```

### Roadmap
There isn't a current roadmap for this project. As we launch this project, we do not have any outstanding requirements for the next release. If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/gomit/issues/new) and/or submit a [pull request](https://github.com/intelsdi-x/gomit/pulls).

## Contributing
We love contributions! 

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License
GoMit is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Maintainers
The maintainers for GoMit are the same as [snap](http://github.com/intelsdi-x/snap). 

Amongst the many awesome contributors, there are the maintainers. These maintainers may change over time, but they are:
* Committed to reviewing pull requests, issues, and addressing comments/questions as quickly as possible
* A point of contact for questions

<table border="0" cellspacing="0" cellpadding="0">
  <tr>
    <td width="125"><a href="https://github.com/andrzej-k"><sub>@andrzej-k</sub><img src="https://avatars.githubusercontent.com/u/13486250" alt="@andrzej-k"></a></td>
    <td width="125"><a href="https://github.com/candysmurf"><sub>@candysmurf</sub><img src="https://avatars.githubusercontent.com/u/13841563" alt="@candysmurf"></a></td>
    <td width="125"><a href="https://github.com/danielscottt"><sub>@danielscottt</sub><img src="https://avatars.githubusercontent.com/u/1194436" alt="@danielscottt"></a></td>
    <td width="125"><a href="https://github.com/geauxvirtual"><sub>@geauxvirtual</sub><img src="https://avatars.githubusercontent.com/u/1395030" alt="@geauxvirtual"></a></td>
  </tr>
  <tr>
    <td width="125"><a href="https://github.com/mjbrender"><sub>@mjbrender</sub><img src="https://avatars.githubusercontent.com/u/1744971" width="100" alt="@mjbrender"></a></td>
    <td width="125"><a href="http://github.com/jcooklin"><sub>@jcooklin</sub><img src="https://avatars.githubusercontent.com/u/862968" alt="@jcooklin"></a></td>
    <td width="125"><a href="https://github.com/tiffanyfj"><sub>@tiffanyfj</sub><img src="https://avatars.githubusercontent.com/u/12282848" width="100" alt="@tiffanyfj"></a></td>
  </tr>
</table>

## Thank You
And **thank you!** Your contribution is incredibly important to us.
