import { createContext, useMemo, useReducer, useState } from 'react';
import { Row } from '../../../components/table/DataTable';

export enum DirectoryNodeType {
  DIRECTORY = 'DIRECTORY',
  FILENAME = 'FILENAME',
}

export enum MetricType {
  LINE = 'LINE'
}

export interface MetricData {
  covered: number,
  total: number,
  percentageCovered: number
}

export interface Node extends Row<Node> {
  name: string,
  metrics: Map<MetricType, MetricData>,
  rows: Node[],
}

export interface Path extends Node {
  path: string,
  type: DirectoryNodeType,
  loaded: boolean,
}

export interface SummaryContextValue {
  data: Node[],
  api: Api,
  params: Params,
  isLoading: boolean,
  configLoaded: boolean;
}

export interface Platform {
  platform: string,
  bucket: string,
  builder: string,
  coverageTool: string,
  uiName: string,
  availableRevision: string,
  avaialbleModifierId: string,
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
        ref: '',
        revision: '',
        unitTestsOnly: false,
        platform: '',
        builder: '',
        bucket: '',
        platformList: [] as Platform[],
      },
      isLoading: false,
      configLoaded: false,
    },
);

type SummaryContextProviderProps = {
  platform: string,
  unitTestsOnly: boolean,
  revision: string,
  children?: React.ReactNode,
}

export interface Params {
  host: string,
  project: string,
  ref: string,
  revision: string,
  unitTestsOnly: boolean,
  platform: string,
  builder: string,
  bucket: string,
  platformList: Platform[]
}

export interface Api {
  updatePlatform: (platform: string) => void,
  updateRevision: (revision: string) => void,
  updateUnitTestsOnly: (unitTestOnly: boolean) => void
}

interface LoadingState {
  count: number,
  isLoading: boolean,
}

type LoadingAction =
  | { type: 'start' }
  | { type: 'end' }

function loadingCountReducer(state: LoadingState, action: LoadingAction): LoadingState {
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
  if (filteredPlatforms.length > 0) {
    return filteredPlatforms[0];
  }
  return null;
}

export const SummaryContextProvider = (props: SummaryContextProviderProps) => {
  const [host] = useState('');
  const [project] = useState('');
  const [ref] = useState('');
  const [revision, setRevision] = useState(props.revision);
  const [platform, setPlatform] = useState(props.platform);
  const [builder, setBuilder] = useState('');
  const [bucket, setBucket] = useState('');
  const [unitTestsOnly, setUnitTestsOnly] = useState(props.unitTestsOnly);
  const [platformList] = useState([] as Platform[]);

  const [configLoaded] = useState(false);
  const [loading] = useReducer(loadingCountReducer, { count: 0, isLoading: false });
  const [data] = useState([]);

  const params: Params = useMemo(() => ({
    host, project, ref, revision, unitTestsOnly,
    platform, builder, bucket, platformList,
  }), [
    host, project, ref, revision, unitTestsOnly,
    platform, builder, bucket, platformList,
  ]);

  const api: Api = {
    updatePlatform: (updatedPlatform: string) => {
      const filteredPlatform = platformList.filter((p) => p.platform === updatedPlatform)[0];
      setBucket(filteredPlatform.bucket);
      setBuilder(filteredPlatform.builder);
      setPlatform(filteredPlatform.platform);
      setRevision(filteredPlatform.availableRevision);
    },
    updateUnitTestsOnly: (unitTestOnly: boolean) => {
      setUnitTestsOnly(unitTestOnly);
    },
    updateRevision: (revision: string) => {
      setRevision(revision);
    },
  };

  return (
    <SummaryContext.Provider value={{ data, isLoading: loading.isLoading, api, params, configLoaded }}>
      {props.children}
    </SummaryContext.Provider>
  );
};


export default SummaryContext;
