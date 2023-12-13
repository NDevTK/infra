import { useContext, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { ComponentContext, updateComponentsUrl } from '../../components/ComponentContext';
import { TrendsContext } from './TrendsContext';
import { Params } from './LoadTrends';

export const UNIT_TESTS_ONLY = 'isTest';
export const PLATFORM = 'plat';
export const PATHS = 'paths';
export const PRESETS = 'presets';

function createSearchParams(components: string[], params: Params, isAbsTrend: boolean) {
  const search = new URLSearchParams();
  updateComponentsUrl(components, search);
  search.set(UNIT_TESTS_ONLY, `${params.unitTestsOnly}`);
  params.paths.forEach((path) => search.append(PATHS, path));
  if (isAbsTrend) {
    search.set(PLATFORM, params.platform);
  }

  return search;
}


function TrendsSearchParams() {
  const { params, isAbsTrend } = useContext(TrendsContext);
  const { components } = useContext(ComponentContext);
  const [, setSearchParams] = useSearchParams();

  useEffect(() => {
    setSearchParams(createSearchParams(components, params, isAbsTrend));
  }, [setSearchParams, params, components]);

  return (<></>);
}

export default TrendsSearchParams;
