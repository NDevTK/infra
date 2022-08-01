// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { ElementType } from 'react';

interface Props {
  component: ElementType
}

/**
 * Renders a component dynamically, useful for components defined in a List or a Map.
 *
 * @param {ReactElement} param0  The component to render.
 * @return {ReactElement} A renderable React component.
 */
export function DynamicComponentNoProps({ component }: Props) {
  const TheComponent = component;
  return <TheComponent />;
}
