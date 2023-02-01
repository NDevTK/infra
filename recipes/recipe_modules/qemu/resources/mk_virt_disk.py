#!/usr/bin/python3

# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from collections import namedtuple
import argparse
import subprocess
import pathlib
import os
import re

# Make a virtual disk image. This script creates a virtual disk image with given
# partition table and partitions. It does this without a need for root/sudo
# access. The obvious way for doing this is to create a blank file and then
# partition it. This would mean using parted to create a partition table and
# partitions and then formatting the partitions to required file systems. The
# issue with this approach is that, the mkfs tool that is commonly used for
# partition doesn't accept start or end point. If you give a file to mkfs, it
# will format the entire thing. This is undesirable as it would erase our
# partition table. (Note: mformat will accept the start and end points).
#
# Workaround: We create a small blank image and run parted to write a partition
# table on it. Then for each partition we create a blank image file and run mkfs
# on it. Next we append all the partition images to the partition table image
# and then run parted again to write the partition entries in the partition
# table. This works and the resulting image is usable in a VM.

SUPPORTED_PARTITION_TABLES = ['gpt', 'msdos']
SUPPORTED_FS_TYPE = ['ext3', 'ntfs', 'fat']

# Allowed size multipliers
SIZE_MULT = {
    "KiB": 2**10,  # Kibibyte 1024 Bytes
    "MiB": 2**20,  # Mebibyte 1048576 Bytes
    "GiB": 2**30,  # Gibibyte 1073741824 Bytes
    "TiB": 2**40,  # Tebibyte 1099511627776 Bytes
    "kB": 10**3,  # Kilobyte 1000 Bytes
    "KB": 10**3,  # Kilobyte 1000 Bytes
    "MB": 10**6,  # Megabyte 1000000 Bytes
    "GB": 10**9,  # Gigabyte 1000000000 Bytes
    "TB": 10**12,  # Terabyte 1000000000000 Bytes
    "B": 1,  # 1 Byte
}

# Space to be used to write partition table at start and backup at end
PREFFERED_BLOCK_SIZE = 2 * SIZE_MULT['KiB']


class ParsePartition(argparse.Action):
  ''' ParsePartition reads the partition input from the command line
  '''

  def __init__(self, option_strings, dest, nargs=None, **kwargs):
    super(ParsePartition, self).__init__(option_strings, dest, **kwargs)

  def __call__(self, parser, namespace, values, option_string=None):
    Partition = namedtuple("Partition", "type fs size name")
    kwargs = {}
    for element in values.split(','):
      k, v = element.split('=')
      kwargs[k] = v if k != 'size' else Size.conv_to_bytes(v)
    args = getattr(namespace, self.dest)
    if not args:
      setattr(namespace, self.dest, [Partition(**kwargs)])
    else:
      args.append(Partition(**kwargs))


class Size(argparse.Action):
  ''' Size reads the human friendly memory size input and converts it to bytes
  '''

  def conv_to_bytes(size):
    ''' conv_to_bytes accepts size strings with units from SIZE_MULT and
    returns size in bytes.

    Args:
      * size: size string with any if the units from SIZE_MULT

    Returns: size in bytes
    '''
    for k, v in SIZE_MULT.items():
      if k in size:
        return v * int(size.replace(k, ''))
    return int(size)

  def __init__(self, option_strings, dest, nargs=None, **kwargs):
    super(Size, self).__init__(option_strings, dest, **kwargs)

  def __call__(self, parser, namespace, values, option_string=None):
    size = Size.conv_to_bytes(values)
    setattr(namespace, self.dest, size)


def dd(infile, outfile, bs=0, seek=0, count=0):
  ''' Run dd with given options.

  Note: A faster way to create blank files of various sizes is to use seek.
  For example: To create a 100MB file you can do
          dd if=/dev/zero of=<target> bs=1M count=100
  This will write a 100MB of zeroes to <target> which can take a while. If we
  are not worried about the contents of the file <target> being zeroed. We can
  instead seek to 99M and then write 1M which is faster.
          dd if=/dev/zero of=<target> bs=1M seek=99 count=1
  Formatting utility will clean the data wherever required and we don't have to
  worry about writing zeroes everywhere.

  Args:
    * infile: File to copy from
    * outfile: File to copy to
    * bs: Block size to use
    * seek: Number of blocks to skip in output file
    * count: Number of blocks to copy
  '''
  cmd = ['dd', 'if={}'.format(infile), 'of={}'.format(outfile)]
  if bs:
    cmd.append('bs={}'.format(bs))
  if seek:
    cmd.append('seek={}'.format(seek))
  if count:
    cmd.append('count={}'.format(count))
  _run(cmd)


