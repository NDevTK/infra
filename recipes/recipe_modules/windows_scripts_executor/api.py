# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from recipe_engine import recipe_api
from recipe_engine.recipe_api import Property
# Windows command helpers
from . import offline_winpe_customization as offwinpecust
from . import online_windows_customization as onwincust
from . import sources
from . import helper
from PB.recipes.infra.windows_image_builder import windows_image_builder as wib
from PB.recipes.infra.windows_image_builder import sources as src_pb


class WindowsPSExecutorAPI(recipe_api.RecipeApi):
  """API for using Windows PowerShell scripts."""

  def __init__(self, *args, **kwargs):
    super(WindowsPSExecutorAPI, self).__init__(*args, **kwargs)
    self._scripts = self.resource
    self._workdir = ''
    self._sources = None
    self._configs_dir = None
    self._customizations = []
    self._executable_cust = []
    self._config = wib.Image()

  def init(self):
    """ init initializes all the dirs and sub modules required."""
    with self.m.step.nest('Initialize the config engine'):
      self._sources = sources.Source(self.m.path['cache'].join('Pkgs'), self.m)

      self._configs_dir = self.m.path['cleanup'].join('configs')
      helper.ensure_dirs(self.m.file, [self._configs_dir])

  def init_customizations(self, config):
    """ init_customizations initializes the given config and returns
    list of customizations

    Args:
      * config: wib.Image proto config
    """
    # ensure that arch is specified in the image
    if config.arch == wib.Arch.ARCH_UNSPECIFIED:
      raise self.m.step.StepFailure('Missing arch in config')

    arch = wib.Arch.Name(config.arch).replace('ARCH_', '').lower()

    custs = []  # list of customizations from customizations.py
    # initialize all customizations
    for cust in config.customizations:
      if cust.WhichOneof('customization') == 'offline_winpe_customization':
        custs.append(
            offwinpecust.OfflineWinPECustomization(
                image=config,
                cust=cust,
                arch=arch,
                scripts=self._scripts,
                configs=self._configs_dir,
                module=self.m,
                source=self._sources))
      if cust.WhichOneof('customization') == 'online_windows_customization':
        custs.append(
            onwincust.OnlineWindowsCustomization(
                image=config,
                cust=cust,
                arch=arch,
                scripts=self._scripts,
                configs=self._configs_dir,
                module=self.m,
                source=self._sources))

    return custs

  def process_customizations(self, custs, ctx):
    """ process_customizations pins all the volatile srcs, generates
    canonnical configs, filters customizations that need to be executed.

    Args:
      * custs: List of customizations from customization.py
      * ctx: dict containing the context for the customization

    Returns list of customizations that can be executed.
    """
    with self.m.step.nest('Process the customizations'):
      self.pin_customizations(custs, ctx)
      self.gen_canonical_configs(custs)
      return custs

  def pin_customizations(self, customizations, ctx):
    """ pin_customizations pins all the sources in the customizations

    Args:
      * customizations: List of Customizations object from customizations.py
      * ctx: dict containing the context for the customization
    """
    for cust in customizations:
      with self.m.step.nest('Pin resources from {}'.format(cust.name())):
        cust.pin_sources(ctx)

  def gen_canonical_configs(self, customizations):
    """ gen_canonical_configs strips all the names in the config and returns
    individual configs containing one customization per image.

    Example:
      Given an Image

        Image{
          arch: x86,
          name: "windows10_x86_GCE",
          customizations: [
            Customization{
              OfflineWinPECustomization{
                name: "winpe_networking"
                image_dest: GCSSrc{
                  bucket: "chrome-win-wim"
                  source: "rel/win10_networking.wim"
                }
                ...
              }
            },
            Customization{
              OfflineWinPECustomization{
                name: "winpe_diskpart"
                image_src: Src{
                  gcs_src: GCSSrc{
                    bucket: "chrome-win-wim"
                    source: "rel/win10_networking.wim"
                  }
                }
                ...
              }
            }
          ]
        }

      Writes two configs: windows10_x86_GCE-winpe_networking.cfg with

        Image{
          arch: x86,
          name: "",
          customizations: [
            Customization{
              OfflineWinPECustomization{
                name: ""
                image_dest: GCSSrc{
                  bucket: "chrome-win-wim"
                  source: "rel/win10_networking.wim"
                }
                ...
              }
           }
          ]
        }

      and windows10_x86_GCE-winpe_diskpart.cfg with

        Image{
          arch: x86,
          name: "",
          customizations: [
            Customization{
              OfflineWinPECustomization{
                name: ""
                image_src: Src{
                  gcs_src: GCSSrc{
                    bucket: "chrome-win-wim"
                    source: "rel/win10_networking.wim"
                  }
                }
                ...
              }
            }
          ]
        }

      to disk, calculates the hash for each config and sets the key for each
      of them. The strings representing name of the image, customization,...
      etc,. are set to empty before calculating the hash to maintain the
      uniqueness of the hash.

    Args:
      * customizations: List of Customizations object from customizations.py
    """
    for cust in customizations:
      # create a new image object, with same arch and containing only one
      # customization
      canon_image = wib.Image(
          arch=cust.image().arch, customizations=[cust.get_canonical_cfg()])
      name = cust.name()
      # write the config to disk
      cfg_file = self._configs_dir.join('{}-{}.cfg'.format(
          cust.image().name, name))
      self.m.file.write_proto(
          'Write config {}'.format(cfg_file),
          cfg_file,
          canon_image,
          codec='BINARY')
      # estimate the unique hash for the config (identifier for the image built
      # by this config)
      key = self.m.file.file_hash(cfg_file)
      cust.set_key(key)
      # save the config to disk as <key>.cfg
      key_file = self._configs_dir.join('{}.cfg'.format(key))
      self.m.file.copy('Copy {} to {}'.format(cfg_file, key_file), cfg_file,
                       key_file)

    return customizations

  def filter_executable_customizations(self, customizations):
    """ filter_executable_customizations generates a list of customizations
    that need to be executed.

    Args:
      * customizations: List of Customizations object from customizations.py
    """
    exec_customizations = []
    for cust in customizations:
      outputs = cust.outputs
      for output in outputs:
        if output and not self._sources.exists(output):
          # add to executable list if any of the outputs, doesn't exist
          exec_customizations.append(cust)
          break
    return exec_customizations

  def download_all_packages(self, custs):
    """ download_all_packages downloads all the packages referenced by given
    custs.

    Args:
      * custs: List of Customizations object from customizations.py
    """
    for cust in custs:
      with self.m.step.nest('Download resources for {}'.format(cust.name())):
        cust.download_sources()

  def execute_customizations(self, custs):
    """ Executes the windows image builder user config.

    Args:
      * custs: List of Customizations object from customizations.py
    """
    for cust in custs:
      with self.m.step.nest('execute config {}'.format(cust.image().name)):
        cust.execute_customization()

  def get_executable_configs(self, custs):  # pragma: no cover
    """ get_executable_config takes list of customizations and returns a list
    of wib.Image proto objects that can be used to trigger other builders

    Args:
      * custs: list of customizations refs generated by process_customizations
    """
    images = []
    for cust in custs:
      image = wib.Image()
      image.CopyFrom(cust.image())
      for _ in image.customizations:
        image.customizations.pop()
      image.customizations.append(cust.customization())
      images.append(image)
    return images
