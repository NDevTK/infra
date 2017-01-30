irc_commit_bot

### Building

### Deploying

Use the hash from `docker build`:

`
...
Successfully built 269892d3f65a

./push.sh 269892d3f65a
`

### System architecture

Flow of a loop is as follows (bolded items are external dependencies):

* Talk to *gitiles*, and do a `git log` to see the last ~100 commits.
* Look for new ones (last commit we've seen is stored in *Google Cloud
  Datastore*)
* Post new commits that I care about (under the announcePath in my config) to
  IRC.

Note that the bot colorizes commits if they have TBRs or NOTRY=true.
