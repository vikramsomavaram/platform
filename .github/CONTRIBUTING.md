# Contributing to Tribe ✨

Tribe can power many online stores & aggregators across the internet, and your help making it even more awesome will be greatly appreciated :)

There are many ways to contribute to the project!

- [Translating strings into your language](https://github.com/tribehq/platform/wiki/Translating-Tribe).
- Answering questions on the [Tribe discord community](https://discord.gg/s5zEucn).
- Testing open [issues](https://github.com/tribehq/platform/issues) or [pull requests](https://github.com/tribehq/platform/pulls) and sharing your findings in a comment.
- Testing tribe beta versions and release candidates.
- Submitting fixes, improvements, and enhancements.
- To disclose a security issue to our team, please submit a report via email to [security@tribe.cab](mail-to://security@tribe.cab) .

If you wish to contribute code, please read the information in the sections below. Then [fork](https://help.github.com/articles/fork-a-repo/) Tribe code, commit your changes, and [submit a pull request](https://help.github.com/articles/using-pull-requests/) 🎉

We use the `help wanted` label to mark issues that are suitable for new contributors. You can find all the issues with this label [here](https://github.com/tribehq/platform/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22).

Tribe Commerce is licensed under the GPLv3+, and all contributions to the project will be released under the same license. You maintain copyright over any contribution you make, and by submitting a pull request, you are agreeing to release that contribution under the GPLv3+ license.

If you have questions about the process to contribute code or want to discuss details of your contribution, you can contact Tribe core developers on the #core channel in the [Tribe discord community](https://discord.gg/s5zEuc).

## Getting started

- [How to set up Tribe development environment](https://github.com/tribehq/platform/wiki/How-to-set-up-Tribe-development-environment)
- [Git Flow](https://github.com/tribehq/platform/wiki/Tribe-Git-Flow)
- [String localisation guidelines](https://github.com/tribehq/platform/wiki/String-localisation-guidelines)

## Coding Guidelines and Development 🛠

- Ensure you stick to the [Golang Effective Go Coding Standards](https://golang.org/doc/effective_go.html)
- Run our build process described in the document on [how to set up Tribe development environment](https://github.com/tribehq/platform/wiki/How-to-set-up-Tribe-development-environment), it will install our pre-commit hook, code sniffs, dependencies, and more.
- Whenever possible please fix pre-existing code standards errors in the files that you change. It is ok to skip that for larger files or complex fixes.
- Ensure you use LF line endings in your code editor. Use [EditorConfig](http://editorconfig.org/) if your editor supports it so that indentation, line endings and other settings are auto configured.
- When committing, reference your issue number (#1234) and include a note about the fix.
- Ensure that your code supports the minimum supported versions of Golang.
- Push the changes to your fork and submit a pull request on the master branch of the Tribe Platform repository.
- Make sure to write good and detailed commit messages (see [this post](https://chris.beams.io/posts/git-commit/) for more on this) and follow all the applicable sections of the pull request template.
- Please avoid modifying the changelog directly. These will be updated by the Tribe core team.

## Feature Requests 🚀

Feature requests can be [submitted to our issue tracker](https://github.com/tribehq/platform/issues/new?template=Feature_request.md). Be sure to include a description of the expected behavior and use case, and before submitting a request, please search for similar ones in the closed issues.

Feature request issues will remain closed until we see sufficient interest via comments and [👍 reactions](https://help.github.com/articles/about-discussions-in-issues-and-pull-requests/) from the community.

You can see a [list of current feature requests which require votes here](https://github.com/tribehq/platform/issues?q=label%3A%22votes+needed%22+label%3Aenhancement+sort%3Areactions-%2B1-desc+is%3Aclosed).