# Contributing to Lyve Cloud Terraform Provider

First off, thanks for taking the time to contribute! ❤️

All types of contributions are encouraged and valued. See the [Table of Contents](#table-of-contents) for different ways to help and details about how this project handles them. Please make sure to read the relevant section before making your contribution. It will make it a lot easier for us maintainers and smooth out the experience for all involved. The community looks forward to your contributions. 🎉

## Table of Contents

- [Code Of Conduct](#code-of-conduct)
- [Asking Questions](#asking-questions)
- [Security Policy](#security-policy)
- [Ways to Contribute](#ways-to-contribute)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Enhancements](#suggesting-enhancements)
- [Contributing Code](#contributing-code)
- [Sign Your Commits](#sign-your-commits)

## Code Of Conduct

This project is governed by the [code of conduct](https://github.com/Seagate/.github/blob/main/CODE_OF_CONDUCT.md). You are expected to follow this as you contribute to the project.
Please report all unacceptable behavior to [opensource@seagate.com](mailto:opensource@seagate.com).

## Asking Questions

Before you ask a question, it is best to search for existing [Issues](https://github.com/Seagate/terraform-provider-lyvecloud/issues) that might help you. In case you have found a suitable issue and still need clarification, you can write your question in this issue.

If you then still feel the need to ask a question and need clarification, we recommend the following:

- Open an [Issue](https://github.com/Seagate/terraform-provider-lyvecloud/issues/new/) with the label `question`.
- Provide as much context as you can about what you're running into.

We will then take care of the issue as soon as possible.

## Security Policy

Seagate takes security seriously. If you wish to report a security flaw, or find more information about Seagate's security policy, see [SECURITY.md](SECURITY.md).

## Ways to Contribute

### Reporting Bugs

To report a bug, please create a new issue with the label `bug` and provide a one line description for the issue title.

In your report, provide as much detail as possible.

We use GitHub issues to track bugs and errors. If you run into an issue with the project:

- Open an [Issue]([/issues/new](https://github.com/Seagate/terraform-provider-lyvecloud/issues/new/)).
- Explain the behavior you would expect and the actual behavior.
- Please provide as much context as possible and describe the *reproduction steps* that someone else can follow to recreate the issue on their own.
- Provide the stack trace, input and the output.


### Suggesting Enhancements

Create a new issue with the label `enhancement` describing what it is.
If this is a feature available through other Terraform providers, please let us know the provider, version, and what options it uses so we can check into it.

### Contributing Code

Like to code and want to be more involved? Awesome!

Here's the best way to submit your code:

  1. Fork the repository
  2. Make your changes. Update version number for the provider.
  3. Make sure to [sign your commits](#sign-your-commits).
  3. Submit a pull request. If it is associated with an issue, add a comment in the issue referencing the pull request.
  4. A Seagate developer will review it. Please make any additional changes that are suggested.
  5. Pull request is accepted! Celebrate!

## Sign Your Commits

### DCO
Licensing is important to open source projects. It provides some assurances that
the software will continue to be available based under the terms that the
author(s) desired. We require that contributors sign off on commits submitted to
our project's repositories. The [Developer Certificate of Origin
(DCO)](https://developercertificate.org/) is a way to certify that you wrote and
have the right to contribute the code you are submitting to the project.

You sign-off by adding the following to your commit messages. Your sign-off must
match the git user and email associated with the commit.

    This is my commit message

    Signed-off-by: Your Name <your.name@example.com>

Git has a `-s` command line option to do this automatically:

    git commit -s -m 'This is my commit message'

If you forgot to do this and have not yet pushed your changes to the remote
repository, you can amend your commit with the sign-off by running 

    git commit --amend -s 
