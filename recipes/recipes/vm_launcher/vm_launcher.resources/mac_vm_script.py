#!/usr/bin/env python

# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import subprocess
import os

# Path to the VM app's .swift file
home_dir = os.path.expanduser('~')


def launch_dual(binary_path: str, bundles: list, cpus: int):
    """
    Launch two virtual machines

    Parameters:
        - binary_path: The file location of the binary for launching Mac Virtual Machines.
        - bundles: A list of the virtual machine bundles.
        - cpus: The amount of cpus dedicated to each machine.
    """
    VM_COUNT = '2'    # Used to tell the VM launcher how much memory to allocate (Full for 1, half for 2)
    if len(bundles) == 2:
        vm_alpha = subprocess.Popen([binary_path, bundles[0], cpus, VM_COUNT])
        vm_beta = subprocess.Popen([binary_path, bundles[1], cpus, VM_COUNT])
    else:
        print("Two VMs require 2 Bundles")


def launch_single(binary_path: str, bundle: list, cpus: int):
    """
    Launch two virtual machines

    Parameters:
        - binary_path: The file location of the binary for launching Mac Virtual Machines.
        - bundle: The target Mac VM bundle.
        - cpus: The amount of cpus dedicated to the machine.
    """
    VM_COUNT = '1'    # Used to tell the VM launcher how much memory to allocate (Full for 1, half for 2)
    vm_alpha = subprocess.Popen([binary_path, bundle[0], cpus, VM_COUNT])


class ScriptInput():
    """
    A template for the inputs for the script
    """
    binary_path = ""
    cpu_count = 0
    bundle = []
    dual = False


def launch_virtual_machines():
    """
    Parse the input arguments and launch either 1 or 2 Mac Virtual Machines
    """

    # First argument is the binary swift file location
    # "../Purple_Panda-Swift.app/Contents/MacOS/Purple_Panda-Swift"

    arguments = ScriptInput()

    parser = argparse.ArgumentParser()
    parser.add_argument('--binary_path', nargs=1, default="../Purple_Panda-Swift.app/Contents/MacOS/Purple_Panda-Swift",
                        help='The path the Virtual Machine launching binary.')
    parser.add_argument('--cpu_count', type=ascii, nargs='?', default="4",
                        help='The number of cpus allocated to the virtual machine(s).')
    parser.add_argument('-d', '--dual', action='store_true',
                        help='Dual launch two virtual machines')
    parser.add_argument('--bundle', nargs='+',
                        help='The path to the Virtual Machine Bundles')

    # Parse the input from the command line
    parser.parse_args(namespace=arguments)

    path = arguments.binary_path[0]
    bundles = arguments.bundle
    cpus = arguments.cpu_count

    if arguments.dual:
        launch_dual(path, bundles, cpus)
    else:
        launch_single(path, bundles, cpus)


if __name__ == "__main__":
    launch_virtual_machines()
