# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""This module provides a base class with a general feature to serialize data.

A subclass should define "public" attributes in the class instead of constructor
, and they will be serialized to a dict or deserialized from a dict. Constructor
in the subclass should accept key-value parameters only and it should pass
uninterpreted key-value ones to the base class SerializableObject.

Example usage:
1. Compose serializable objects together.

  class MyObjectA(SerializableObject):
    a = int
    _unused = 'Private class attributes are not serialized'

    def __init__(self, param='', **kwargs):
      super(MyObjectA, self).__init__(**kwargs)
      self._private_attr = 'Private attributes allowed :), but not serialized'
      self._param = param
      self.public_attr = 'Public instance attributes are not allowed :('

    @property
    def unused(self):
      return 'all properties are not serialized'

  class MyObjectB(SerializableObject):
    b = dict
    o = MyObjectA

    def func(self):
      self.o.a = 10

  obj_a = MyObjectA(param='value')
  obj_a.a = 1
  obj_b = MyObjectB()
  obj_b.b = {'key': 'value'}
  obj_b.o = obj_a

  # Alternative is to pass parameters to constructor.
  obj_b = MyObjectB(b={'key': 'value'}, o=obj_a)

  data = obj_b.ToDict()  # {'b': {'key': 'value'}, 'o': {'a': 1}}
  obj_b2 = MyObjectB.FromDict(data)  # Should equal to obj_b.

2. Use customized type validation function.

  class Future(object):
    pass

  def ValidateType(attribute_name, attribute_value):
    # input: attribute name and its value.
    # output: bool. True if valid; otherwise False.
    return isinstance(attribute_value, Future)

  obj_a = MyObjectA(type_validation_func=ValidateType, a=Future())
  assert isinstance(obj_a.a, Future), 'this should pass'
"""

import types


class SerializableObject(object):

  def __init__(self, type_validation_func=None, **kwargs):
    self._type_validation_func = type_validation_func
    self._data = {}
    for name, value in kwargs.iteritems():
      setattr(self, name, value)

  def __setattr__(self, name, value):
    if name.startswith('_'):  # Allow private instance attributes.
      object.__setattr__(self, name, value)
      return

    attribute_type = self._GetDefinedAttributes().get(name)
    assert attribute_type is not None, 'Attribute %s is undefined' % name
    if not self._type_validation_func:
      assert isinstance(value, attribute_type), 'Expected %s, but got %s' % (
          attribute_type.__name__, type(value).__name__)
    elif not isinstance(value, attribute_type):
      assert self._type_validation_func(name, value), (
          'Value for %s of type %s failed a customized type validation' % (
              name, type(value).__name__))
    self._data[name] = value

  def __getattribute__(self, name):
    # __getattr__ won't work. Because dynamically-defined attributes are
    # expected to be in the class so that they are directly accessible and
    # __getattr__ won't be triggered.
    # __getattribute__ is always triggered upon accessing of any attribute,
    # function, or method by an instance, e.g. self.__class__.
    # However, functions like __setattr__ need to access _GetDefinedAttributes
    # with an instance of the subclass.
    if name.startswith('_'):  # Avoid infinite loop to access private functions.
      return object.__getattribute__(self, name)
    if name in self._GetDefinedAttributes():
      return self._data[name]
    return object.__getattribute__(self, name)

  @classmethod
  def _GetDefinedAttributes(cls):
    if not hasattr(cls, '_dynamic_definitions'):
      d = {}
      for name in dir(cls):
        if name.startswith('_'):
          continue  # Ignore private attributes.
        value = getattr(cls, name)
        if isinstance(value, property):
          continue  # Ignore properties.
        if type(value) in (types.MethodType, types.FunctionType):
          continue  # Ignore functions and methods.
        d[name] = value
      setattr(cls, '_dynamic_definitions', d)
    return cls._dynamic_definitions

  def ToDict(self):
    """Serializes all defined public attributes and returns a dict."""
    data = {}
    defined_attributes = self._GetDefinedAttributes()
    for name, value in self._data.iteritems():
      if issubclass(defined_attributes[name], SerializableObject):
        value = value.ToDict()
      data[name] = value
    return data

  @classmethod
  def FromDict(cls, data):
    """Deserializes the given data and returns an instance of this Class."""
    assert isinstance(data, dict), ('Expecting a dict, but got %s' %
                                    type(data).__name__)
    instance = cls()
    defined_attributes = cls._GetDefinedAttributes()
    for name, value in data.iteritems():
      assert name in defined_attributes, ('%s is not defined in %s' %
                                          (name, cls.__name__))
      if issubclass(defined_attributes[name], SerializableObject):
        value = defined_attributes[name].FromDict(value)
      setattr(instance, name, value)
    return instance
