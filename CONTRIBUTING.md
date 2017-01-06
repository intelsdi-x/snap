#Contributing to Snap

Snap is Apache 2.0 licensed and accepts contributions via GitHub pull requests. This document
will cover how to contribute code, report issues, build the project and run the tests.

## Contributing Code

Contributing code to Snap is a snap (pun intended):
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
* **Medium** tests involve two or more features and test the interaction between those features. For those with previous testing experience, you might think of these as *integration* tests, but because there are a large number of other types of tests that fall into this category a more generic term is needed. The question being answered by these tests is whether or not the interactions between a feature and its nearest neighbors interoperate the way that they are expected to. *Medium* tests can rely on access to local services (a local database, for example), the local filesystem, multiple threads of execution, sleep statements and even access to the (local) network. However, reliance on access to external systems and services (systems and services not available on the localhost) in *medium* tests is discouraged. In general, we should expect that these tests return a result in 5 minutes or less, although some *medium* tests may return a result in much less time than that (depending on local system load). These tests can typically be automated and the set of *medium* tests will be run against any builds prior to their release.
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

## Reporting Security Issues

The Snap team take security very seriously. If you have any issue regarding security,
please notify us by sending an email to snap-security@intel.com and not by creating a GitHub issue.
We will follow up with you promptly with more information and a plan for remediation.
While we are not offering a security bounty, we would love to send some Snap swag your way along with our
deepest gratitude for your assistance in making Snap a more secure product.

## Notes on GitHub Usage
It's worth noting that we don't use all the native GitHub features for issue management. For instance, it's uncommon for us to assign issues to the developer who will address it. Here are notes on what we do use.

### TL;DR Labels
We use a number of labels for context in the main framework of Snap. Plugin repository labels will keep it simple. If you want to contribute to Snap, here are the most helpful ones for you:

1. **help-wanted** ([link](https://github.com/intelsdi-x/snap/labels/help-wanted)) - some specific issues maintainers would like help addressing
2. **type/rfc** ([link](https://github.com/intelsdi-x/snap/labels/type%2Frfc)) - we need active feedback on *how best* to solve these issues
3. **plugin-wishlist** ([link](https://github.com/intelsdi-x/snap/labels/plugin-wishlist)) - these are a great opportunity to write a plugin

### Issue Labels
Snap maintainers have a set of labels we use to keep up with issues. They are separated into namespaces:

* **type/** - the category of issue. All issues will have one or more
* **reviewed/** - indicator a maintainer reviewed the issue. All issues should have one or more
* **breaking-change/** - added to an Issue to note its merge would result in a change to existing behavior throughout the framework
* **component/** - issues related to a particular package in the framework
* **area/** - issues related to an overall theme and does not map to a single package
* **effort/** - amount of work to do related to resolving or merging this code change

Other indicators:
* **reviewed/on-hold** - an idea that gained momentum but has not yet been put into a maintainer's queue to complete. Used to inform any trackers of this status
* **tracked** - this issue is in the JIRA backlog for the team working on Snap
* **reviewed/duplicate** - used to tag issues which are identical to other issues _OR_ which are resolved by the same fix of another issue (either case)
* **reviewed/wont-fix** - the universal sign that we won't fix this issue. This tag is important to use as we separate out the nice-to-have features from the maintainer's agreed upon strategic direction
* **wip-do-not-merge** - was made to clarify that a PR was just beginning to be worked, specifically for a PR to indicate it is not ready for review yet

The difference between bugs, features and enhancements can be confusing. To be extra clear, we reduced it down to two options. Here are their definitions:
* **type/bug** - the classic definition of missing or misbehaving code from existing functionality (this includes malfunctioning tests)
* **type/feature-or-enhancement** - any new functionality or improvements/enhancements to existing functionality. We use one label because it's prioritized in identical ways during sprint planning

For the sake of clarity, here are a few scenarios you might see play out.

As a maintainer:
* An issue is opened stating that Snap is not working. Upon review, the maintainer finds it is an issue with a plugin. She will label the issue with **reviewed/wrong-repo** and open a new issue under the plugin where she tags the original issue reporter, links the original issue and labels it **bug** (which is available in plugins repositories).
* An issue is opened stating that Snap is not working. It turns out to be related to Snap's functionality. The maintainer will label it **type/bug**. She has time to write the fix to this issue immediately, so she labels the issue as **reviewed/in-progress**. She finds it maps to the Scheduler package and adds additional context with **component/scheduler**. As she begins to write the fix, she opens a PR that says "Fixes #" for the previous issue and labels it **wip-do-not-merge**. When she wants another maintainer to review her PR, she will remove the **wip-do-not-merge** label.
* As PR is opened that will change Snap functionality (examples at [#977](https://github.com/intelsdi-x/snap/pull/977) & [#803](https://github.com/intelsdi-x/snap/pull/803)). A maintainer labels it **reviewed/needs-2nd-review** and proceeds with the normal code review. If the initial maintainer labels LGTM, another maintainer must review it again. A discussion must take place with a technical lead before merging.
* A PR is opened which changes the metadata structure for a plugin. A maintainer labels it **reviewed/needs-2nd-review** and adds whatever **breaking-change/** labels are appropriate. If the initial maintainer labels LGTM, another maintainer must review it again. A discussion must take place with a technical lead before merging. This corresponding issue is added to a milestone that corresponds with its targeted release (real example at [#871](https://github.com/intelsdi-x/snap/issues/871)).
* A PR is opened that edits a small amount of markdown or string output text. A maintainer labels it **effort/small**, gives it a quick review to ensure it renders, writes LGTM and merges it themselves (example: [#1139](https://github.com/intelsdi-x/snap/issues/1139)).
* An issue is opened that a maintainer believes could be solved quickly and with no impact outside of its package. She labels it **effort/small** and **help-wanted** to let external contributors know they can pick this up.

And as a contributor:
* A contributor has an idea to improve Snap. He opens an issue with guidelines on how to fix it. A maintainer likes the idea, label it **type/feature-or-enhancement**. Once a maintainer or contributor begins working on the issue, it's labeled **reviewed/in-progress**. A PR is opened early in the development of the feature and labeled **wip-do-not-merge**. The label is removed once it's time for a maintainer to review the PR.
* A contributor has an idea to improve Snap. He opens an issue with guidelines on how to fix it and the maintainer labels it **type/feature-or-enhancement**. A maintainer believes the approach requires more user input and labels it **type/rfc** to indicate it's an open discussion on implementation. Once a maintainer or contributor begins working on the issue, it's labeled **reviewed/in-progress**. A PR is opened early in the development of the feature and labeled **wip-do-not-merge**. The label is removed once it's time for a maintainer to review the PR. Whoever authors the PR should check back on the original issues thread for further feedback until code is merged.
* A contributor wants to understand more about Snap and opens an issue. A maintainer sees its resolution will be an answer to the contributor, so she labels it **type/question**. The question is closed once an answer is given. If good ideas of how to improve Snap come up during that thread, she may open other issues and label them **type/** based on whether they are missing docs, improvements or bugs.

If you read through all of this, you're awesome, well-informed and ready to dive in!
