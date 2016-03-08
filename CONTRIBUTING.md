#Contributing to Snap

Snap is Apache 2.0 licensed and accepts contributions via GitHub pull requests. This document
will cover how to contribute code, report issues, build the project and run the tests.

## Contributing Code

Contribution of code to snap is a snap (pun intended):
- Fork the project to your own repository
- Create a topic branch from where you want to base your work (usually master)
- Make commit(s) (following commit guidelines below)
- Add any needed test coverage
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
