/* Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
*/

import { fireEvent, screen } from '@testing-library/react';
import { act } from 'react-dom/test-utils';
import { Button } from '@mui/material';
import { Platform, SummaryNode } from '../../../api/coverage';
import { renderWithAuth } from '../../auth/testUtils';
import * as Coverage from '../../../api/coverage';
import { Params } from './LoadSummary';
import {
  SummaryContext,
  SummaryContextProvider,
  SummaryContextValue,
} from './SummaryContext';

export interface OptionalParams {
  host?: string,
  project?: string,
  ref?: string,
  revision?: string,
  unitTestsOnly?: boolean,
  platform?: string,
  builder?: string,
  bucket?: string,
  platformList?: Platform[]
}

export function createParams(params? : OptionalParams) : Params {
  return {
    host: params?.host || '',
    project: params?.project || '',
    gitilesRef: params?.ref || '',
    revision: params?.revision || '',
    unitTestsOnly: params?.unitTestsOnly || false,
    platform: params?.platform || '',
    builder: params?.builder || '',
    bucket: params?.builder || '',
    platformList: params?.platformList || [],
  };
}

async function contextRender(
    ui: (value: SummaryContextValue) => React.ReactElement,
    { props } = { props: createParams() },
) {
  await act(async () => {
    renderWithAuth(
        <SummaryContextProvider {...props}>
          <SummaryContext.Consumer>
            {(value) => ui(value)}
          </SummaryContext.Consumer>
        </SummaryContextProvider>,
    );
  },
  );
}

describe('SummaryContext params', () => {
  beforeEach(() => {
    jest.spyOn(Coverage, 'getProjectDefaultConfig').mockResolvedValue(
        {
          gitilesHost: 'chromium.googlesource.com',
          gitilesProject: 'chromium/src',
          gitilesRef: 'refs/heads/main',
          builderConfig: [
            {
              platform: 'linux',
              bucket: 'test-bucket',
              builder: 'test-builder',
              uiName: 'Linux',
              latestRevision: '12345',
            },
          ] as Platform[],
        },
    );

    jest.spyOn(Coverage, 'getSummaryCoverage').mockResolvedValue([
      {
        'name': 'apps/',
        'path': '//apps/',
        'summaries': [
          {
            'covered': 451,
            'name': 'line',
            'total': 1028,
          },
        ],
        'isDir': true,
        'children': [] as SummaryNode[],
      },
    ]);

    jest.spyOn(Coverage, 'getSummaryCoverage').mockResolvedValue([
      {
        'name': 'apps/',
        'path': '//apps/',
        'summaries': [
          {
            'covered': 451,
            'name': 'line',
            'total': 1028,
          },
        ],
        'isDir': true,
        'children': [] as SummaryNode[],
      },
    ]);

    jest.spyOn(Coverage, 'getTeams').mockResolvedValue({
      teams: [
        {
          'id': '1346050280',
          'name': '1346050280',
          'components': [
            'Internals>Instrumentation',
            'Internals>Instrumentation>Memory',
          ],
        },
        {
          'id': '1346052032',
          'name': '1346052032',
          'components': [
            'Infra>OmahaProxy',
          ],
        },
      ],
    });
  });

  it('should update platform', async () => {
    await contextRender((value) => (
      <Button data-testid='updatePlatform' onClick={
        () => value.api.updatePlatform('linux')
      }>{'platform-' + value.params.platform}</Button>
    ));
    await act(async () => {
      fireEvent.click(screen.getByTestId('updatePlatform'));
    });
    expect(screen.getByText('platform-linux')).toBeInTheDocument();
  });

  it('should update unitTestsOnly', async () => {
    await contextRender((value) => (
      <Button data-testid='updateUnitTestsOnly' onClick={
        () => value.api.updateUnitTestsOnly(true)
      }>{'unit-tests-only-' + value.params.unitTestsOnly}</Button>
    ));
    await act(async () => {
      fireEvent.click(screen.getByTestId('updateUnitTestsOnly'));
    });
    expect(screen.getByText('unit-tests-only-true')).toBeInTheDocument();
  });

  it('should update revision', async () => {
    await contextRender((value) => (
      <Button data-testid='updateRevision' onClick={
        () => value.api.updateRevision('2345')
      }>{'revision-' + value.params.revision}</Button>
    ));
    await act(async () => {
      fireEvent.click(screen.getByTestId('updateRevision'));
    });
    expect(screen.getByText('revision-2345')).toBeInTheDocument();
  });
});
