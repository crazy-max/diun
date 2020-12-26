# Contributing

Hi there! I'm thrilled that you'd like to contribute to this project. Your help is essential for keeping it great.

Contributions to this project are [released](https://help.github.com/articles/github-terms-of-service/#6-contributions-under-repository-license)
to the public under the [project's open source license]({{ config.repo_url }}/blob/master/LICENSE).

## Submitting a pull request

1. [Fork]({{ config.repo_url }}fork) and clone the repository
2. Configure and install the dependencies: `go mod download`
3. Create a new branch: `git checkout -b my-branch-name`
4. Make your changes
5. Validate: `docker buildx bake validate`
6. Test your code: `docker buildx bake test`
7. Build the project: `docker buildx bake artifact-all image-all`
8. Push to your fork and [submit a pull request]({{ config.repo_url }}compare)
9. Pat your self on the back and wait for your pull request to be reviewed and merged.

Here are a few things you can do that will increase the likelihood of your pull request being accepted:

* Make sure the `README.md` and any other relevant **documentation are kept up-to-date**.
* I try to follow [SemVer v2.0.0](https://semver.org/). Randomly breaking public APIs is not an option.
* Keep your change as focused as possible. If there are multiple changes you would like to make that are not dependent upon each other, consider submitting them as **separate pull requests**.
* Write a [good commit message](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).

## Resources

* [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
* [Using Pull Requests](https://docs.github.com/free-pro-team@latest/github/collaborating-with-issues-and-pull-requests/about-pull-requests)
* [GitHub Help](https://docs.github.com)
