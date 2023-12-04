import { act } from 'react-dom/test-utils';
import { Button } from '@mui/material';
import { fireEvent, screen } from '@testing-library/react';
import { renderWithAuth } from '../../auth/testUtils';
import { Platform } from '../../../api/coverage';
import * as Coverage from '../../../api/coverage';
import {
  TrendsContext,
  TrendsContextProvider,
  TrendsContextValue,
} from './TrendsContext';
import { Params } from './LoadTrends';


export interface OptionalParams {
  presets?: string[],
  paths?: string[],
  unitTestsOnly?: boolean,
  platform?: string,
  builder?: string,
  bucket?: string,
  platformList: Platform[]
}

export function createParams(params? : OptionalParams) : Params {
  return {
    unitTestsOnly: params?.unitTestsOnly || false,
    platform: params?.platform || '',
    builder: params?.builder || '',
    bucket: params?.builder || '',
    platformList: params?.platformList || [],
    presets: params?.presets || [],
    paths: params?.paths || [],
  };
}

async function contextRender(
    ui: (value: TrendsContextValue) => React.ReactElement,
    { props } = { props: createParams() },
) {
  await act(async () => {
    renderWithAuth(
        <TrendsContextProvider {...props}>
          <TrendsContext.Consumer>
            {(value) => ui(value)}
          </TrendsContext.Consumer>
        </TrendsContextProvider>,
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

    jest.spyOn(Coverage, 'getAbsoluteCoverageTrends').mockResolvedValue(
        {
          data: [
            {
              'date': '2023-06-11',
              'covered': 78,
              'total': 100,
            },
            {
              'date': '2023-06-12',
              'covered': 81,
              'total': 100,
            },
          ],
        },
    );

    jest.spyOn(Coverage, 'getIncrementalCoverageTrends').mockResolvedValue(
        {
          data: [
            {
              'date': '2023-06-11',
              'covered': 78,
              'total': 100,
            },
            {
              'date': '2023-06-12',
              'covered': 81,
              'total': 100,
            },
            {
              'date': '2023-06-13',
              'covered': 84,
              'total': 100,
            },
          ],
        },
    );
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

  it('should update paths', async () => {
    await contextRender((value) => (
      <Button data-testid='updatePaths' onClick={
        () => value.api.updatePaths(['p1', 'p2', 'p3'])
      }>{'paths-' + value.params.paths[1]}</Button>
    ));
    await act(async () => {
      fireEvent.click(screen.getByTestId('updatePaths'));
    });
    expect(screen.getByText('paths-p2')).toBeInTheDocument();
  });

  it('should update presets', async () => {
    await contextRender((value) => (
      <Button data-testid='updatePresets' onClick={
        () => value.api.updatePresets(['pr1', 'pr2', 'pr3'])
      }>{'presets-' + value.params.presets[2]}</Button>
    ));
    await act(async () => {
      fireEvent.click(screen.getByTestId('updatePresets'));
    });
    expect(screen.getByText('presets-pr3')).toBeInTheDocument();
  });
});

describe('SummaryContext fetch trends', () => {
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

    jest.spyOn(Coverage, 'getAbsoluteCoverageTrends').mockResolvedValue(
        {
          data: [
            {
              'date': '2023-06-11',
              'covered': 78,
              'total': 100,
            },
            {
              'date': '2023-06-12',
              'covered': 81,
              'total': 100,
            },
          ],
        },
    );

    jest.spyOn(Coverage, 'getIncrementalCoverageTrends').mockResolvedValue(
        {
          data: [
            {
              'date': '2023-06-11',
              'covered': 78,
              'total': 100,
            },
            {
              'date': '2023-06-12',
              'covered': 81,
              'total': 100,
            },
            {
              'date': '2023-06-13',
              'covered': 84,
              'total': 100,
            },
          ],
        },
    );
  });

  it('should fetch absolute coverage trends', async () => {
    await contextRender((value) => (
      <Button data-testid='loadAbsTrends' onClick={
        () => value.api.loadAbsTrends()
      }>{'presets-' + value.data.length}</Button>
    ));
    await act(async () => {
      fireEvent.click(screen.getByTestId('loadAbsTrends'));
    });
    expect(screen.getByText('presets-2')).toBeInTheDocument();
  });

  it('should fetch incremental coverage trends', async () => {
    await contextRender((value) => (
      <Button data-testid='loadIncTrends' onClick={
        () => value.api.loadIncTrends()
      }>{'presets-' + value.data.length}</Button>
    ));
    await act(async () => {
      fireEvent.click(screen.getByTestId('loadIncTrends'));
    });
    expect(screen.getByText('presets-3')).toBeInTheDocument();
  });
});