def mklabel(disk_image, partition_table_type, parted):
  ''' Run mklabel with given options.

  Args:
    * disk_image: The disk image file
    * partition_table_type: Type of partition table. [msdos or gpt]
  '''
  if parted is None:
    parted = 'parted'
  if partition_table_type not in SUPPORTED_PARTITION_TABLES:
    raise Exception('{} not supported'.format(partition_table_type))
  cmd = [parted, '-s', disk_image, 'mklabel', partition_table_type]
  _run(cmd)


def mkpart(disk_image, partition_type, name, file_system_type, start, end,
           parted):
  ''' Run mkpart with given options.

  Args:
    * disk_image: The disk image file to partition
    * partition_type: The type of partition to create. One of primary,
      logical or extended
    * name: name given to the partition
    * file_system_type: The type of filesystem to record in the partition
      table
    * start: The start point for the partition
    * end: The end point for the partition
    * parted: Command to run for parted
  '''
  if parted is None:
    parted = 'parted'
  cmd = [
      parted, '-s', disk_image, 'mkpart', partition_type, name,
      file_system_type, '{}B'.format(start), '{}B'.format(end)
  ]
  _run(cmd)


def rescue(disk_image, start, end, parted):
  ''' Run parted rescue to update partitions in start end range.

  Args:
    * disk_image: The disk image file to partition
    * start: The start point for the partition
    * end: The end point for the partition
  '''
  if parted is None:
    parted = 'parted'
  # parted rescue is a great way to add the filesystems to the partition table.
  # But rescue is only interested in adding one partition a time. So we try to
  # rescue all the partitions in the image by running rescue repeatedly until
  # we don't have a new partition.
  start_sector = ''
  end_sector = end
  rescue_cmd = lambda start, end: [
      parted, '-s', '-f', disk_image, '--', 'rescue', start, end
  ]
  list_cmd = [parted, '-s', '-f', disk_image, '--', 'print', 'all']
  list_header = '^\s*Number\s*Start\s*End\s*Size.*File system.*Flags$'
  new_start = start
  while start_sector != new_start:
    start_sector = new_start
    _run(rescue_cmd(start_sector, end_sector))
    stdout, _ = _run(list_cmd)
    for idx, line in enumerate(stdout.splitlines()):
      if re.match(list_header, line):
        for _, line in enumerate(stdout.splitlines(), start=idx + 1):
          # the next lines record the end of sectors
          tokens = line.split()
          if len(tokens) > 2:
            # read the last sector
            last_sector = tokens[2]
            new_start = last_sector


def _run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE):
  ''' _run the given cmd and return stdout and stderr. Also pretty prints the
  stdout, stderr and the command run.

  Args:
    * cmd: Command to execute
    * stdout: File to write stdout to
    * stderr: File to write stderr to

  Returns stdout and stderr as strings
  '''
  print('~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~')
  print(' '.join(cmd))
  proc = subprocess.Popen(
      cmd, universal_newlines=True, stdout=stdout, stderr=stderr)
  proc.wait()
  stdout_lines = []
  stderr_lines = []
  if proc.stdout:
    print('---------------------------STDOUT---------------------------')
    while line := proc.stdout.readline():
      print(line)
      stdout_lines.append(line)
  if proc.stderr:
    print('---------------------------STDERR---------------------------')
    while line := proc.stderr.readline():
      print(line)
      stderr_lines.append(line)
  if proc.returncode:
    raise Exception('[{}]Failed to exec {}'.format(proc.returncode,
                                                   ' '.join(cmd)))
  return '\n'.join(stdout_lines), '\n'.join(stderr_lines)


def mkfs(fs_type, fs_disk_image, fs_disk_label):
  ''' Run mkfs to format the given image with the filesystem. Only supports
  SUPPORTED_FS_TYPE filesystems.

  Args:
    * fs_type: The type of filesystem to format the image to
    * fs_disk_image: The disk image file to format
    * fs_disk_label: The name for this disk
  '''
  cmd = []
  if fs_type == "fat":
    cmd = ['mkfs', '-V', '-t', fs_type, '-n', fs_disk_label, fs_disk_image]
  if "ext" in fs_type:
    cmd = [
        'mkfs', '-V', '-t', fs_type, '-v', '-c', '-F', '-F', '-E',
        'root_owner={}:{}'.format(os.getuid(), os.getgid()), '-L',
        fs_disk_label, fs_disk_image
    ]
  if fs_type == "ntfs":
    cmd = [
        'mkfs', '-V', '-t', fs_type, '-q', '-F', '-v', '-L', fs_disk_label,
        fs_disk_image
    ]
  if not cmd:
    raise Exception('Unsupported Type {}. Supported type: {}'.format(
        fs_type, SUPPORTED_FS_TYPE))
  _run(cmd)


