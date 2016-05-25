# GoMit
**GoMit** provides facilities for defining, emitting, and handling events within a Go service.

1. [Contributing Code](#contributing-code)
1. [Contributing Examples](#contributing-examples)
1. [Contribute Elsewhere](#contribute-elsewhere)
1. [Thank You](#thank-you)

This repository has dedicated developers from Intel working on updates. The most helpful way to contribute is by reporting your experience through issues. Issues may not be updated while we review internally, but they're still incredibly appreciated.

## Contributing Code
**_IMPORTANT_**: We encourage contributions to the project from the community. We ask that you keep the following guidelines in mind when planning your contribution.

* Whether your contribution is for a bug fix or a feature request, **create an [Issue](https://github.com/intelsdi-x/gomit/issues)** and let us know what you are thinking
* **For bugs**, if you have already found a fix, feel free to submit a Pull Request referencing the Issue you created
* **For feature requests**, we want to improve upon the library incrementally which means small changes at a time. In order ensure your PR can be reviewed in a timely manner, please keep PRs small, e.g. <10 files and <500 lines changed. If you think this is unrealistic, then mention that within the issue and we can discuss it

Once you're ready to contribute code back to this repo, start with these steps:

* Fork the appropriate sub-projects that are affected by your change
* Clone the fork to `$GOPATH/src/github.com/intelsdi-x/`  
	```
	$ git clone https://github.com/<yourGithubID>/<project>.git
	```
* Create a topic branch for your change and checkout that branch  
    ```
    $ git checkout -b some-topic-branch
    ```
* Make your changes and run the test suite if one is provided (see below)
* Commit your changes and push them to your fork
* Open a pull request for the appropriate project
* Contributors will review your pull request, suggest changes, and merge it when it’s ready and/or offer feedback
* To report a bug or issue, please open a new issue against this repository

If you have questions feel free to contact the [maintainers](README.md#maintainers).

## Contributing Examples
The most immediately helpful way you can benefit this project is by cloning the repository, adding some further examples and submitting a pull request. 

Have you written a blog post about how you use GoMit? Send it to us!

## Contribute Elsewhere
This repository is one several Intel SDI-X projects. See other projects at https://github.com/intelsdi-x/

## Thank You
And **thank you!** Your contribution is incredibly important to us.