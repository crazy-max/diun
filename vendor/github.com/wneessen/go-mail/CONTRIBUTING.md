<!--
SPDX-FileCopyrightText: The go-mail Authors

SPDX-License-Identifier: MIT
-->

# How to contribute

**Working on your first Pull Request?** You can learn how from this *free* series [How to Contribute to an Open Source Project on GitHub](https://egghead.io/courses/how-to-contribute-to-an-open-source-project-on-github)

## Guidelines for Pull Requests

How to get your contributions merged smoothly and quickly.

* Create **small PRs** that are narrowly focused on **addressing a single concern**. We often times receive PRs that are trying to fix several things at a time, but only one fix is considered acceptable, nothing gets merged and both author's & review's time is wasted. Create more PRs to address different concerns and everyone will be happy.

* For speculative changes, consider opening an issue and discussing it first.

* Provide a good **PR description** as a record of **what** change is being made and **why** it was made. Link to a github issue if it exists.

* Unless your PR is trivial, you should expect there will be reviewer comments that you'll need to address before merging. We expect you to be reasonably responsive to those comments, otherwise the PR will be closed after 2-3 weeks of inactivity.

* Maintain **clean commit history** and use **meaningful commit messages**. PRs with messy commit history are difficult to review and won't be merged. Use `rebase -i upstream/main` to curate your commit history and/or to bring in latest changes from main (but avoid rebasing in the middle of a code review).

* Keep your PR up to date with upstream/main (if there are merge conflicts, we can't really merge your change).

* Exceptions to the rules can be made if there's a compelling reason for doing so.
