# Copyright (c) 2023 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Is fed the source of install-build-deps.py as stdin, is expected to
output JSONPB for a LinuxSystemDeps message."""

import ast
import json
import sys
from typing import cast


def extract_func(tree: ast.Module, funcname: str) -> ast.FunctionDef:
  for stmt in tree.body:
    match stmt:
      case ast.FunctionDef():
        if stmt.name == funcname:
          return stmt

  raise ValueError(f'could not find func {funcname!r} in tree')


def is_pkg_exists(stmt: ast.stmt) -> str | None:
  match stmt:
    case ast.If(
        test=ast.Call(
            func=ast.Name('package_exists'),
            args=[ast.Constant(exist_arg)],
        ),):
      return exist_arg
  return None


def extract_packages(body: list[ast.stmt]) -> list[str]:
  if len(body) > 1:
    raise ValueError(
        f'cannot handle if body with more than one statement: {ast.dump(ast.Module(body=body))}'
    )

  match body[0]:
    case ast.Expr(
        ast.Call(
            func=ast.Attribute(value=ast.Name('packages'), attr='append'),
            args=[ast.Constant(append_value)])):
      return [append_value]

    case ast.Expr(
        ast.Call(
            func=ast.Attribute(value=ast.Name('packages'), attr='extend'),
            args=[ast.List(append_values)])):
      ret = []
      for el in append_values:
        assert isinstance(el, ast.Constant)
        ret.append(el.value)
      return ret

  raise ValueError(
      f'failed to extract package list: {ast.dump(ast.Module(body=body))}')


def dev_list_preprocess(func: ast.FunctionDef) -> ast.FunctionDef:
  for i, stmt in enumerate(func.body):
    match stmt:
      case ast.If(
          test=ast.Compare(
              left=ast.Constant('ELF 64-bit'),
              ops=[ast.In()],
          ),
          body=if_body,
      ):
        func.body[i:i + 1] = if_body
        break

  return func


def arm_list_preprocess(func: ast.FunctionDef) -> ast.FunctionDef:
  for i, stmt in enumerate(func.body):
    match stmt:
      case ast.If(
          test=ast.UnaryOp(
              op=ast.Not(),
              operand=ast.Attribute(
                  value=ast.Name('options'),
                  attr='arm',
              ),
          ),):
        func.body[i:i + 1] = []
        break

  def _is_distro_codename(cmp: ast.stmt) -> tuple[str, ast.If] | None:
    match cmp:
      case ast.If(
          test=ast.Compare(
              left=ast.Call(func=ast.Name('distro_codename')),
              ops=[ast.Eq()],
              comparators=[ast.Constant(codename)],
          ),):
        return codename, cmp
    return None

  for i, stmt in enumerate(func.body):
    codename_if = _is_distro_codename(stmt)
    if codename_if:
      codename, if_stmt = codename_if
      while True:
        if codename == 'jammy':
          func.body[i:i + 1] = if_stmt.body
          return func

        if if_stmt.orelse:
          codename_if = _is_distro_codename(if_stmt.orelse[0])
          if codename_if:
            codename, if_stmt = codename_if

  return func


def nacl_list_preprocess(func: ast.FunctionDef) -> ast.FunctionDef:
  for i, stmt in enumerate(func.body):
    match stmt:
      case ast.If(
          test=ast.UnaryOp(
              op=ast.Not(),
              operand=ast.Attribute(
                  value=ast.Name('options'),
                  attr='nacl',
              ),
          ),):
        func.body[i:i + 1] = []
        break

  return func


def extract_dep_chain(func: ast.FunctionDef) -> list:
  ret = []

  for stmt in func.body:
    match stmt:
      case ast.Assign(targets=[ast.Name('packages')], value=val):
        assert isinstance(val, ast.List)
        for el in val.elts:
          assert isinstance(el, ast.Constant)
          ret.append({"packages": [el.value]})
      case ast.If():
        apt_dep = {}
        ret.append(apt_dep)

        stmt: ast.If

        while True:
          if pkg := is_pkg_exists(stmt):
            apt_dep['conditions'] = [{'package_available': pkg}]
            apt_dep['packages'] = extract_packages(stmt.body)

            match stmt.orelse:
              case []:
                # no else case
                break

              case [ast.Expr()]:
                new_dep = {}
                apt_dep['orelse'] = new_dep
                apt_dep = new_dep
                apt_dep['packages'] = extract_packages(stmt.orelse)
                break

              case [ast.If()]:
                # elif
                new_dep = {}
                apt_dep['orelse'] = new_dep
                apt_dep = new_dep
                stmt = cast(ast.If, stmt.orelse[0])
                continue

          else:
            raise ValueError(f'unparseable statement {ast.dump(stmt)}')

      case ast.AugAssign(
          target=ast.Name('packages'),
          op=ast.Add(),
          value=ast.List(pkgs),
      ):
        for el in pkgs:
          assert isinstance(el, ast.Constant)
          ret.append({"packages": [el.value]})

      case ast.Expr(
          ast.Call(
              func=ast.Attribute(ast.Name('packages'), 'extend'),
              args=[ast.List(pkgs)],
          )):
        for el in pkgs:
          assert isinstance(el, ast.Constant)
          ret.append({"packages": [el.value]})

      case ast.Expr(ast.Call(func=ast.Name('print'))):
        pass

      case ast.Return():
        pass

      case _:
        raise ValueError(f'unparseable statement {ast.dump(stmt)}')

  return ret


def main():
  script_content = sys.stdin.read()
  tree = ast.parse(script_content, "install-build-deps.py")
  assert isinstance(tree, ast.Module)

  # we need dev_list, arm_list, nacl_list and lib_list
  ret = extract_dep_chain(dev_list_preprocess(extract_func(tree, "dev_list")))
  ret += extract_dep_chain(extract_func(tree, "lib_list"))
  ret += extract_dep_chain(arm_list_preprocess(extract_func(tree, "arm_list")))
  ret += extract_dep_chain(
      nacl_list_preprocess(extract_func(tree, "nacl_list")))

  # This is the JSONPB form of a LinuxSystemDeps message.
  json.dump({'apt_dep': ret}, sys.stdout)

  return 0


if __name__ == '__main__':
  sys.exit(main())
