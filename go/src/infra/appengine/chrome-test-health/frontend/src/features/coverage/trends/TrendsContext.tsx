/* Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
*/

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
  useState,
} from 'react';
import { AuthContext } from '../../auth/AuthContext';
import {
  CoverageTrend,
  GetProjectDefaultConfigResponse,
  Platform,
} from '../../../api/coverage';
import { ComponentContext } from '../../components/ComponentContext';
import { loadProjectDefaultConfig } from '../summary/LoadSummary';
import {
  Params,
  loadAbsoluteCoverageTrends,
  loadIncrementalCoverageTrends,
} from './LoadTrends';

export interface Api {
  updatePlatform: (platform: string) => void,
  updateUnitTestsOnly: (unitTestOnly: boolean) => void,
  updatePaths: (paths: string[]) => void,
  updatePresets: (presets: string[]) => void,
  loadAbsTrends: () => void,
  loadIncTrends: () => void,
}

export interface TrendsContextValue {
  data: CoverageTrend[],
  api: Api,
  params: Params,
  isLoading: boolean,
  isConfigLoaded: boolean;
  isAbsTrend: boolean;
}

interface TrendsContextProviderProps {
  platform: string,
  unitTestsOnly: boolean,
  paths: string[],
  presets: string[],
  isAbsTrend: boolean,
  children?: React.ReactNode,
}

interface LoadingState {
  count: number,
  isLoading: boolean,
}

type LoadingAction =
  | { type: 'start' }
  | { type: 'end' }

function loadingCountReducer(
    state: LoadingState,
    action: LoadingAction,
): LoadingState {
  const newState = { ...state };
  switch (action.type) {
    case 'start':
      newState.count++;
      break;
    case 'end':
      newState.count--;
      break;
  }
  newState.isLoading = newState.count !== 0;
  return newState;
}

export function filterPlatform(availablePlatforms: Platform[], platform: string): Platform | null {
  const filteredPlatforms = availablePlatforms.filter((p) => p.platform === platform);
  return filteredPlatforms.length > 0 ? filteredPlatforms[0] : null;
}

export const TrendsContext = createContext<TrendsContextValue>(
    {
      data: [],
      api: {
        updatePlatform: () => {/**/},
        updateUnitTestsOnly: () => {/**/},
        updatePaths: () => {/**/},
        updatePresets: () => {/**/},
        loadAbsTrends: () => {/**/},
        loadIncTrends: () => {/**/},
      },
      params: {
        unitTestsOnly: false,
        platform: '',
        builder: '',
        bucket: '',
        platformList: [] as Platform[],
        paths: [] as string[],
        presets: [] as string[],
      },
      isLoading: false,
      isConfigLoaded: false,
      isAbsTrend: false,
    },
);

export const TrendsContextProvider = (props: TrendsContextProviderProps) => {
  // ------------ Local State ------------------
  const { auth } = useContext(AuthContext);
  const { components } = useContext(ComponentContext);

  const LUCI_PROJECT = 'chromium';

  const [paths, setPaths] = useState(props.paths);
  const [presets, setPresets] = useState(props.presets);
  const [platform, setPlatform] = useState(props.platform);
  const [builder, setBuilder] = useState('');
  const [bucket, setBucket] = useState('');
  const [unitTestsOnly, setUnitTestsOnly] = useState(props.unitTestsOnly);
  const [platformList, setPlatformList] = useState([] as Platform[]);
  const [isConfigLoaded, setIsConfigLoaded] = useState(false);
  const [loading, loadingDispatch] = useReducer(loadingCountReducer, { count: 0, isLoading: false });
  const [data, setData] = useState([] as CoverageTrend[]);
  const isAbsTrend = props.isAbsTrend;

  const params: Params = useMemo(() => ({
    unitTestsOnly, platform, builder, bucket, platformList,
    presets, paths,
  }), [
    unitTestsOnly, platform, builder, bucket, platformList,
    presets, paths,
  ]);

  const api: Api = {
    updatePlatform: (updatedPlatform: string) => {
      const filteredPlatform = filterPlatform(platformList, updatedPlatform);
      if (filteredPlatform) {
        params.bucket = filteredPlatform.bucket;
        params.builder = filteredPlatform.builder;
        params.platform = filteredPlatform.platform;
        setBucket(filteredPlatform.bucket);
        setBuilder(filteredPlatform.builder);
        setPlatform(filteredPlatform.platform);
      }
    },
    updateUnitTestsOnly: (unitTestOnly: boolean) => {
      setUnitTestsOnly(unitTestOnly);
    },
    updatePaths: (paths: string[]) => {
      setPaths(paths);
    },
    updatePresets: (presets: string[]) => {
      setPresets(presets);
    },
    loadAbsTrends: () => {
      if (auth === undefined) {
        return;
      }
      loadingDispatch({ type: 'start' });
      setData([] as CoverageTrend[]);
      loadAbsoluteCoverageTrends(
          auth,
          params,
          components,
          (trends: CoverageTrend[]) => {
            setData(trends);
            loadingDispatch({ type: 'end' });
          },
          loadFailure,
      );
    },
    loadIncTrends: () => {
      if (auth === undefined) {
        return;
      }
      loadingDispatch({ type: 'start' });
      setData([] as CoverageTrend[]);
      loadIncrementalCoverageTrends(
          auth,
          params,
          (trends: CoverageTrend[]) => {
            setData(trends);
            loadingDispatch({ type: 'end' });
          },
          loadFailure,
      );
    },
  };

  // -------------- EFFECTS -------------------
  useEffect(() => {
    loadConfig(params);
  }, []);

  // ----------------- Callbacks --------------
  const loadFailure = useCallback((error: any) => {
    loadingDispatch({ type: 'end' });
    throw error;
  }, [loadingDispatch]);

  const loadConfig = useCallback((params: Params) => {
    if (auth === undefined) {
      return;
    }
    loadingDispatch({ type: 'start' });
    loadProjectDefaultConfig(
        auth,
        LUCI_PROJECT,
        (response: GetProjectDefaultConfigResponse) => {
          let platform = params.platform;
          let filteredPlatform = filterPlatform(response.builderConfig, platform);
          if (filteredPlatform == null) {
            filteredPlatform = response.builderConfig[0];
            platform = filteredPlatform.platform;
          }

          setPlatformList(response.builderConfig);
          setPlatform(platform);
          setBuilder(filteredPlatform?.builder || '');
          setBucket(filteredPlatform?.bucket || '');
          setIsConfigLoaded(true);
          loadingDispatch({ type: 'end' });
        },
        loadFailure,
    );
  }, [auth, setPlatform, setBuilder, setBucket, loadFailure]);


  return (
    <TrendsContext.Provider value={{
      data,
      isLoading: loading.isLoading,
      api,
      params,
      isConfigLoaded,
      isAbsTrend,
    }}>
      {props.children}
    </TrendsContext.Provider>
  );
};

export default TrendsContext;
