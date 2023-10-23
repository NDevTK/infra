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
  GetProjectDefaultConfigResponse,
  Platform,
  SummaryNode,
} from '../../../api/coverage';
import { ComponentContext } from '../../components/ComponentContext';
import {
  DataActionType,
  Node,
  Params,
  Path,
  dataReducer,
  loadProjectDefaultConfig,
  loadSummary,
} from './LoadSummary';

export interface Api {
  updatePlatform: (platform: string) => void,
  updateRevision: (revision: string) => void,
  updateUnitTestsOnly: (unitTestOnly: boolean) => void
}

export interface SummaryContextValue {
  data: Node[],
  api: Api,
  params: Params,
  isLoading: boolean,
  isConfigLoaded: boolean;
}

interface SummaryContextProviderProps {
  platform: string,
  unitTestsOnly: boolean,
  revision: string,
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

export function isPath(object: any): object is Path {
  return 'path' in object;
}

export function filterPlatform(availablePlatforms: Platform[], platform: string): Platform | null {
  const filteredPlatforms = availablePlatforms.filter((p) => p.platform === platform);
  return filteredPlatforms.length > 0 ? filteredPlatforms[0] : null;
}

export const SummaryContext = createContext<SummaryContextValue>(
    {
      data: [],
      api: {
        updatePlatform: () => {/**/},
        updateUnitTestsOnly: () => {/**/},
        updateRevision: () => {/**/},
      },
      params: {
        host: '',
        project: '',
        gitilesRef: '',
        revision: '',
        unitTestsOnly: false,
        platform: '',
        builder: '',
        bucket: '',
        platformList: [] as Platform[],
      },
      isLoading: false,
      isConfigLoaded: false,
    },
);

export const SummaryContextProvider = (props: SummaryContextProviderProps) => {
  // ------------ Local State ------------------
  const { auth } = useContext(AuthContext);
  const { components } = useContext(ComponentContext);

  const LUCI_PROJECT = 'chromium';

  const [host, setHost] = useState('');
  const [project, setProject] = useState('');
  const [gitilesRef, setGitilesRef] = useState('');
  const [revision, setRevision] = useState(props.revision);
  const [platform, setPlatform] = useState(props.platform);
  const [builder, setBuilder] = useState('');
  const [bucket, setBucket] = useState('');
  const [unitTestsOnly, setUnitTestsOnly] = useState(props.unitTestsOnly);
  const [platformList, setPlatformList] = useState([] as Platform[]);
  const [isConfigLoaded, setIsConfigLoaded] = useState(false);
  const [loading, loadingDispatch] = useReducer(loadingCountReducer, { count: 0, isLoading: false });
  const [data, dataDispatch] = useReducer(dataReducer, []);

  const params: Params = useMemo(() => ({
    host, project, gitilesRef, revision, unitTestsOnly,
    platform, builder, bucket, platformList,
  }), [
    host, project, gitilesRef, revision, unitTestsOnly,
    platform, builder, bucket, platformList,
  ]);

  const api: Api = {
    updatePlatform: (updatedPlatform: string) => {
      const filteredPlatform = filterPlatform(platformList, updatedPlatform);
      if (filteredPlatform) {
        params.bucket = filteredPlatform.bucket
        params.builder = filteredPlatform.builder
        params.platform = filteredPlatform.platform;
        params.revision = filteredPlatform.latestRevision;
        setBucket(filteredPlatform.bucket);
        setBuilder(filteredPlatform.builder);
        setPlatform(filteredPlatform.platform);
        setRevision(filteredPlatform.latestRevision);
      }
    },
    updateUnitTestsOnly: (unitTestOnly: boolean) => {
      setUnitTestsOnly(unitTestOnly);
    },
    updateRevision: (revision: string) => {
      setRevision(revision);
    },
  };

  // -------------- EFFECTS -------------------
  useEffect(() => {
    loadConfig(params);
  }, []);

  useEffect(() => {
    if (isConfigLoaded) {
      if (components.length == 0) {
        loadSummaryData();
      } else {
        loadSummaryDataByComponents();
      }
    }
  }, [params, components]);

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
          setHost(response.gitilesHost);
          setProject(response.gitilesProject);
          setGitilesRef(response.gitilesRef);

          let revision = params.revision;
          let platform = params.platform;
          let filteredPlatform = filterPlatform(response.builderConfig, platform);
          if (filterPlatform == null) {
            filteredPlatform = response.builderConfig[0];
            platform = filteredPlatform.platform;
            if (params.revision == '') {
              revision = filteredPlatform.latestRevision;
            }
          }

          setPlatformList(response.builderConfig);
          setPlatform(platform);
          setRevision(revision);
          setBuilder(filteredPlatform?.builder || '');
          setBucket(filteredPlatform?.bucket || '');
          setIsConfigLoaded(true);
          loadingDispatch({ type: 'end' });
        },
        loadFailure,
    );
  }, [auth, revision, params, setHost, setProject, setGitilesRef, setRevision,
    setPlatform, setBuilder, setBucket, loadFailure]);

  const loadPathNode = useCallback((node: Node) => {
    if (auth === undefined) {
      return;
    }
    if (isPath(node) && !node.loaded) {
      loadingDispatch({ type: 'start' });
      loadSummary(
          auth,
          params,
          node.path,
          [],
          (summaryNodes: SummaryNode[]) => {
            dataDispatch({
              type: DataActionType.MERGE_DIR,
              summaryNodes,
              loaded: false,
              onExpand: loadPathNode,
              parentId: node.id,
            });
            loadingDispatch({ type: 'end' });
          },
          loadFailure,
      );
    }
  }, [loadingDispatch, dataDispatch, loadFailure, auth, params]);

  const loadSummaryData = useCallback(() => {
    if (auth === undefined) {
      return;
    }
    loadingDispatch({ type: 'start' });
    dataDispatch({ type: DataActionType.CLEAR_DIR });
    loadSummary(
        auth,
        params,
        '//',
        [],
        (summaryNodes: SummaryNode[]) => {
          dataDispatch({
            type: DataActionType.MERGE_DIR,
            summaryNodes,
            loaded: false,
            onExpand: loadPathNode,
          });
          loadingDispatch({ type: 'end' });
        },
        loadFailure,
    );
  }, [auth, params, loadingDispatch, dataDispatch, loadPathNode, loadFailure]);

  const loadSummaryDataByComponents = useCallback(() => {
    if (auth === undefined) {
      return;
    }
    loadingDispatch({ type: 'start' });
    dataDispatch({ type: DataActionType.CLEAR_DIR });

    loadSummary(
        auth,
        params,
        "",
        components,
        (summaryNodes: SummaryNode[]) => {
          dataDispatch({
            type: DataActionType.BUILD_TREE,
            summaryNodes,
            onExpand: loadPathNode,
          });
          loadingDispatch({ type: 'end' });
        },
        loadFailure,
    );
  }, [auth, components, params, loadingDispatch,
    dataDispatch, loadPathNode, loadFailure]);

  return (
    <SummaryContext.Provider value={{
      data,
      isLoading: loading.isLoading,
      api,
      params,
      isConfigLoaded,
    }}>
      {props.children}
    </SummaryContext.Provider>
  );
};

export default SummaryContext;
