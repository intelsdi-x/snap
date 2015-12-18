#Contributing to snap

snap is Apache 2.0 licensed and accepts contributions via GitHub pull requests. This document
will cover how to contribute code, report issues, build the project and run the tests.

## Contributing code

Contributing code to snap is a snap (pun intended).
- Fork the project to your own repository
- Create a topic branch from where you want to base your work (usually master)
- Make commit(s) (following commit guidelines below)
- Add any needed test coverage
- Push your commit(s) to your repository
- Open a pull request against the original repo and follow the pull request guidelines below

The maintainers of the repo utilize a "Looks Good To Me" (LGTM) message in the pull request.

### Commit Guidelines

Commits should have logical groupings. A bug fix should be a single commit. A new feature
should be a single commit. 

Commit messages should be clear on what is being fixed or added to the code base. If a
commit is addressing an open issue, please start the commit message with "Fix #XXX" or 
"Feature #XXX". This will help make the generated changelog for each release easy to read
with what commits were fixes and what commits were features.

### Pull Request Guidelines

Pull requests can contain a single commit or multiple commits. If a pull request adds
a feature but also fixes two bugs, then the pull request should have three commits, one
commit each for the feature and two bug fixes.

Your pull request should be rebased against the current master branch. Please do not merge
the current master branch in with your topic branch, nor use the Update Branch button provided
by GitHub on the pull request page.

## Reporting Issues

Reporting issues are very beneficial to the project. Before reporting an issue, please review current
open issues to see if there are any matches. If there is a match, comment with a +1, or "Also seeing this issue".
If any environment details differ, please add those with your comment to the matching issue.

When reporting an issue, details are key. Include the following:
- OS version
- snap version
- Environment details (virtual, physical, etc.)
- Steps to reproduce
- Actual results
- Expected results