#!/usr/bin/env python

from setuptools import setup

setup(name='pants',
      version='1.2',
      description='Testing package (pants)',
      url='https://www.example.com/pants',
      packages=['pants'],
      depends=['shirt==3.14'],
     )
