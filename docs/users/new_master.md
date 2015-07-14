# Setting Up a New Master

1. To set up a new master, you need three ports. File an [Infra-Labs
   ticket](https://code.google.com/p/chromium/issues/entry?labels=Type-Bug,Pri-2,Infra-Labs,Restrict-View-Google)
   requesting ports for a new master and ask which VLAN your master should
   be on (and hence which `master_base_class` you should use). The port numbers
   and `master_base_class` settings go into your `builders.pyl` config file (see below).
2. Create a new master directory in
   [build/masters](https://chromium.googlesource.com/chromium/tools/build/+/master/masters/) or
   [build_internal/masters](https://chrome-internal.googlesource.com/chrome/tools/build/+/master/masters/)
3. Create a [builders.pyl file](builders.pyl.md) in the master directory
   describing the builders on this master. Set the port numbers and
   `master_base_class` to the values you got in step 1.
4. File an [Infra-Labs
   ticket](https://code.google.com/p/chromium/issues/entry?labels=Type-Bug,Pri-2,Infra-Labs,Restrict-View-Google)
   for build slaves, and specify how many of each configuration you will need.
   After slaves are allocated, specify them in the
   [builders.pyl file](builders.pyl.md).
5. Use [buildbot-tool](https://chromium.googlesource.com/chromium/tools/build/+/master/scripts/tools/buildbot-tool)
   to generate the rest of the master configuration.
6. Add your new master to the long list in
   [masters_test.py](/chromium/tools/build/+/master/tests/masters_test.py):
   `'master.foo.bar': 'FooBar',`. You can find the class name `FooBar` in the
   `master_site_config.py` that was generated in your master directory.
7. Commit what you have, then file a third, final [Infra-Labs
   ticket](https://code.google.com/p/chromium/issues/entry?labels=Type-Bug,Pri-2,Infra-Labs)
   asking for the appropriate URLs to be set up for your master.
