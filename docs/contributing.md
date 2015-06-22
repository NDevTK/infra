<!--
Copyright 2015 The Chromium Authors. All rights reserved.
Use of this source code is governed by a BSD-style license that can be
found in the LICENSE file.
-->

# Contributing to infra.git

## Standard workflow

Starting with an [configured checkout](source.md), here's the standard workflow
to make a modification to the source code:

* Make sure your code is up-to-date: run `gclient sync` anywhere in
  infra.git.
* Create a new branch: `git new-branch <branch-name>`.
* Make modifications, commit them using `git commit`.
* Upload your modification for review by running `git cl upload`. This
  step runs the tests, which must pass. If they don't, go back to the
  previous step to fix any issue. You can use [test.py](../test.py) to run
  tests anytime. Make sure you've added reviewers for your modifications and
  sent an email.
* Once your code has been approved, you can commit it by clicking the
  "Commit" checkbox on the code review tool, or by running
  `git cl land` if you're a Chromium committer.

## Deployment process

As of July 2015 there is no formal deployment process. To make your
changes live, you have to ask on <chrome-troopers@google.com>.

## How to add a dependency

Sometimes it is necessary to add a new Python package to the virtual
environment. See instructions in [/bootstrap/](../bootstrap/README.md).