def cat(file_out, *files):
  ''' Cat files into file_out

  Args:
    * file_out: The name of the output file
    * files: Files to concatenate in order
  '''
  cmd = ['cat', *files]
  out = open(file_out, 'w')
  _run(cmd, stdout=out)
  out.close()


def create_partition_table(partition_type, partition_size, imagefile,
                           backup_file, parted):
  ''' create_partition_table creates an image and writes a partition table
  to it

  Args:
    * partition_type: partition table type
    * partition_size: size of the partition table
    * imagefile: image filename
    * backup_file: File to write backup table to
  '''
  if parted is None:
    parted = 'parted'
  bs = PREFFERED_BLOCK_SIZE
  count = 1
  if partition_size < PREFFERED_BLOCK_SIZE:
    # Try setting to block size
    bs = partition_size
  else:
    count, r = divmod(partition_size, bs)
    if r > 0:
      # Add one to preserve alignment
      count = count + 1

  dd('/dev/zero', imagefile, bs=bs, count=1, seek=(count - 1))
  mklabel(imagefile, partition_type, parted)
  dd('/dev/zero', backup_file, bs=bs, count=1, seek=(count - 1))


def create_filesystem(filesystem_type, filesystem_size, vol_name, imagefile):
  ''' create_filesystem create an image and format it with given filesystem

  Args:
    * filesystem_type: The filesystem type to format the image to
    * filesystem_size: The size of the image
    * vol_name: Name given to this partition
    * imagefile: The image file to format
  '''
  bs = PREFFERED_BLOCK_SIZE
  count = 1
  if filesystem_size < PREFFERED_BLOCK_SIZE:
    # Try setting to block size
    bs = filesystem_size
  else:
    count, r = divmod(filesystem_size, bs)
    if r > 0:
      # Add one to preserve alignment
      count = count + 1
  dd('/dev/zero', imagefile, bs=bs, count=1, seek=(count - 1))
  mkfs(filesystem_type, imagefile, vol_name)


def main():
  desc = ''' Create a virtual hard disk image, partitioned using given scheme.
             and formatted to the given filesystem
         '''

  parser = argparse.ArgumentParser(description=desc)
  # Partition table type
  parser.add_argument(
      '-pt',
      '--partition-table',
      metavar='msdos',
      type=str,
      help='Partition table type. One of {}.'.format(
          SUPPORTED_PARTITION_TABLES),
      default='msdos')
  # Partition table size
  parser.add_argument(
      '-ps',
      '--partition-table-size',
      metavar=PREFFERED_BLOCK_SIZE,
      type=str,
      help='Size of partition table',
      action=Size,
      default=PREFFERED_BLOCK_SIZE)
  # Partitions
  parser.add_argument(
      '-p',
      '--partition',
      metavar='<partition_options>',
      type=str,
      help=''' Partition to be created in the image.
               The following options are available.
               type: The partition type entry. {primary | logical | extended}
               fs: The file system to format the partition to;
               size: size of the filesystem image to create;
               name: name of the partition to create;
           ''',
      nargs='+',
      action=ParsePartition,
      required=True)

  parser.add_argument(
      '-g',
      '--gparted',
      metavar='parted',
      type=pathlib.Path,
      help='Optional path to parted')

  parser.add_argument(
      'image',
      metavar='disk.img',
      type=pathlib.Path,
      help='Image file to create')

  args = parser.parse_args()
  print(args)

  imagefile = str(args.image)
  partition_table = args.partition_table
  partition_table_size = args.partition_table_size
  partition_table_image = '{}.part'.format(imagefile)
  partition_table_backup = '{}.blank'.format(imagefile)
  imagesize = partition_table_size
  parted = str(args.gparted) if args.gparted else args.gparted
  # create a image with partition table on it
  create_partition_table(partition_table, partition_table_size,
                         partition_table_image, partition_table_backup, parted)

  fs_images = []
  for idx, part in enumerate(args.partition):
    filesystem_size = part.size
    filesystem_type = part.fs
    partition_entry_type = part.type
    partition_name = part.name
    filesystem_image = '{}-{}-{}.fs'.format(partition_name, filesystem_type,
                                            idx)
    fsi = str(args.image.parent.joinpath(filesystem_image))
    # create a filesystem image file
    create_filesystem(filesystem_type, filesystem_size, partition_name, fsi)
    fs_images.append(fsi)

  cat(imagefile, partition_table_image, *fs_images, partition_table_backup)
  rescue(imagefile, '0', '100%', parted)
  for fs_image in fs_images:
    os.remove(fs_image)
  os.remove(partition_table_image)
  os.remove(partition_table_backup)


main()
