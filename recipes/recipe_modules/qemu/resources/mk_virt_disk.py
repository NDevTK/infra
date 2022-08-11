#!/usr/bin/python3

# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from collections import namedtuple
import argparse
import subprocess
import pathlib
import os

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
PREFFERED_BLOCK_SIZE = 4096

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


class ParsePartition(argparse.Action):
  ''' ParsePartition reads the partition input from the command line
  '''

  def __init__(self, option_strings, dest, nargs=None, **kwargs):
    super(ParsePartition, self).__init__(option_strings, dest, **kwargs)

  def __call__(self, parser, namespace, values, option_string=None):
    Partition = namedtuple("Partition", "pe_type pe_fs fs size")
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
  print(' '.join(cmd))
  proc = subprocess.Popen(cmd, universal_newlines=True)
  proc.wait()


def mklabel(disk_image, partition_table_type):
  ''' Run mklabel with given options.

  Args:
    * disk_image: The disk image file
    * partition_table_type: Type of partition table. [msdos or gpt]
  '''
  if partition_table_type not in SUPPORTED_PARTITION_TABLES:
    raise Exception('{} not supported'.format(partition_table_type))
  cmd = ['parted', '-s', disk_image, 'mklabel', partition_table_type]
  print(' '.join(cmd))
  parted = subprocess.Popen(cmd, universal_newlines=True)
  parted.wait()


def mkpart(disk_image, partition_type, file_system_type, start, end):
  ''' Run mkpart with given options.

  Args:
    * disk_image: The disk image file to partition
    * partition_type: The type of partition to create. One of primary,
      logical or extended
    * file_system_type: The type of filesystem to record in the partition
      table
    * start: The start point for the partition
    * end: The end point for the partition
  '''
  cmd = [
      'parted', '-s', disk_image, 'mkpart', partition_type, file_system_type,
      '{}B'.format(start), '{}B'.format(end)
  ]
  print(' '.join(cmd))
  parted = subprocess.Popen(cmd, universal_newlines=True)
  parted.wait()


def mkfs(fs_type, fs_disk_image):
  ''' Run mkfs to format the given image with the filesystem.

  Args:
    * fs_type: The type of filesystem to format the image to
    * fs_disk_image: The disk image file to format
  '''
  cmd = ['mkfs', '-V', '-t', fs_type, fs_disk_image]
  print(' '.join(cmd))
  proc = subprocess.Popen(cmd, universal_newlines=True)
  proc.wait()


def create_partition_table(partition_type, partition_size, imagefile):
  ''' create_partition_table creates an image and writes a partition table
  to it

  Args:
    * partition_type: partition table type
    * partition_size: size of the partition table
    * imagefile: image filename
  '''
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

  dd('/dev/zero', imagefile, bs=bs, count=count)
  mklabel(imagefile, partition_type)


def create_filesystem(filesystem_type, filesystem_size, imagefile):
  ''' create_filesystem create an image and format it with given filesystem

  Args:
    * filesystem_type: The filesystem type to format the image to
    * filesystem_size: The size of the image
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
  dd('/dev/zero', imagefile, bs=bs, count=count)
  mkfs(filesystem_type, imagefile)


def add_partition_to_image(imagefile, start, partition_table_filesystem_type,
                           partition_table_filesystem, filesystem_image,
                           filesystem_size):
  ''' add_partition_to_image adds the given filesystem image to the partition
  image and updates the partiiton image
  Args:
    * imagefile: The image file to update.
    * start: The start of filesystem in the image file
    * partition_table_filesystem_type: One of primary, logical or extended in
      partition table entry for the filesystem.
    * partition_table_filesystem: The filesystem type entry in partition table
    * filesystem_image: image file formatted with a filesystem to append to
      imagefile
    * filesystem_size: size of the image file being appended
  Returns the total size of the image
  '''
  bs = PREFFERED_BLOCK_SIZE
  seek, r = divmod(start, bs)
  if r > 0:
    # Not aligned. Update by Byte
    bs = 1
    seek = start
  # Number of blocks to copy
  count, r = divmod(filesystem_size, bs)
  if r > 0:
    # Add one to preserve alignment
    count += 1

  dd(filesystem_image, imagefile, bs=bs, seek=seek, count=count)
  end = start + (count * bs)
  mkpart(imagefile, partition_table_filesystem_type, partition_table_filesystem,
         start, end - 1)
  return end


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
      metavar='1024',
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
               pe_type: The partition type entry. One of primary, logical or extended;
               pe_fs: The partition table entry of the filesystem type;
               fs: The file system to format the partition to;
               size: size of the filesystem image to create;
           ''',
      nargs='+',
      action=ParsePartition,
      required=True)

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
  filesystem_image = '{}.fs'.format(imagefile)
  imagesize = partition_table_size

  # create a image with partition table on it
  create_partition_table(partition_table, partition_table_size, imagefile)

  for part in args.partition:
    filesystem_size = part.size
    filesystem_type = part.fs
    partition_entry_type = part.pe_type
    partition_entry_fs = part.pe_fs
    # create a filesystem image file
    create_filesystem(filesystem_type, filesystem_size, filesystem_image)
    # Add the partition to the partition table
    imagesize = add_partition_to_image(imagefile, imagesize,
                                       partition_entry_type, partition_entry_fs,
                                       filesystem_image, filesystem_size)
    os.remove(filesystem_image)


main()
