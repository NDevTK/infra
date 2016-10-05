# Contributing to infra repositories.

[TOC]


## Standard workflow

Starting with an [configured checkout](source.md), here's the standard workflow
to make a modification to the source code:

1. Make sure your code is up-to-date: run `gclient sync` anywhere in
   infra.git.
2. Create a new branch: `git new-branch <branch-name>`.
3. Make modifications, commit them using `git commit`.
4. Upload your modification for review by running `git cl upload`. This
   step runs the tests, which must pass. If they don't, go back to the
   previous step to fix any issue. You can use [test.py](../test.py) to run
   tests anytime. Make sure you've added reviewers for your modifications and
   sent an email.
5. Once your code has been approved, you can commit it by clicking the
   "Commit" checkbox on the code review tool, or by running
   `git cl land` if you're a Chromium committer.


## See also

* [Deployment of code](deployment.md)
* [Adding a Python dependency to infra.git](../bootstrap/README.md)
