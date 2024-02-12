# Contacting Infra Troopers

This page can be found at: [go/bugatrooper]

Have an issue with a piece of build infrastructure?
Our troopers are here to help. Learn more about troopering at:
[go/trooper]

Oncall hours: we cover working hours in the Pacific timezone:
+ 1600 - 0000 UTC (900 - 1700 MTV)

There is no designated oncall coverage for EMEA hours and APAC coverage is limited to some services owned by the Task Distribution team. Volunteers in
those regions may provide assistance on mailing lists for urgent issues, but
there's no guarantee.

If you are contacting a trooper to see if there is an issue with a service,
visit the [ChOps Status Dashboard] first.
If the "Current Status" of the service shows red/yellow, that means there is a known
disruption or outage, and the trooper is already aware. No need to contact us further!

The primary way to contact a trooper is via [issues.chromium.org] using
the templates and priorities established below. If you need to find the current
trooper for a specific service, check [build.chromium.org], or
[vi/chrome_infra] (internal link). If crbug.com is down
and you are unable to file a bug, please contact the team on
[infra-dev@chromium.org].

Small or non-urgent questions can also be posted in the [#ops] Chromium slack
channel or the [chops-hangout channel] (internal).

If you know your issue is with the physical hardware, or otherwise should be
handled by the Systems team, please follow their
[Rules of Engagement] (internal).

## Bug Templates

For fastest response, please use the provided templates:

*   **[go/luci-bug]**: general requests for most cases.
*   [go/buildbucket-bug]: for Buildbucket.
*   [go/swarming-bug]: for Swarming.
*   [go/luci-cv-bug]: for CV.
*   Permission/ACL requests:
    *   [Git repos]: file a bug at [go/fix-chrome-git] for read/write access to
        specific git repos.
    *   [Buildbucket BQ access]: Task Orchestration trooper will handle this.
    *   [Google Storage, CIPD, other]: chops-security team will handle these
        requests.
*   [Machine restart requests]: if a machine appears to be offline and you
    know that it's managed by the Labs team.
*   [Mobile device restart requests]: if a mobile device appears to be offline
    and you know that it's managed by the Labs team.

Also make sure to include the machine name (e.g. build11-m1)
as well as the builder name (Builder: win-archive-rel) when applicable.

## Priority Levels

Priorities are set using the `Priority=N` label. Use the following as your guideline:

*   **P0**: Immediate attention desired. The owner will stop everything they
    are doing and investigate.
    *   These reserved for massive outages, release blocking or multi-developer
        blocking productivity issues.
    *   Examples: CQ no longer committing changes.
*   **P1**: Respond within 24 hours, resolution within 1 week
    *   These are non-P0 blocking issues that need attention from a trooper
    *   Examples: disk full on device, device offline, pending time high issues.
*   **P2**: Respond within 1 week, resolution is variable, depending on the issue
    *   These are non-blocking issues or requests that need attention from a trooper
    *   Examples: Non-blocking bugs or feature improvement suggestions
*   **P3**: Non-urgent. It is ok to wait or unassign.
    *   These are non-urgent issues or nice to have changes.
    *   Examples: Large change that will require major infrastructure changes or
        something that is a moonshot.


## More Information

Common Non-Trooper Requests:

*   [Contact a Git Admin (go/git-admin-bug)]
*   [File Chrome OS infra bug (go/cros-infra-bug)]
*   [Check the Chrome OS on-call channel (go/crosoncall)] (internal)

<!-- links are sorted by order of apparition -->
[go/bugatrooper]: http://go/bugatrooper
[go/trooper]: http://go/trooper
[ChOps Status Dashboard]: https://chopsdash.appspot.com
[issues.chromium.org]: https://issues.chromium.org/issues?q=status:open
[build.chromium.org]: https://build.chromium.org
[vi/chrome_infra]: http://vi/chrome_infra
[infra-dev@chromium.org]: mailto:infra-dev@chromium.org
[#ops]: https://chromium.slack.com/messages/CGM8DQ3ST/
[chops-hangout channel]: http://go/chops-hangout
[Rules of Engagement]: http://shortn/_rMn0A4rYuy
[go/luci-bug]: http://go/luci-bug
[go/buildbucket-bug]: http://go/buildbucket-bug
[go/swarming-bug]: http://go/swarming-bug
[go/luci-cv-bug]: http://go/luci-cv-bug
[Git repos]: http://go/fix-chrome-git
[go/fix-chrome-git]: http://go/fix-chrome-git
[Buildbucket BQ access]: https://bugs.chromium.org/p/chromium/issues/entry?labels=Restrict-View-Google%2CFoundation-Troopers&components=Infra>LUCI>BuildService>Buildbucket&summary=%5BBrief%20description%20of%20problem%5D&comment=Name%20of%20service%20account%20which%20needs%20BQ%20Viewer%20permission%3A%20%0AName%20of%20BQ%20datasets%3A%20cr-buildbucket.%24your_project.builds%0A%0ANote%3A%20we%20don't%20grant%20BQ%20Job%20User%20permissions%20on%20cr-buildbucket%3B%20BQ%20queries%20should%20be%20done%20via%20your%20own%20cloud%20project.
[Google Storage, CIPD, other]: http://go/chopssec-crbug
[Machine restart requests]: http://go/chrome-labs-fixit-bug
[Mobile device restart requests]: http://go/chrome-labs-fixit-bug
[Contact a Git Admin (go/git-admin-bug)]: http://go/git-admin-bug
[File Chrome OS infra bug (go/cros-infra-bug)]: http://go/cros-infra-bug
[Check the Chrome OS on-call channel (go/crosoncall)]: http://go/crosoncall