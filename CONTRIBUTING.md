# Contributing to Snap

Contributing to Snap is a snap (pun intended :turtle:).

Snap is Apache 2.0 licensed and accepts contributions via GitHub. This document will cover how to contribute to the project and report issues. Testing is covered in [BUILD_AND_TEST.md](docs/BUILD_AND_TEST.md).

## Topics

* [Reporting Security Issues](#reporting-security-issues)
* [Reporting Issues or Feature Requests](#reporting-issues-or-feature-requests)
   * [What is an RFC?](#what-is-an-rfc)
* [Contributing Code](#contributing-code)
   * [Commit Guidelines](#commit-guidelines)
   * [Testing Guidelines](#testing-guidelines)
   * [Pull Request Guidelines](#pull-request-guidelines)
* [Submitting a Plugin](#submitting-a-plugin)
* [Notes on GitHub Usage](#notes-on-github-usage)
   * [TL;DR Labels](#tldr-labels)
   * [Issue Labels](#issue-labels)

## Reporting Security Issues

The Snap team take security very seriously. If you have any issue regarding security,
please notify us by sending an email to sys_snaptel-security@intel.com. **Do not create a GitHub issue.** We will follow up with you promptly with more information and a plan for remediation. While we are not offering a security bounty, we would love to send some Snap swag your way along with our deepest gratitude for your assistance in making Snap a more secure product.

## Reporting Issues or Feature Requests

Your contribution through issues is very beneficial to the project. We appreciate all reported issues or new requests. They come in a few different forms described below. Before opening a new issue, we appreciate you reviewing open issues to see if there are any similar requests. If there is a match, comment with a +1, "Also seeing this issue" or something like that. If any environment details differ, please add those with your comment to the matching issue.

When **reporting an issue**, details are key. Include the following:
- OS version
- Snap version
- Environment details (virtual, physical, etc.)
- Steps to reproduce
- Expected results
- Actual results

When **requesting a feature**, context is key. Some of the questions we want to answer are:
- What does this allow a user to do that they cannot do now?
- What environment is this meant for (containerized, virtualized, bare metal)?
- How urgent is the need (nice to have feature or need to have)?
- Does this align with the goals of Snap?

When **proposing a RFC**, understanding the use case is key. This type of issue includes:
- A start and end date for comments
- A set of required participants
- A proposal with a straightforward explanation as well as what is in and out of scope

Additionaly any ideas regarding design or implementation are always welcome and feel free to include them as part of an issue, feature request or RFC description.

### What is an RFC?

The Snap maintainers use RFCs, loosely based off the [EITF practice](https://www.ietf.org/rfc.html), to discuss proposed architectural and organizational improvements to the project. While most RFCs will originate from the maintainers, we welcome all users suggesting improvements through the RFC process. If you're not sure if you need to open one, you can always [discuss your idea on Slack](http://slack.snap-telemetry.io) beforehand.

## Contributing Code

To submit code:
- Create a fork of the project
- Create a topic branch from where you want to base your work (usually master)
- Make commit(s) (following [commit guidelines](#commit-guidelines) below)
- Add tests to cover contributed code
- Push your commit(s) to your topic branch on your fork
- Open a pull request against Snap master branch that follows [PR guidelines](#pull-request-guidelines)

The maintainers of Snap use a "Looks Good To Me" (LGTM) message in the pull request to note that the commits are ready to merge. After one or more maintainer states LGTM, we will merge. If you have questions or comments on your code, feel free to correct these in your branch through new commits.

### Commit Guidelines

Commits should have logical groupings. A bug fix should be a single commit. A new feature
should be a single commit.

Commit messages should be clear on what is being fixed or added to the code base. If a
commit is addressing an open issue, please start the commit message with "Fix: #XXX" or
"Feature: #XXX". This will help make the generated changelog for each release easy to read
with what commits were fixes and what commits were features.


### Testing Guidelines

For any pull request submitted, the maintainers of Snap require `small` tests that cover the code being modified and/or features being added; `medium` and `large` tests are also welcome but are not required. This breakdown of tests into `small`, `medium`, and `large` is a new taxonomy adopted by the Snap team in May, 2016 and is [described in detail here](docs/BUILD_AND_TEST.md#test-types-in-snap).


### Pull Request Guidelines

The goal of the following guidelines is to have Pull Requests (PRs) that are fairly easy to review and comprehend, and code that is easy to maintain in the future. Ideally those guidelines should be applied in any Snap related repositories (including plugins, plugin libraries, etc.).

* **One PR addresses one problem**, which ideally resolves one Issue (but can be related to part of an Issue or more than one Issue)
* **One commit per PR** (resolving only one Issue). Please squashed unneeded commits before opening or merging a PR
* **Keep it readable for human reviewers.** We prefer to see a subset of functionality (code) with tests and documentation over delivering them separately
* **Code changes start with a GitHub Issue** that is resolved by the PR (referencing related issues is helpful as well)
* **You can post comments to your PR** especially when you'd like to give reviewers some more context
* **Don't forget commenting code** to help reviewers understand and to keep [our Go Report Card](https://goreportcard.com/report/github.com/intelsdi-x/snap) at A+
* **Partial work is welcome** and you can submit it with the PR title including [WIP]
* **Please, keep us updated.** We will try our best to merge your PR, but please notice that PRs may be closed after 30 days of inactivity. This covers cases like: failing tests, unresolved conflicts against master branch or unaddressed review comments.

Your pull request should be rebased against the current master branch. Please do not merge
the current master branch in with your topic branch, nor use the Update Branch button provided
by GitHub on the pull request page.

## Submitting a Plugin

Sharing a new plugin is one of the most impressive ways to jump into the Snap community as a contributor. This process is as simple as adding your plugin information to the [list in plugis.yml](docs/plugin.yml). Read more about what is required of plugin authors in [Plugin Authoring](docs/PLUGIN_AUTHORING.md).

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

If you read through all of this, you're awesome, well-informed and ready to contribute! :squirrel:
