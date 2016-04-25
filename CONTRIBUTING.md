#Contributing to Snap

Snap is Apache 2.0 licensed and accepts contributions via GitHub pull requests. This document
will cover how to contribute code, report issues, build the project and run the tests.

## Contributing Code

Contribution of code to snap is a snap (pun intended):
- Fork the project to your own repository
- Create a topic branch from where you want to base your work (usually master)
- Make commit(s) (following commit guidelines below)
- Add tests to cover contributed code (if necessary)
- Push your commit(s) to your repository
- Open a pull request against the original repo and follow the pull request guidelines below

The maintainers of the repo utilize a "Looks Good To Me" (LGTM) message in the pull request. After one or more maintainer states LGTM, we will merge. If you have questions or comments on your code, feel free to correct these in your branch through new commits.

### Commit Guidelines

Commits should have logical groupings. A bug fix should be a single commit. A new feature
should be a single commit.

Commit messages should be clear on what is being fixed or added to the code base. If a
commit is addressing an open issue, please start the commit message with "Fix: #XXX" or
"Feature: #XXX". This will help make the generated changelog for each release easy to read
with what commits were fixes and what commits were features.

### Testing Guidelines

For any pull request submitted, the maintainers of Snap require `small` tests that cover the code being modified and/or features being added; `medium` and `large` tests are also welcome but are not required. This breakdown of tests into `small`, `medium`, and `large` is a new taxonomy adopted by the Snap team in May, 2016. These three test types can be best described as follows:
* **Small** tests are written to exercise behavior within a single function or module. While of you might think of these as *unit* tests, a more generic term seems appropriate to avoid any confusion. In general, there is no reliance in a *small* test on any external systems (databases, web servers, etc.), and responses expected from such external services (if any) will be *mocked* or *faked*. When we say reliance on “external systems” we are including reliance on access to the network, the filesystem, external systems (eg. databases), system properties, multiple threads of execution, or the use of sleep statements as part of the test. These tests should be the easiest to automate and the fastest to run (returning a result in a minute or less, with most returning a result in a few seconds or less). These tests will be run automatically on any pull requests received from a contributor, and all *small* tests must pass before a pull request will be reviewed.
* **Medium** tests involve two or more features and test the interaction between those features. For those with previous testing experience, you might think of these as *integration* tests, but because there are a large number of other types of tests that fall into this category a more generic term is needed. The question being answered by these tests is whether or not the interactions between a feature and it’s nearest neighbors interoperate the way that they are expected to. *Medium* tests can rely on access to local services (a local database, for example), the local filesystem, multiple threads of execution, sleep statements and even access to the (local) network. However, reliance on access to external systems and services (systems and services not available on the localhost) in *medium* tests is discouraged. In general, we should expect that these tests return a result in 5 minutes or less, although some *medium* tests may return a result in much less time than that (depending on local system load). These tests can typically be automated and the set of *medium* tests will be run against any builds prior to their release.
* **Large** tests represent typical user scenarios and might be what some of you would think of as *functional* tests. However, as was the case with the previous two categories, we felt that the more generic term used by the Google team seemed to be appropriate here. For these tests, reliance on access to the network, local services (like databases), the filesystem, external systems, multiple threads of execution, system properties, and the use of sleep statements within the tests are all supported. Some of these tests might be run manually as part of the release process, but every effort is made to ensure that even these *large* tests can be automated (where possible). The response times for testing of some of these user scenarios could be 15 minutes or more (eg. it may take some time to bring the system up to an equilibrium state when load testing), so there are situations where these *large* tests will have to be triggered manually even if the test itself is run as an automated test.

This taxonomy is the same taxonomy used by the Google Test team and was described in a posting to the Google Testing Blog that can be found [here](http://googletesting.blogspot.com/2010/12/test-sizes.html).


### Pull Request Guidelines

Pull requests can contain a single commit or multiple commits. The most important part is that _**a single commit maps to a single fix**_. Here are a few scenarios:
*  If a pull request adds a feature but also fixes two bugs, then the pull request should have three commits, one commit each for the feature and two bug fixes
* If a PR is opened with 5 commits that was work involved to fix a single issue, it should be rebased to a single commit
* If a PR is opened with 5 commits, with the first three to fix one issue and the second two to fix a separate issue, then it should be rebased to two commits, one for each issue

Your pull request should be rebased against the current master branch. Please do not merge
the current master branch in with your topic branch, nor use the Update Branch button provided
by GitHub on the pull request page.

## Reporting Issues

Reporting issues are very beneficial to the project. Before reporting an issue, please review current
open issues to see if there are any matches. If there is a match, comment with a +1, or "Also seeing this issue".
If any environment details differ, please add those with your comment to the matching issue.

When reporting an issue, details are key. Include the following:
- OS version
- Snap version
- Environment details (virtual, physical, etc.)
- Steps to reproduce
- Actual results
- Expected results

## Notes on GitHub Usage
It's worth noting that we don't use all the native GitHub features for issue management. For instance, it's uncommon for us to assign issues to the developer who will address it. Here are notes on what we do use.

### Issue Labels
Snap maintainers have a set of labels we'll use to keep up with issues that are organized:

<img src="http://i.imgur.com/epDE8RO.jpg"  alt="GitHub Tagging Strategy" width="500">

* **bug** - the classic definition of missing or misbehaving code from existing functionality (this includes malfunctioning tests)
* **feature request** - any new functionality or improvements/enhancements to existing functionality. Note that we use a single term for this (instead of both feature & enhancement labels) since it's prioritized in identical ways during sprint planning
* **question** - discussions related to snap, its administration or other details that do not outline how to address the request
* **RFC** - short for [request for comment](https://en.wikipedia.org/wiki/Request_for_Comments). These are discussions of snap features requests that include detailed opinions of implementation that are up for discussion

We also add contextual notes we'll use to provide more information regarding an issue:

  * **in progress** - we're taking action (right now). It's best not to develop your own solution to an issue in this state. Comments are welcome
  * **help wanted** - A useful flag to show this issue would benefit from community support. Please comment or, if it's not in progress, say you'd like to take on the request
  * **on hold** - an idea that gained momentum but has not yet been put into a maintainer's queue to complete. Used to inform any trackers of this status
  * **tracked** - this issue is in the JIRA backlog for the team working on snap
  * **duplicate** - used to tag issues which are identical to other issues _OR_ which are resolved by the same fix of another issue (either case)
  * **wontfix** - the universal sign that we won't fix this issue. This tag is important to use as we separate out the nice-to-have features from our strategic direction
